// crowdsec-bouncer — APISIX forward-auth 侧车（ADR-0012 行为层 P2）。
//
// 工作方式：
//  1. 启动时 GET /v1/decisions/stream?startup=true 取 LAPI 全量 ban 列表；
//     之后每 CROWDSEC_STREAM_INTERVAL 拉一次增量（startup=false），合并到内存 banlist。
//  2. 暴露 GET /check：读 X-Original-Forwarded-For 第一跳 IP；命中返 403，未命中返 200。
//  3. APISIX forward-auth 插件按 /check 的 status 放行或拦截。
//
// 选型注记：不走 lua-resty-crowdsec 的 APISIX 插件路径，因为：
//   - 我们的栈本来就是 Go，新增一份维护更轻；
//   - APISIX 插件目录侵入性强（改 config.yaml + 增 Lua 源码包），CI / 升级面更大；
//   - forward-auth 模型语义清晰，将来换成别的决策源（WAF / 自研策略引擎）不用改 APISIX 侧。
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Decision 复刻 LAPI /v1/decisions/stream 返回的精简字段。
// 仅关心 ip / range scope + ban type；captcha / throttle 交给 Turnstile 路径（P2 后续）。
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

// entry 是 banlist 的单条：到期时间 + scope（Ip / Range）+ 原因。
type entry struct {
	expireAt time.Time
	isRange  bool
	cidr     *net.IPNet
	ip       net.IP
	scenario string
}

type bouncer struct {
	lapiURL   string
	apiKey    string
	interval  time.Duration
	client    *http.Client

	mu    sync.RWMutex
	bans  map[int64]*entry
	ready atomic.Bool
}

func newBouncer(lapiURL, apiKey string, interval time.Duration) *bouncer {
	return &bouncer{
		lapiURL:  strings.TrimRight(lapiURL, "/"),
		apiKey:   apiKey,
		interval: interval,
		client:   &http.Client{Timeout: 10 * time.Second},
		bans:     map[int64]*entry{},
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
		// 全量快照：清空旧缓存。
		b.bans = make(map[int64]*entry, len(body.New))
	}
	added, removed := 0, 0
	for _, d := range body.New {
		// LAPI 只回 ban 类决策？不一定，显式过滤。
		if strings.ToLower(d.Type) != "ban" {
			continue
		}
		e := parseDecision(d)
		if e == nil {
			continue
		}
		b.bans[d.ID] = e
		added++
	}
	for _, d := range body.Deleted {
		if _, ok := b.bans[d.ID]; ok {
			delete(b.bans, d.ID)
			removed++
		}
	}
	log.Printf("lapi stream (startup=%t): +%d -%d total=%d", startup, added, removed, len(b.bans))
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
		// country / as / username 等此处不处理（与 IP 检查不匹配）。
		return nil
	}
}

// match 判定某 IP 是否命中任一未过期 ban。
func (b *bouncer) match(ip net.IP) (bool, string) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	now := time.Now()
	for _, e := range b.bans {
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

// loop 每隔 interval 拉增量。内联 startup 首次，失败会重试。
func (b *bouncer) loop(ctx context.Context) {
	// 首次拉全量，失败不退出（LAPI 可能还没起来），按 interval 重试直到成功。
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
// APISIX 默认把原请求 IP 放在 X-Forwarded-For 第一跳；此处也兼容 X-Real-IP。
func extractClientIP(r *http.Request) net.IP {
	for _, h := range []string{"X-Original-Forwarded-For", "X-Forwarded-For", "X-Real-IP"} {
		if v := r.Header.Get(h); v != "" {
			// XFF 可能是逗号分隔列表，取第一跳
			if i := strings.Index(v, ","); i >= 0 {
				v = v[:i]
			}
			v = strings.TrimSpace(v)
			if ip := net.ParseIP(v); ip != nil {
				return ip
			}
		}
	}
	// 退化到 TCP 源（一般是 APISIX 自身，没什么意义，但返回方便调试）
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return net.ParseIP(host)
}

func (b *bouncer) handleCheck(w http.ResponseWriter, r *http.Request) {
	ip := extractClientIP(r)
	if os.Getenv("DEBUG") == "1" {
		log.Printf("/check headers=%v remote=%s picked=%v", r.Header, r.RemoteAddr, ip)
	}
	if ip == nil {
		// 没有 IP 就放行；APISIX 侧应总会注入 XFF。
		w.WriteHeader(http.StatusOK)
		return
	}
	banned, scenario := b.match(ip)
	if banned {
		w.Header().Set("X-Polaris-Bouncer", "block")
		w.Header().Set("X-Polaris-Bouncer-Scenario", scenario)
		http.Error(w, "banned", http.StatusForbidden)
		return
	}
	w.Header().Set("X-Polaris-Bouncer", "allow")
	w.WriteHeader(http.StatusOK)
}

func (b *bouncer) handleReady(w http.ResponseWriter, r *http.Request) {
	if b.ready.Load() {
		b.mu.RLock()
		n := len(b.bans)
		b.mu.RUnlock()
		fmt.Fprintf(w, "ready bans=%d\n", n)
		return
	}
	http.Error(w, "pulling initial snapshot", http.StatusServiceUnavailable)
}

func main() {
	lapi := os.Getenv("CROWDSEC_LAPI_URL")
	if lapi == "" {
		lapi = "http://crowdsec:8080"
	}
	key := os.Getenv("CROWDSEC_BOUNCER_KEY")
	if key == "" {
		log.Fatal("CROWDSEC_BOUNCER_KEY required")
	}
	intervalStr := os.Getenv("CROWDSEC_STREAM_INTERVAL")
	interval := 10 * time.Second
	if intervalStr != "" {
		if d, err := time.ParseDuration(intervalStr); err == nil {
			interval = d
		}
	}
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	if _, err := url.Parse(lapi); err != nil {
		log.Fatalf("bad CROWDSEC_LAPI_URL: %v", err)
	}

	b := newBouncer(lapi, key, interval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.loop(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/check", b.handleCheck)
	mux.HandleFunc("/ready", b.handleReady)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok\n"))
	})

	log.Printf("crowdsec-bouncer: lapi=%s interval=%s listen=%s",
		lapi, interval.String(), addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
