# ADR-0002: APISIX etcd 模式

## Status

Accepted（2026-04-19，取代初版 standalone 决策）

## Context

AI 服务接入后路由快速增长，需要运行时动态调整模型权重/熔断/限流；APISIX Dashboard 已退役（2023），自建管理 UI 需要 Admin API；Coraza WAF 插件需动态配置规则。standalone 模式无 Admin API，无法满足。

## Decision

**APISIX 从 standalone 模式升级为 etcd（传统）模式。**

### 配置变化

| 项 | standalone（旧） | etcd（新） |
|----|----------------|-----------|
| 配置存储 | `apisix.yaml` 文件 | etcd |
| 配置方式 | 编辑 YAML + 热加载 | **Admin API（RESTful）** |
| 新增依赖 | 无 | etcd（开发单节点，生产 3 节点 HA） |

### AI Gateway 能力

| 插件 | 能力 |
|------|------|
| ai-proxy | 多 LLM 统一代理，协议转换 |
| ai-proxy-multi | 多模型负载均衡、故障转移、动态权重 |
| ai-rate-limiting | token 级限流（TPM 配额） |
| ai-rag | RAG 增强生成 |

### AI 场景特殊要求

| 传统假设 | AI 现实 | APISIX 支持 |
|----------|--------|:-----------:|
| 请求 < 1s | LLM 推理 30-120s | 超长 timeout |
| 响应一次返回 | LLM 流式输出（SSE） | 原生支持 |
| 限流按请求数 | LLM 按 token 计费 | ai-rate-limiting |
| 单一后端 | 多模型 Provider | ai-proxy-multi |
| 后端等价 | 模型成本差异 100x | 成本感知路由 |

### 部署

- 编排：`components/etcd/docker-compose.yml`
- plane: platform, role: config-store
- 管理 UI：`services/platform-admin/`（ADR-0013）

### Git 源 ↔ etcd 同步模型

**结构配置（Route/Upstream/SSL/Consumer/Service/GlobalRule、Coraza CRS 规则、全局 logger）走 GitOps，运行时策略（限流阈值、插件参数）走 UI 直改。**

| 配置类型 | 源真相 | 改法 | 审计 |
|----------|-------|------|------|
| 结构配置 | Git：`components/apisix/routes/*.yaml` | PR review → CI → `apisix-apply-routes.sh` → etcd | Git history |
| 运行时策略 | etcd | platform-admin UI → Admin API | platform-admin 操作日志 |

工具：
- `scripts/apisix-apply-routes.sh` — 解析 Git 源 YAML，幂等 PUT 到 Admin API（默认不清理 etcd 中 YAML 未声明的资源，避免误删 UI 改动）
- `scripts/apisix-export-routes.sh` — 从 etcd 反向导出为 YAML，供 diff 对比与回流 Git

回流闭环：UI 改动 → `export-routes.sh` → diff `routes/` → 人工确认 → PR → 合并 → CI apply。

为什么不直接用 standalone YAML：失去热加载 / 多节点协调 / AI Gateway 运行时 Admin API 调用能力，未来 K8s 迁移还得再改一次。

### K8s 备选

Higress（阿里，Apache 2.0）基于 Envoy，内置 Console UI + Gateway API 原生。当前 Docker Compose 阶段不迁移（需 Nacos，非 K8s 部署未经大规模验证）。**K8s 迁移时评估。**

## Consequences

- Admin API 可用，管理 UI 有标准 RESTful API
- AI Gateway 插件可运行时动态配置
- 新增 etcd 依赖
- 配置审计转移到 Admin API 操作日志 + platform-admin 记录
