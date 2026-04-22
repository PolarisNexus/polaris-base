// Package auth 封装 OIDC ID token 校验。
//
// 基于 Authentik 公开的 JWKS endpoint 手动实现 RS256 签名校验；
// 不引入 go-oidc / go-jose 等额外依赖。
package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Verifier 校验 Authentik 颁发的 ID token。
type Verifier struct {
	issuer   string
	audience string
	http     *http.Client
	jwksURL  string

	mu       sync.RWMutex
	keys     map[string]*rsa.PublicKey // kid -> key
	fetched  time.Time
	cacheTTL time.Duration
}

// Claims 为 platform-admin 关心的 ID token 字段子集。
type Claims struct {
	Sub               string   `json:"sub"`
	Iss               string   `json:"iss"`
	Aud               audience `json:"aud"`
	Exp               int64    `json:"exp"`
	Iat               int64    `json:"iat"`
	Email             string   `json:"email"`
	PreferredUsername string   `json:"preferred_username"`
	Groups            []string `json:"groups"`
}

// audience 兼容单 string 与数组两种 aud 形态。
type audience []string

func (a *audience) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		*a = []string{s}
		return nil
	}
	var arr []string
	if err := json.Unmarshal(b, &arr); err != nil {
		return err
	}
	*a = arr
	return nil
}

// NewVerifier 懒初始化，首次 Verify 时拉 .well-known/openid-configuration 定位 JWKS。
// issuer 保留原始形态（含尾斜杠），OIDC Core 要求与 token 的 iss 精确匹配。
func NewVerifier(issuer, audience string) *Verifier {
	return &Verifier{
		issuer:   issuer,
		audience: audience,
		http:     &http.Client{Timeout: 5 * time.Second},
		keys:     map[string]*rsa.PublicKey{},
		cacheTTL: 10 * time.Minute,
	}
}

// Verify 校验 token 签名、issuer、audience、过期时间，成功返回 claims。
func (v *Verifier) Verify(ctx context.Context, raw string) (*Claims, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, errors.New("token: malformed")
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("token: header decode: %w", err)
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("token: header parse: %w", err)
	}
	if header.Alg != "RS256" {
		return nil, fmt.Errorf("token: alg %s unsupported", header.Alg)
	}

	key, err := v.key(ctx, header.Kid)
	if err != nil {
		return nil, err
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("token: sig decode: %w", err)
	}
	signed := []byte(parts[0] + "." + parts[1])
	if err := verifyRS256(key, signed, sig); err != nil {
		return nil, fmt.Errorf("token: signature: %w", err)
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("token: payload decode: %w", err)
	}
	var c Claims
	if err := json.Unmarshal(payload, &c); err != nil {
		return nil, fmt.Errorf("token: payload parse: %w", err)
	}
	if c.Iss != v.issuer {
		return nil, fmt.Errorf("token: iss %q != %q", c.Iss, v.issuer)
	}
	if !contains(c.Aud, v.audience) {
		return nil, fmt.Errorf("token: aud %v missing %q", c.Aud, v.audience)
	}
	if time.Now().Unix() >= c.Exp {
		return nil, errors.New("token: expired")
	}
	return &c, nil
}

func (v *Verifier) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.RLock()
	k, ok := v.keys[kid]
	fresh := time.Since(v.fetched) < v.cacheTTL
	v.mu.RUnlock()
	if ok && fresh {
		return k, nil
	}
	if err := v.refresh(ctx); err != nil {
		return nil, err
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	if k, ok := v.keys[kid]; ok {
		return k, nil
	}
	return nil, fmt.Errorf("jwks: kid %q not found", kid)
}

func (v *Verifier) refresh(ctx context.Context) error {
	if v.jwksURL == "" {
		url, err := v.discoverJWKS(ctx)
		if err != nil {
			return err
		}
		v.jwksURL = url
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	resp, err := v.http.Do(req)
	if err != nil {
		return fmt.Errorf("jwks: fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jwks: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("jwks: read: %w", err)
	}
	var set struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			Alg string `json:"alg"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(body, &set); err != nil {
		return fmt.Errorf("jwks: parse: %w", err)
	}
	keys := map[string]*rsa.PublicKey{}
	for _, k := range set.Keys {
		if k.Kty != "RSA" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		e := 0
		for _, b := range eBytes {
			e = e<<8 | int(b)
		}
		keys[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}
	}
	v.mu.Lock()
	v.keys = keys
	v.fetched = time.Now()
	v.mu.Unlock()
	return nil
}

func (v *Verifier) discoverJWKS(ctx context.Context) (string, error) {
	url := strings.TrimRight(v.issuer, "/") + "/.well-known/openid-configuration"
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := v.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("oidc discovery: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("oidc discovery: status %d", resp.StatusCode)
	}
	var meta struct {
		JWKSURI string `json:"jwks_uri"`
		Issuer  string `json:"issuer"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return "", fmt.Errorf("oidc discovery: parse: %w", err)
	}
	if meta.JWKSURI == "" {
		return "", errors.New("oidc discovery: jwks_uri empty")
	}
	return meta.JWKSURI, nil
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
