# crowdsec-bouncer — APISIX forward-auth 侧车

> 决策见 ADR-0012（CrowdSec 行为层 + bouncer 生效路径）。

把 CrowdSec LAPI 的 decisions 转成 APISIX 网关处能执行的 allow / block。
按请求由 APISIX `forward-auth` 插件调 `/check`，status code 决定放行还是拦截。

## 工作方式

```
客户端 → APISIX ┬─ forward-auth ─→ crowdsec-bouncer:/check ─→ (内存 banlist 扫描)
                └─ 放行则继续走 Coraza + 业务路由
```

- 启动时 `GET /v1/decisions/stream?startup=true` 拉全量快照
- 之后每 `CROWDSEC_STREAM_INTERVAL`（默认 10s）拉增量（`startup=false`）
- 只收 `type=ban` 的决策；`captcha` / `throttle` 交给后续 Turnstile 路径
- `/check` 依次读 `X-Original-Forwarded-For` / `X-Forwarded-For` / `X-Real-IP` 第一跳作客户端 IP
- 命中返 403 + `X-Polaris-Bouncer: block` + `X-Polaris-Bouncer-Scenario: <scenario>`；未命中返 200

## 环境变量

| 变量 | 默认 | 说明 |
|------|------|------|
| `CROWDSEC_LAPI_URL` | `http://crowdsec:8080` | LAPI 地址（容器内用服务名） |
| `CROWDSEC_BOUNCER_KEY` | —（必填） | 与 `components/crowdsec` 的 `BOUNCER_KEY_apisix` 对齐 |
| `CROWDSEC_STREAM_INTERVAL` | `10s` | 增量轮询间隔，生产可拉长到 30s~60s |
| `HTTP_ADDR` | `:8080` | 监听地址 |
| `DEBUG` | `0` | `1` 时打印每次 `/check` 的头部，仅排障 |

## 冒烟

```bash
# 1. 封禁 docker 桥 IP（APISIX 对本机流量看到的源）
docker exec polaris-base-crowdsec-1 cscli decisions add -i 172.20.0.1 -d 5m -R smoke

# 2. 等 10s 让 bouncer 拉到增量
sleep 12

# 3. 经 APISIX 请求应被挡
curl -i http://localhost:9080/health
# HTTP/1.1 403 Forbidden
# X-Polaris-Bouncer: block
# X-Polaris-Bouncer-Scenario: smoke

# 4. 解封
docker exec polaris-base-crowdsec-1 cscli decisions delete -i 172.20.0.1
```

## 不做

- **challenge / captcha**：Turnstile 集成在独立路径（ADR-0012 未来）
- **per-request LAPI 查询**：会给 LAPI 打出 N 倍 QPS；始终走 stream 缓存
- **Prometheus metrics**：P3 时和 APISIX 指标一起纳入 observability 栈

## 约束

- 必须能连到 LAPI；连不上时首次拉取会阻塞重试直到成功（此期间 `/ready` 返 503，`/check` 全放行 —— fail-open）
- 决策过期时间按 LAPI 的 `duration` 解析；解析失败退化到 24h 后内存清理
