# crowdsec-bouncer — APISIX forward-auth 侧车

> 决策见 ADR-0012（CrowdSec 行为层 + bouncer 生效路径）。

把 CrowdSec LAPI 的 decisions 转成 APISIX 网关处能执行的 allow / block。
按请求由 APISIX `forward-auth` 插件调 `/check`，status code 决定放行还是拦截。

## 工作方式

```
客户端 → APISIX ┬─ forward-auth ─→ bouncer:/check ──┬─ ban     → 401/403 返客户端（挑战页 / banned）
                │                                   ├─ captcha → 401 + Turnstile 挑战页
                │                                   ├─ 有有效 cookie → 200
                │                                   └─ 无匹配    → 200
                ├─ /__polaris/captcha/verify 路由 → bouncer:/captcha/verify
                │                                   └─ CF siteverify → Set-Cookie → 303 Redirect
                └─ 放行则继续走 Coraza + 业务路由
```

- 启动时 `GET /v1/decisions/stream?startup=true` 拉全量快照；之后每 `CROWDSEC_STREAM_INTERVAL`（默认 10s）拉增量
- 按决策 `type` 分两池：`ban` 直接 403；`captcha` 检查 cookie，无效则返挑战页
- `/check` 依次读 `X-Original-Forwarded-For` / `X-Forwarded-For` / `X-Real-IP` 第一跳作客户端 IP
- 挑战页内嵌 Turnstile 组件，token 提交到 `/__polaris/captcha/verify` → bouncer 调 CF siteverify → 签名 cookie `polaris_captcha_pass = <exp>.<hmac(exp)>`
- HMAC 密钥 `CAPTCHA_COOKIE_SECRET` 不绑 IP（容器链路里 IP 不稳定），靠 `HttpOnly` + TTL 短化控泄漏面

## 环境变量

| 变量 | 默认 | 说明 |
|------|------|------|
| `CROWDSEC_LAPI_URL` | `http://crowdsec:8080` | LAPI 地址（容器内用服务名） |
| `CROWDSEC_BOUNCER_KEY` | —（必填） | 与 `components/crowdsec` 的 `BOUNCER_KEY_apisix` 对齐 |
| `CROWDSEC_STREAM_INTERVAL` | `10s` | 增量轮询间隔，生产可拉长到 30s~60s |
| `HTTP_ADDR` | `:8080` | 监听地址 |
| `DEBUG` | `0` | `1` 时打印每次 `/check` 的头部，仅排障 |
| `TURNSTILE_SITE_KEY` | `1x00000000000000000000AA` | Cloudflare Turnstile 站点 key；默认是 CF 官方 dev 测试 key（总是通过），生产必须覆盖 |
| `TURNSTILE_SECRET_KEY` | `1x0000000000000000000000000000000AA` | 同上，siteverify 侧 key |
| `CAPTCHA_COOKIE_SECRET` | dev 占位字符串 | HMAC 签名密钥（sha256 归一到 32 字节）；生产必须覆盖，否则 cookie 可被伪造 |
| `CAPTCHA_COOKIE_TTL` | `1h` | 通过挑战后 cookie 的有效期 |
| `CAPTCHA_COOKIE_SECURE` | `0` | 生产（HTTPS）设 `1`，浏览器只在 TLS 下回传 cookie |

## 冒烟

### ban 路径

```bash
docker exec polaris-base-crowdsec-1 cscli decisions add -i 172.20.0.1 -d 5m -R smoke
sleep 12
curl -i http://localhost:9080/health
# HTTP/1.1 403 Forbidden
# X-Polaris-Bouncer: block
docker exec polaris-base-crowdsec-1 cscli decisions delete -i 172.20.0.1
```

### captcha 路径

```bash
docker exec polaris-base-crowdsec-1 cscli decisions add -i 172.20.0.1 -d 5m -R captcha_smoke -t captcha
sleep 12

# 首请求返挑战页
curl -i http://localhost:9080/health    # 401 + X-Polaris-Bouncer: challenge + Turnstile HTML

# 模拟提交（CF dev 测试 key 下 token 值任意，siteverify 必过）
curl -i -c cookies.txt \
  -d "cf-turnstile-response=dev&return_to=/health" \
  http://localhost:9080/__polaris/captcha/verify
# 303 See Other, Set-Cookie: polaris_captcha_pass=<exp>.<hmac>

# 带 cookie 再请求，通行
curl -b cookies.txt -i http://localhost:9080/health   # HTTP 200

docker exec polaris-base-crowdsec-1 cscli decisions delete -i 172.20.0.1
```

## 不做

- **per-request LAPI 查询**：会给 LAPI 打出 N 倍 QPS；始终走 stream 缓存
- **IP 绑定的 cookie**：容器链路 IP 不稳；挑战 cookie 只保"客户端曾解过题"，短 TTL + HttpOnly
- **Prometheus metrics**：P3 时和 APISIX 指标一起纳入 observability 栈

## 约束

- 必须能连到 LAPI；连不上时首次拉取会阻塞重试直到成功（此期间 `/ready` 返 503，`/check` 全放行 —— fail-open）
- 决策过期时间按 LAPI 的 `duration` 解析；解析失败退化到 24h 后内存清理
