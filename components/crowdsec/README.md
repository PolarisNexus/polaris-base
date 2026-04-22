# CrowdSec — 行为层 Bot / 恶意 IP 检测

> ADR-0012 WAF 分层中的行为层。请求层 Coraza 在 APISIX 进程内，另有 Turnstile / FingerprintJS Pro 两个 SaaS 层。

## 集成路径

- **日志源**：通过 `acquis.yaml` 用 `docker` 数据源订阅 `apisix` 容器 stdout（零文件挂载）
- **决策消费**（P2）：APISIX 侧部署 Lua bouncer（参考 [cs-lua-bouncer](https://github.com/crowdsecurity/cs-lua-bouncer)），每请求查 `http://crowdsec:8080/v1/decisions`，认证头 `X-Api-Key: ${CROWDSEC_BOUNCER_KEY}`
- **管理控制台**：platform-admin P1 已覆盖 LAPI 管理面（决策 CRUD + 告警查询，走 machine 账号 JWT，见 ADR-0013）

## 默认集合

`crowdsecurity/nginx` + `base-http-scenarios` + `http-cve`，覆盖扫描器、爆破、通用 HTTP 攻击、CVE 批量探测。扩展见 https://hub.crowdsec.net/

## 运维

```bash
docker compose exec crowdsec cscli decisions list
docker compose exec crowdsec cscli alerts list
docker compose exec crowdsec cscli decisions delete --ip 1.2.3.4
docker compose exec crowdsec cscli collections install crowdsecurity/whitelist-good-actors
```

## 约束

- 生产必须覆盖 `CROWDSEC_BOUNCER_KEY` 和 `CROWDSEC_AGENT_PASSWORD`（`openssl rand -hex 32`）
- 挂载 `docker.sock` 仅读，限单机；K8s 迁移改 DaemonSet + journald 数据源

## 管理账号

容器启动时根据 `AGENT_USERNAME` / `AGENT_PASSWORD` 自动注册一个 machine 账号（默认 `platform-admin`），供管理控制台走 `/v1/watchers/login` 拿 JWT 调管理类 API（创建/删除决策、查告警）。

bouncer key 与 machine 账号职责分离：bouncer 只查 decisions，machine 账号具备写权限。
