# ADR-0012: WAF——Coraza-in-APISIX + CrowdSec 行为层

## Status

Accepted（2026-04-19，2026-04-20 补"行为层与 SaaS 分层"）

## Context

基座需要多层次 WAF 防护：
- **请求层**（注入/XSS/RCE/路径穿越）：特征规则可覆盖
- **行为层**（扫描器、爆破、爬虫、CVE 批量利用、已知恶意 IP）：需日志分析 + 社区威胁情报
- **人机验证**：可疑请求挑战验证
- **设备识别**：跨会话追踪

## Decision

**分层组合，按覆盖能力选择部署形态：**

| 层 | 引擎 | 部署形态 | 理由 |
|----|------|---------|------|
| 请求层 WAF | **Coraza** + OWASP CRS | APISIX Wasm 插件（无独立容器） | CNCF Sandbox，APISIX 原生集成，进程内零跳 |
| 行为层 | **CrowdSec** CE | 独立容器（`components/crowdsec/`） | 开源 + 社区威胁情报（CTI）；日志分析类必须独立进程 |
| 人机验证 | **Cloudflare Turnstile** | SaaS（APISIX 插件侧调 API） | 免费、隐私友好、无容器 |
| 设备指纹 | **FingerprintJS Pro** | SaaS（前端 SDK + 后端验签） | OSS 自托管无强选手；Pro SaaS 识别率高 |

管理 UI 在 `services/platform-admin/`（ADR-0013）。

### Coraza（请求层）

- `coraza-proxy-wasm` v0.6.0 Wasm 插件，进程内运行
  - 二进制由 `apisix-coraza-init` 首次启动下载至 `apisix_wasm` volume，APISIX 只读挂载（`components/apisix/docker-compose.yml`）
  - 插件在 `config.yaml` `wasm.plugins` 注册，priority 7999，access 阶段
- OWASP CRS 3.x 规则集嵌入 wasm 自带，通过 `Include @owasp_crs/*.conf` 启用
- 规则启停 / 路由例外走 Git 源（`components/apisix/routes/95-coraza.yaml`）→ PR review → CI apply；UI 编辑器 P2 再评估（ADR-0013 决策）
- 访问/攻击日志经 APISIX `elasticsearch-logger` 插件推送至 ES（`routes/90-access-log.yaml`）

#### MVP 限制与后续演进

`coraza-proxy-wasm` 默认不将命中规则信息（`rule_id` / `severity` / `matched_data`）写入 APISIX 请求上下文，仅发到 stderr 审计日志。因此 P1 攻击日志页按 `status >= 400` 过滤，`Rule ID` 等字段留空。

演进路径（任一即可）：
1. 自定义 Coraza WASM 打包：在 `onRequestHeaders` 阶段将命中信息回写 response header（如 `X-Polaris-Coraza-RuleId`），`elasticsearch-logger` 通过自定义字段透传
2. ES ingest pipeline：订阅 APISIX stderr 的 Coraza 审计行，grok 解析成独立 index，按 `request_id` 与访问日志 join

### CrowdSec（行为层）

- 日志源：APISIX 经 `http_configuration_snippet` 额外写一份 combined 格式 `access.log` 到共享卷 `apisix_logs`（`source: file`）。WSL2 / 部分 Docker Desktop 下 `source: docker` 的 daemon API 协议有兼容问题，故统一走 file。
- 默认集合：`crowdsecurity/nginx` + `base-http-scenarios` + `http-cve`
- 社区 CTI：自动订阅社区黑名单（已知扫描器 / 僵尸网络 IP）
- K8s 迁移时改用 DaemonSet + file tailer 指向共享 EmptyDir 即可

#### bouncer 生效路径（2026-04-23 P2 落地）

APISIX 通过 [forward-auth 插件](https://apisix.apache.org/docs/apisix/plugins/forward-auth/) 按请求调侧车 `services/crowdsec-bouncer`（Go）。
bouncer 在内存里维护 LAPI `/v1/decisions/stream` 的增量快照（启动时 `startup=true` 全量、之后 10s `startup=false` 增量），每请求 `O(N)` 扫描未过期 ban（N 通常 <万量级，单请求延迟 <1ms）。

**为何不走官方 `lua-resty-crowdsec` 直接塞进 APISIX 插件目录**：
- 栈统一：我们已在 Go，新增一份 Lua 源码 + 插件注册会扩大 CI / 升级面；
- 解耦：forward-auth 语义清晰，将来替换决策源（AI Gateway token 配额、自研策略引擎）不动 APISIX 侧；
- 失败语义可选：bouncer 不可达时 APISIX forward-auth 默认 fail-open（ADR-0012 可用性优于风控），生产可通过降低 `timeout` + 告警收紧。

**客户端 IP 解析**：APISIX 默认用 TCP 源地址作 `$http_x_forwarded_for`；生产部署在 LB / CDN 后需配 `nginx_config.http.real_ip_header: X-Forwarded-For` 和 `real_ip_from` 信任网段，否则 bouncer 看到的永远是 LB IP，决策失效。

### Turnstile（人机验证）

- APISIX 自定义插件：
  - 对可疑请求（CrowdSec 标记但未封禁 / 命中特定规则）返回 Turnstile challenge 页
  - 回调验证 token 后放行
- 无需额外容器

### FingerprintJS Pro（设备识别）

- 前端 SDK 生成 visitorId → 经 `X-Device-FP` header 上送
- APISIX 插件调 FingerprintJS Server API 验签
- 设备级限流（IP+FP 组合）通过 Admin API 下发
- 成本敏感时可延后启用

## AI 专属威胁分工

| 威胁 | 归属层 | 应对 |
|------|-------|------|
| Prompt Injection | 应用层 | 业务服务内 prompt 过滤 |
| 模型窃取（高频抓取） | **CrowdSec** + AI Gateway token 限流 | 行为检测 + TPM 配额 |
| 训练数据投毒 | MLOps 安全域 | 本 ADR 范围外 |

## Consequences

- 基础 WAF 零额外容器（Coraza Wasm）
- 行为层新增一个容器（CrowdSec）+ 挂载 `docker.sock`（单机限制）
- 人机验证/指纹走 SaaS，运营成本换研发成本
- CRS 规则需跟进版本更新
- Turnstile / FingerprintJS SaaS 属外部依赖，网络隔离场景不可用
