// crowdsec-bouncer — APISIX forward-auth 侧车（ADR-0012 行为层 P2）。
//
// 工作方式：
//  1. 启动时 GET /v1/decisions/stream?startup=true 取 LAPI 全量快照；
//     之后每 CROWDSEC_STREAM_INTERVAL 拉一次增量（startup=false），合并到内存。
//  2. /check 按 IP 命中哪个集合决定状态：
//       ban       → 403（APISIX 拦截）
//       captcha   → 有有效 cookie 返 200；无则返 401 + 挑战页（forward-auth 回传给客户端）
//       其他      → 200
//  3. /captcha/verify 接收 Turnstile token，调 Cloudflare siteverify，通过后签 cookie。
//
// 选型注记：不走 lua-resty-crowdsec 的 APISIX 插件路径。栈统一、失败语义可切、将来
// 换决策源（WAF / 自研策略引擎）不用动 APISIX 侧。
package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Decision 复刻 LAPI /v1/decisions/stream 返回的精简字段。
type Decision struct {
	ID       int64  `json:"id"`
	Duration string `json:"duration"`
	Origin   string `json:"origin"`
	Scenario string `json:"scenario"`
	Scope    string `json:"scope"`
	Type     string `json:"type"`
	Value    string `json:"value"`
}

type streamResp struct {
	New     []Decision `json:"new"`
	Deleted []Decision `json:"deleted"`
}

// entry 是决策的单条：到期时间 + scope（Ip / Range）+ 原因。
type entry struct {
	expireAt time.Time
	isRange  bool
	cidr     *net.IPNet
	ip       net.IP
	scenario string
}

type bouncer struct {
	lapiURL  string
	apiKey   string
	interval time.Duration
	client   *http.Client

	turnstileSiteKey   string
	turnstileSecretKey string
	cookieSecret       []byte
	cookieTTL          time.Duration
	cookieName         string

	mu       sync.RWMutex
	bans     map[int64]*entry
	captchas map[int64]*entry
	ready    atomic.Bool
}

const defaultCookieName = "polaris_captcha_pass"

func newBouncer(lapiURL, apiKey string, interval time.Duration) *bouncer {
	return &bouncer{
		lapiURL:    strings.TrimRight(lapiURL, "/"),
		apiKey:     apiKey,
		interval:   interval,
		client:     &http.Client{Timeout: 10 * time.Second},
		bans:       map[int64]*entry{},
		captchas:   map[int64]*entry{},
		cookieName: defaultCookieName,
	}
}

// pull 调一次 LAPI stream。startup=true 代表请求全量快照。
func (b *bouncer) pull(ctx context.Context, startup bool) error {
	u := fmt.Sprintf("%s/v1/decisions/stream?startup=%t", b.lapiURL, startup)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Api-Key", b.apiKey)
	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("lapi stream: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("lapi stream: status %d", resp.StatusCode)
	}
	var body streamResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("lapi stream: decode: %w", err)
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if startup {
		b.bans = make(map[int64]*entry, len(body.New))
		b.captchas = map[int64]*entry{}
	}
	addedBan, addedCap := 0, 0
	for _, d := range body.New {
		e := parseDecision(d)
		if e == nil {
			continue
		}
		switch strings.ToLower(d.Type) {
		case "ban":
			b.bans[d.ID] = e
			addedBan++
		case "captcha":
			b.captchas[d.ID] = e
			addedCap++
		}
	}
	removed := 0
	for _, d := range body.Deleted {
		if _, ok := b.bans[d.ID]; ok {
			delete(b.bans, d.ID)
			removed++
		}
		if _, ok := b.captchas[d.ID]; ok {
			delete(b.captchas, d.ID)
			removed++
		}
	}
	log.Printf("lapi stream (startup=%t): +ban=%d +captcha=%d -%d | bans=%d captchas=%d",
		startup, addedBan, addedCap, removed, len(b.bans), len(b.captchas))
	return nil
}

func parseDecision(d Decision) *entry {
	scope := strings.ToLower(d.Scope)
	dur, err := time.ParseDuration(d.Duration)
	exp := time.Now().Add(24 * time.Hour)
	if err == nil {
		exp = time.Now().Add(dur)
	}
	e := &entry{expireAt: exp, scenario: d.Scenario}
	switch scope {
	case "ip":
		ip := net.ParseIP(d.Value)
		if ip == nil {
			return nil
		}
		e.ip = ip
		return e
	case "range":
		_, n, err := net.ParseCIDR(d.Value)
		if err != nil {
			return nil
		}
		e.isRange = true
		e.cidr = n
		return e
	default:
		return nil
	}
}

func (b *bouncer) matchIn(m map[int64]*entry, ip net.IP) (bool, string) {
	now := time.Now()
	for _, e := range m {
		if now.After(e.expireAt) {
			continue
		}
		if e.isRange {
			if e.cidr.Contains(ip) {
				return true, e.scenario
			}
			continue
		}
		if e.ip.Equal(ip) {
			return true, e.scenario
		}
	}
	return false, ""
}

func (b *bouncer) match(ip net.IP) (kind string, scenario string) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if hit, s := b.matchIn(b.bans, ip); hit {
		return "ban", s
	}
	if hit, s := b.matchIn(b.captchas, ip); hit {
		return "captcha", s
	}
	return "", ""
}

func (b *bouncer) loop(ctx context.Context) {
	for {
		if err := b.pull(ctx, true); err != nil {
			log.Printf("startup pull failed: %v; retry in %s", err, b.interval)
			select {
			case <-ctx.Done():
				return
			case <-time.After(b.interval):
			}
			continue
		}
		b.ready.Store(true)
		break
	}
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := b.pull(ctx, false); err != nil {
				log.Printf("incremental pull failed: %v", err)
			}
		}
	}
}

// extractClientIP 从 APISIX forward-auth 透传的头部定位真实客户端 IP。
func extractClientIP(r *http.Request) net.IP {
	for _, h := range []string{"X-Original-Forwarded-For", "X-Forwarded-For", "X-Real-IP"} {
		if v := r.Header.Get(h); v != "" {
			if i := strings.Index(v, ","); i >= 0 {
				v = v[:i]
			}
			v = strings.TrimSpace(v)
			if ip := net.ParseIP(v); ip != nil {
				return ip
			}
		}
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return net.ParseIP(host)
}

// signCookie 返回 "<expiry_unix>.<hex hmac>" 格式的字符串。
// 不绑 IP：容器网络里客户端每跳 IP 不稳定；cookie HttpOnly 已有基础防盗。
func (b *bouncer) signCookie(expiry time.Time) string {
	exp := strconv.FormatInt(expiry.Unix(), 10)
	mac := hmac.New(sha256.New, b.cookieSecret)
	mac.Write([]byte(exp))
	return exp + "." + hex.EncodeToString(mac.Sum(nil))
}

// verifyCookie 返回 cookie 是否有效（HMAC 正确 + 未过期）。
func (b *bouncer) verifyCookie(v string) bool {
	parts := strings.SplitN(v, ".", 2)
	if len(parts) != 2 {
		return false
	}
	exp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || time.Now().Unix() >= exp {
		return false
	}
	mac := hmac.New(sha256.New, b.cookieSecret)
	mac.Write([]byte(parts[0]))
	want := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(want), []byte(parts[1]))
}

func (b *bouncer) cookieValid(r *http.Request) bool {
	c, err := r.Cookie(b.cookieName)
	if err != nil {
		return false
	}
	return b.verifyCookie(c.Value)
}

func (b *bouncer) handleCheck(w http.ResponseWriter, r *http.Request) {
	// 内部控制路径（captcha verify 回调）永远放行，避免 forward-auth 再触发挑战。
	if uri := r.Header.Get("X-Forwarded-Uri"); strings.HasPrefix(uri, "/__polaris/") {
		w.Header().Set("X-Polaris-Bouncer", "internal")
		w.WriteHeader(http.StatusOK)
		return
	}
	ip := extractClientIP(r)
	if os.Getenv("DEBUG") == "1" {
		log.Printf("/check headers=%v remote=%s picked=%v", r.Header, r.RemoteAddr, ip)
	}
	if ip == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	kind, scenario := b.match(ip)
	switch kind {
	case "ban":
		w.Header().Set("X-Polaris-Bouncer", "block")
		w.Header().Set("X-Polaris-Bouncer-Scenario", scenario)
		http.Error(w, "banned", http.StatusForbidden)
	case "captcha":
		if b.cookieValid(r) {
			w.Header().Set("X-Polaris-Bouncer", "captcha-pass")
			w.WriteHeader(http.StatusOK)
			return
		}
		b.writeChallenge(w, r, scenario)
	default:
		w.Header().Set("X-Polaris-Bouncer", "allow")
		w.WriteHeader(http.StatusOK)
	}
}

// challengeTmpl 是最小化的 Turnstile 挑战页。提交 token 到 /__polaris/captcha/verify，
// 成功后 JS 跳回 returnURL（APISIX 透传的 X-Forwarded-Uri）。
var challengeTmpl = template.Must(template.New("c").Parse(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <title>Polaris 验证 · {{.Scenario}}</title>
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <script src="https://challenges.cloudflare.com/turnstile/v0/api.js" async defer></script>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
           display: flex; align-items: center; justify-content: center;
           min-height: 100vh; margin: 0; background: #fafafa; }
    .card { background: #fff; padding: 40px 48px; border-radius: 8px;
            box-shadow: 0 2px 16px rgba(0,0,0,.08); max-width: 440px; text-align: center; }
    h1 { font-size: 18px; margin: 0 0 8px; color: #222; font-weight: 600; }
    p { margin: 0 0 24px; color: #666; font-size: 14px; line-height: 1.6; }
    .scenario { font-family: monospace; color: #999; font-size: 12px; margin-top: 16px; }
  </style>
</head>
<body>
  <div class="card">
    <h1>Polaris 安全验证</h1>
    <p>系统对此 IP 的近期请求标记为可疑，完成下方挑战后放行。</p>
    <form id="f" method="POST" action="/__polaris/captcha/verify">
      <input type="hidden" name="return_to" value="{{.ReturnTo}}">
      <div class="cf-turnstile" data-sitekey="{{.SiteKey}}" data-callback="onok"></div>
    </form>
    <div class="scenario">场景 · {{.Scenario}}</div>
  </div>
  <script>
    function onok(){ document.getElementById('f').submit(); }
  </script>
</body>
</html>`))

func (b *bouncer) writeChallenge(w http.ResponseWriter, r *http.Request, scenario string) {
	returnTo := r.Header.Get("X-Forwarded-Uri")
	if returnTo == "" {
		returnTo = "/"
	}
	w.Header().Set("X-Polaris-Bouncer", "challenge")
	w.Header().Set("X-Polaris-Bouncer-Scenario", scenario)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusUnauthorized)
	_ = challengeTmpl.Execute(w, map[string]string{
		"SiteKey":  b.turnstileSiteKey,
		"Scenario": scenario,
		"ReturnTo": returnTo,
	})
}

// handleVerify 接 POST form：`cf-turnstile-response` + `return_to`。
// 调 Cloudflare siteverify，通过后签 cookie，303 redirect 回 return_to。
func (b *bouncer) handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	token := r.FormValue("cf-turnstile-response")
	if token == "" {
		http.Error(w, "missing turnstile token", http.StatusBadRequest)
		return
	}
	returnTo := r.FormValue("return_to")
	if returnTo == "" || !strings.HasPrefix(returnTo, "/") {
		returnTo = "/"
	}
	clientIP := extractClientIP(r)
	if ok, detail := b.turnstileVerify(r.Context(), token, clientIP); !ok {
		log.Printf("turnstile verify failed: %s", detail)
		http.Error(w, "challenge failed", http.StatusForbidden)
		return
	}
	expiry := time.Now().Add(b.cookieTTL)
	http.SetCookie(w, &http.Cookie{
		Name:     b.cookieName,
		Value:    b.signCookie(expiry),
		Path:     "/",
		Expires:  expiry,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		// Secure 在有 TLS 的生产环境打开；dev HTTP 下设 true 浏览器不会存 cookie
		Secure: os.Getenv("CAPTCHA_COOKIE_SECURE") == "1",
	})
	http.Redirect(w, r, returnTo, http.StatusSeeOther)
}

type siteverifyResp struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

func (b *bouncer) turnstileVerify(ctx context.Context, token string, clientIP net.IP) (bool, string) {
	form := url.Values{}
	form.Set("secret", b.turnstileSecretKey)
	form.Set("response", token)
	if clientIP != nil {
		form.Set("remoteip", clientIP.String())
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://challenges.cloudflare.com/turnstile/v0/siteverify",
		strings.NewReader(form.Encode()))
	if err != nil {
		return false, err.Error()
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := b.client.Do(req)
	if err != nil {
		return false, "siteverify: " + err.Error()
	}
	defer resp.Body.Close()
	var sr siteverifyResp
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return false, "decode: " + err.Error()
	}
	if !sr.Success {
		return false, "cf error-codes: " + strings.Join(sr.ErrorCodes, ",")
	}
	return true, ""
}

func (b *bouncer) handleReady(w http.ResponseWriter, r *http.Request) {
	if b.ready.Load() {
		b.mu.RLock()
		nb, nc := len(b.bans), len(b.captchas)
		b.mu.RUnlock()
		fmt.Fprintf(w, "ready bans=%d captchas=%d\n", nb, nc)
		return
	}
	http.Error(w, "pulling initial snapshot", http.StatusServiceUnavailable)
}

func main() {
	lapi := envDefault("CROWDSEC_LAPI_URL", "http://crowdsec:8080")
	key := os.Getenv("CROWDSEC_BOUNCER_KEY")
	if key == "" {
		log.Fatal("CROWDSEC_BOUNCER_KEY required")
	}
	interval := parseDurationDefault(os.Getenv("CROWDSEC_STREAM_INTERVAL"), 10*time.Second)
	addr := envDefault("HTTP_ADDR", ":8080")
	if _, err := url.Parse(lapi); err != nil {
		log.Fatalf("bad CROWDSEC_LAPI_URL: %v", err)
	}

	b := newBouncer(lapi, key, interval)

	// Cloudflare 公开 dev 测试 key（总是通过），实际部署必须覆盖
	// https://developers.cloudflare.com/turnstile/troubleshooting/testing/
	b.turnstileSiteKey = envDefault("TURNSTILE_SITE_KEY", "1x00000000000000000000AA")
	b.turnstileSecretKey = envDefault("TURNSTILE_SECRET_KEY", "1x0000000000000000000000000000000AA")
	b.cookieTTL = parseDurationDefault(os.Getenv("CAPTCHA_COOKIE_TTL"), time.Hour)
	cookieSecret := os.Getenv("CAPTCHA_COOKIE_SECRET")
	if cookieSecret == "" {
		// dev fallback，生产一定要覆盖；用固定值以便重启后 cookie 不立即失效
		cookieSecret = "polaris-dev-cookie-secret-change-in-prod"
		log.Printf("WARNING: CAPTCHA_COOKIE_SECRET not set, using dev default")
	}
	// 用 sha256 把任意长度 secret 归一到 32 字节
	h := sha256.Sum256([]byte(cookieSecret))
	b.cookieSecret = h[:]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.loop(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/check", b.handleCheck)
	mux.HandleFunc("/captcha/verify", b.handleVerify) // APISIX 会把 /__polaris/captcha/verify 代理到这
	mux.HandleFunc("/ready", b.handleReady)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok\n"))
	})

	log.Printf("crowdsec-bouncer: lapi=%s interval=%s listen=%s turnstile_site=%s cookie_ttl=%s",
		lapi, interval.String(), addr, b.turnstileSiteKey, b.cookieTTL)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func envDefault(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func parseDurationDefault(s string, d time.Duration) time.Duration {
	if s == "" {
		return d
	}
	v, err := time.ParseDuration(s)
	if err != nil {
		return d
	}
	return v
}
