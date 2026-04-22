package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Client 是 CrowdSec LAPI 管理端的最小客户端。
// LAPI 分两种凭据：
//   - machine（username+password → /v1/watchers/login 拿 JWT）——读写 /v1/alerts、删除 decision
//   - bouncer（X-Api-Key）——读 /v1/decisions（GET list / stream）
//
// JWT 默认 1h，这里提前 5 分钟刷新。
type Client struct {
	baseURL    string
	username   string
	password   string
	bouncerKey string
	http       *http.Client

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

func NewClient(baseURL, username, password, bouncerKey string) *Client {
	return &Client{
		baseURL:    baseURL,
		username:   username,
		password:   password,
		bouncerKey: bouncerKey,
		http:       &http.Client{Timeout: 10 * time.Second},
	}
}

type loginResp struct {
	Code   int    `json:"code"`
	Token  string `json:"token"`
	Expire string `json:"expire"`
}

func (c *Client) login(ctx context.Context) error {
	body, _ := json.Marshal(map[string]string{
		"machine_id": c.username,
		"password":   c.password,
	})
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/watchers/login", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("crowdsec login: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("crowdsec login: %d %s", resp.StatusCode, raw)
	}
	var lr loginResp
	if err := json.Unmarshal(raw, &lr); err != nil {
		return err
	}
	exp, err := time.Parse(time.RFC3339, lr.Expire)
	if err != nil {
		exp = time.Now().Add(55 * time.Minute)
	}
	c.token = lr.Token
	c.expiresAt = exp
	return nil
}

func (c *Client) authToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != "" && time.Until(c.expiresAt) > 5*time.Minute {
		return c.token, nil
	}
	if err := c.login(ctx); err != nil {
		return "", err
	}
	return c.token, nil
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any) ([]byte, error) {
	tok, err := c.authToken(ctx)
	if err != nil {
		return nil, err
	}
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 401 {
		// token 失效一次性重试：清空再重登。
		c.mu.Lock()
		c.token = ""
		c.mu.Unlock()
		tok, err = c.authToken(ctx)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+tok)
		resp, err = c.http.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		raw, _ = io.ReadAll(resp.Body)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("crowdsec %s %s: %d %s", method, path, resp.StatusCode, raw)
	}
	return raw, nil
}

// Decision / Alert 是 LAPI 返回的精简模型，命名与 API 对齐。
type Decision struct {
	ID        int64  `json:"id"`
	Origin    string `json:"origin"`
	Type      string `json:"type"`
	Scope     string `json:"scope"`
	Value     string `json:"value"`
	Scenario  string `json:"scenario"`
	Duration  string `json:"duration"`
	Simulated bool   `json:"simulated"`
}

type Alert struct {
	ID           int64       `json:"id"`
	MachineID    string      `json:"machine_id"`
	Scenario     string      `json:"scenario"`
	ScenarioHash string      `json:"scenario_hash"`
	EventsCount  int32       `json:"events_count"`
	StartAt      string      `json:"start_at"`
	StopAt       string      `json:"stop_at"`
	Message      string      `json:"message"`
	Source       AlertSource `json:"source"`
	Decisions    []Decision  `json:"decisions"`
}

type AlertSource struct {
	IP    string `json:"ip"`
	Scope string `json:"scope"`
}

func (c *Client) ListDecisions(ctx context.Context, q url.Values) ([]Decision, error) {
	raw, err := c.doBouncer(ctx, "GET", "/v1/decisions", q)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var out []Decision
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// doBouncer 走 bouncer 凭据（X-Api-Key），仅用于 decisions 读路径。
func (c *Client) doBouncer(ctx context.Context, method, path string, query url.Values) ([]byte, error) {
	if c.bouncerKey == "" {
		return nil, fmt.Errorf("crowdsec: bouncer key not configured; /v1/decisions requires X-Api-Key")
	}
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", c.bouncerKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("crowdsec %s %s: %d %s", method, path, resp.StatusCode, raw)
	}
	return raw, nil
}

// CreateDecision 手动封禁。LAPI 要求提交 alert 结构（含 1 条 decision）。
type createAlertReq struct {
	Message         string     `json:"message"`
	Scenario        string     `json:"scenario"`
	ScenarioHash    string     `json:"scenario_hash"`
	ScenarioVersion string     `json:"scenario_version"`
	Source          AlertSrc   `json:"source"`
	StartAt         string     `json:"start_at"`
	StopAt          string     `json:"stop_at"`
	Capacity        int        `json:"capacity"`
	EventsCount     int        `json:"events_count"`
	Leakspeed       string     `json:"leakspeed"`
	Simulated       bool       `json:"simulated"`
	Decisions       []createDec `json:"decisions"`
}

type AlertSrc struct {
	Scope string `json:"scope"`
	Value string `json:"value"`
}

type createDec struct {
	Duration string `json:"duration"`
	Origin   string `json:"origin"`
	Scenario string `json:"scenario"`
	Scope    string `json:"scope"`
	Type     string `json:"type"`
	Value    string `json:"value"`
}

func (c *Client) CreateDecision(ctx context.Context, scope, value, decisionType, duration, reason string) ([]int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	alert := createAlertReq{
		Message:         reason,
		Scenario:        "manual/" + decisionType,
		ScenarioHash:    "",
		ScenarioVersion: "",
		Source:          AlertSrc{Scope: scope, Value: value},
		StartAt:         now,
		StopAt:          now,
		Capacity:        0,
		EventsCount:     1,
		Leakspeed:       "0",
		Simulated:       false,
		Decisions: []createDec{{
			Duration: duration,
			Origin:   "platform-admin",
			Scenario: reason,
			Scope:    scope,
			Type:     decisionType,
			Value:    value,
		}},
	}
	if _, err := c.do(ctx, "POST", "/v1/alerts", nil, []createAlertReq{alert}); err != nil {
		return nil, err
	}
	// /v1/alerts 返回的是 alertID 列表，不是 decisionID；调用方按 scope+value 回查决策即可。
	return nil, nil
}

func (c *Client) DeleteDecisionByID(ctx context.Context, id int64) (int, error) {
	return c.deleteAndParse(ctx, fmt.Sprintf("/v1/decisions/%d", id), nil)
}

func (c *Client) DeleteDecisionByScopeValue(ctx context.Context, scope, value string) (int, error) {
	q := url.Values{}
	q.Set("scope", scope)
	q.Set("value", value)
	return c.deleteAndParse(ctx, "/v1/decisions", q)
}

type deleteResp struct {
	NbDeleted string `json:"nbDeleted"`
}

func (c *Client) deleteAndParse(ctx context.Context, path string, q url.Values) (int, error) {
	raw, err := c.do(ctx, "DELETE", path, q, nil)
	if err != nil {
		return 0, err
	}
	if len(raw) == 0 {
		return 0, nil
	}
	var r deleteResp
	if err := json.Unmarshal(raw, &r); err != nil {
		return 0, nil
	}
	n := 0
	fmt.Sscanf(r.NbDeleted, "%d", &n)
	return n, nil
}

func (c *Client) ListAlerts(ctx context.Context, q url.Values) ([]Alert, error) {
	raw, err := c.do(ctx, "GET", "/v1/alerts", q, nil)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var out []Alert
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetAlert(ctx context.Context, id int64) (*Alert, error) {
	raw, err := c.do(ctx, "GET", fmt.Sprintf("/v1/alerts/%d", id), nil, nil)
	if err != nil {
		return nil, err
	}
	var a Alert
	if err := json.Unmarshal(raw, &a); err != nil {
		return nil, err
	}
	return &a, nil
}
