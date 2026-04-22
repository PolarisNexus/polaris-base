# ADR-0013: platform-admin 统一管理控制台

## Status

Accepted（2026-04-19）

## Context

- **IAM**：Authentik 自带 Admin UI，无需自建
- **WAF**：Coraza 无管理 UI，需自建
- **Gateway**：APISIX Dashboard 已退役（2023），需自建

WAF + Gateway 管理合并为一个控制台。

## Decision

**新建 `services/platform-admin/` 作为 WAF + Gateway 统一管理控制台**（Go BFF + React + AntD Pro，ADR-0011 服务模式）。

### 架构

```
services/platform-admin/
├── cmd/server/main.go          # Go BFF
├── internal/
│   ├── waf/                    # Coraza 规则/日志管理
│   └── gateway/                # APISIX Admin API 调用
├── web/                        # React + AntD Pro
├── Dockerfile
└── docker-compose.yml
```

### 管理入口

| 入口 | 访问方式 |
|------|---------|
| IAM（Authentik Admin UI） | `https://<host>/authentik/` |
| WAF + Gateway（platform-admin） | `https://<host>/admin/` |

两入口统一通过 Authentik SSO 登录。

### 功能模块

**P1（已完成）**
- Gateway：路由/Upstream 只读浏览、插件配置 PATCH、限流 PATCH
- WAF：ES 攻击日志查询（`status ≥ 400` 口径）
- Bot：CrowdSec 决策 CRUD + 告警查询

**P1 明确不做**
- WAF 规则启停 UI / 路由例外 UI —— 走 `components/apisix/routes/95-coraza.yaml` Git 源 PR。理由：一键关 WAF 的运维风险远大于便利收益；CRS 规则调整频率低，PR review 链路更安全。UI 待 P2 结合权限/审计/双人复核再评估。

**P2**
- AI Gateway（`ai-proxy-multi` / token 配额）
- Turnstile & FingerprintJS Pro 联动配置
- WAF 规则 UI（如届时证明有需求）

**P3**
- Prometheus 面板、灰度发布、插件市场

### 技术要点

- Go BFF 调用 APISIX Admin API（Admin Key 认证）
- Coraza 规则通过 Admin API 更新插件配置
- 攻击日志从 ES 查询
- 前端 OIDC 对接 Authentik SSO
- plane: services, role: platform-admin

## Consequences

- WAF + Gateway 管理集中，避免多个独立前端
- IAM 管理与 WAF/Gateway 管理分离，通过 SSO 统一登录
