# ADR-0014: AI Gateway —— `ai-proxy-multi` + 按用户配额 + platform-admin 用量面板

## Status

Accepted（2026-04-24，Phase I MVP）

## Context

平台后续业务要统一接多家 LLM 提供商（OpenAI、Claude、DeepSeek、Qwen，以及未来自托管的 xinference），面临三类问题：

1. **路由/鉴权集中化**：每个业务方自己管 provider API key、做 retry/fallback、处理 streaming——重复建设且泄漏面大
2. **成本可见性**：没有统一视图看谁在烧 token，哪个 provider 在扛流量
3. **失控风险**：单个 user 的脚本 bug 可以把月度额度打空

ADR-0013 已经把 API Gateway 自建 UI 建起来（platform-admin）。这次要接一条新方向："AI 请求"经网关也走统一入口 + 有配额 + 有可观测。

## Decision

### 核心架构

```
 Caller ─→ APISIX ┬─ Authentik JWT 校验（identity 来源）
                  ├─ ai-proxy-multi：按模型名路由到不同 provider
                  │                   │
                  │                   ├─→ api.openai.com
                  │                   ├─→ api.anthropic.com
                  │                   ├─→ api.deepseek.com
                  │                   ├─→ dashscope.aliyuncs.com (Qwen)
                  │                   └─→ http://xinference:9997  （用户后续自部署）
                  ├─ ai-rate-limiting：按 JWT sub 做 token 硬限（Phase II）
                  └─ elasticsearch-logger：每次调用 → `ai-usage` index
                                           │
                 platform-admin ─────────→ 读 ai-usage 渲染 Usage Dashboard
                                  └────→ Providers / Quotas 只读或 PATCH（运行时参数）
```

**不新建独立 AI Gateway 服务**。APISIX 3.x 原生 `ai-proxy-multi` 足够做路由与 provider 抽象，quota 与用量分析用现有 `ai-rate-limiting` + ES + platform-admin 复用，与 ADR-0013 的"结构配置走 Git、运行时走 UI"模式一致。

### Phase I MVP（本 ADR 范围）

下列在代码侧实现；合起来预计 3-4 天：

| 条目 | 做什么 | 不做什么 |
|---|---|---|
| **APISIX 路由** | 新加 `components/apisix/routes/20-ai-gateway.yaml`：`/ai/v1/chat/completions` + `/ai/v1/embeddings` 走 `ai-proxy-multi`，按模型名分发到 4 个 provider（OpenAI/Claude/DeepSeek/Qwen） | 不做自动 fallback、weighted routing |
| **Dev Mock** | 起一个 mock OpenAI 兼容容器（`services/ai-mock/`，Go ~80 行），本地冒烟不依赖外部 key | 不做 Anthropic / Qwen 协议 mock，MVP 只覆盖 OpenAI 形态 |
| **JWT 鉴权** | 复用 APISIX `openid-connect` 插件，issuer 指向 Authentik 的 platform-admin provider | 不做 per-application API key（Phase III） |
| **用量日志** | `elasticsearch-logger` 写 `ai-usage` index，字段：`sub`（用户）、`model`、`provider`、`prompt_tokens`、`completion_tokens`、`status`、`latency_ms` | 不做 cost ¥ 换算（Phase III 接 pricing 表） |
| **UI "AI 代理" 菜单** | platform-admin 加菜单，3 个只读页：Providers（配置展示）、Usage（按 user/model/provider 聚合）、Quotas（占位） | 不做 Quota CRUD（Phase II） |

### 决定项

- **路由端点前缀** `/ai/v1/*`（不是 `/v1/ai/*`）：AI 生态演进节奏显著快于业务 API（streaming 协议、tool use、function call 规范半年多变），domain-first 让 AI 有独立版本轴；流量 / 成本 / WAF 规则 / 限流也天然按 `/ai/*` 前缀聚合（Cloudflare AI Gateway / Portkey / Helicone 的惯例）。客户端把 baseURL 设成 `https://<gateway>/ai/v1/` 即可
- **provider 选择策略**：按 request body 里的 `model` 字段。`model: gpt-4o*` → OpenAI，`model: claude-*` → Claude，`model: deepseek-*` → DeepSeek，`model: qwen*` → Qwen，`model: local/*` → xinference（用户自托管）。规则在 `ai-proxy-multi` 的 `providers` block 里声明
- **认证主体**：Phase I 只接 Authentik JWT（`sub` claim 当 user id）；Phase III 再加 per-application API key（用 APISIX consumer + key-auth 插件）
- **配额维度**：**按 user**（JWT sub），Phase II 实现
- **API key 保管**：provider 上游 key 存 APISIX etcd（加密 at-rest 由 etcd 本身保证；长期走 `secret://vault/*` —— ADR-0007 的演进）。Git 源里只写占位符，apply 阶段替换
- **用量日志 index**：`ai-usage`（独立于 `apisix-access`），ILM 按天滚动，保留 30d
- **不引入独立 AI Gateway 服务**：APISIX 插件 + platform-admin 已够；独立服务是反向投资

### 对前端代码的 API 约定

所有 LLM 调用统一走：

```
POST /ai/v1/chat/completions
Authorization: Bearer <Authentik ID token>
Content-Type: application/json

{"model": "gpt-4o-mini", "messages": [...]}
```

响应形态与 OpenAI 一致（`ai-proxy-multi` 把上游各家的响应转成 OpenAI schema）。业务代码一律用 OpenAI SDK，把 baseURL 换成我们的网关即可。

## Consequences

**易**：

- 业务方接 LLM 不再自己管 API key / 协议差异 / 计量
- 成本可见：platform-admin 一眼看到谁烧 token，告警可以挂上
- 换 provider 只改 APISIX 路由，业务代码零改动
- 已有 Coraza + CrowdSec bouncer 自动保护 AI 端点（脚本注入、爬虫抓取会被拦）

**难 / 代价**：

- `ai-proxy-multi` 的 provider 兼容性是黑盒——Claude streaming / Qwen 的非标响应字段可能踩坑，需要跟 APISIX 版本升级
- 按 user quota 跨 APISIX 实例共享需要 etcd 做集中计数（`ai-rate-limiting` 内置支持）；高并发时 etcd 写放大要监控
- 用量日志走 ES，ES 故障时日志丢失（Phase III 补兜底持久化）
- Mock OpenAI server 仅覆盖 `chat/completions` + `embeddings`；streaming 语义要单独 mock

## Phase II / III 预览（本 ADR 不实现，列入 backlog）

- **Phase II — Quota 硬限**：
  - `ai-rate-limiting` 按 JWT sub 限 tokens/day；platform-admin UI 加 Quota CRUD（PATCH APISIX 插件 config）
  - 超限返 429 + `Retry-After`；告警接入 CrowdSec（连续超限 IP 自动进 captcha）
- **Phase III — 成本与治理**：
  - pricing 表（yaml）→ usage 日志结合算 ¥/$
  - per-application API key（APISIX consumer + key-auth），业务代码换成 app key 而非 user JWT（服务间调用场景）
  - provider 故障自动切流（`ai-proxy-multi` 的 health + fallback）
  - Prometheus metrics（token throughput、provider p99、error rate）
  - 告警：连续失败率、配额接近、provider 不可达
- **自托管部署**：用户后续 `xinference` 本地 API 进入 provider 列表，路由规则 `model: local/*` → `http://xinference:9997`

## Alternatives Considered

1. **独立 Go AI Gateway 服务（like LiteLLM/Helicone-style 自建）**
   优点：完全可控，定制化强
   缺点：重复 `ai-proxy-multi` 已有能力；又多一层维护；与 ADR-0013 的"APISIX 原生插件优先"不一致
   → **否**

2. **APISIX + 独立配额服务**（APISIX 只做路由，quota 走一个 Go sidecar 类似 crowdsec-bouncer 那样）
   优点：配额逻辑更灵活
   缺点：APISIX `ai-rate-limiting` 已经支持 token 计数 + etcd 共享；自建是重复投入
   → **否**（Phase II 若 `ai-rate-limiting` 不满足再评估）

3. **直接让业务代码调各 provider SDK，不走网关**
   优点：最简单
   缺点：没有统一鉴权 / 成本 / 配额；与 AI-专属威胁防护（ADR-0012 的 token 限流防模型窃取）脱节
   → **否**

4. **只做 OpenAI 兼容一家**（ChatGPT 网关）
   优点：代码最少
   缺点：锁定一家 provider，国产合规场景（数据不出境）无解
   → **否**，MVP 就定义多 provider 架构
