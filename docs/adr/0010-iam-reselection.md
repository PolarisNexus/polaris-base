# ADR-0010: IAM 重选型——Authentik 接替 Casdoor

## Status

Accepted（2026-04-19）

## Context

Casdoor 存在"假开源"问题：商业菜单强制展示、关键功能付费墙、开源版定位为试用级别。ADR-0004 的 Adapter 抽象已就位，切换成本可控。

## Decision

**选定 Authentik 替代 Casdoor。** 复用 Authentik 自带 Admin UI，不自建 IAM 管理界面。

### 选型理由

| 维度 | Authentik |
|------|----------|
| 许可证 | MIT |
| 开源纯度 | 99% commits 在开源版；原则：永不将已有功能移入付费版 |
| Admin UI | 现代 SPA，干净无商业品牌 |
| 认证流 | **Flows 可视化拖拽编排**（杀手特性） |
| 协议 | OIDC/OAuth2/SAML/LDAP/SCIM/RADIUS |
| 多租户 | Tenant 原生支持 |
| 用户管理 | 完整（CRUD/组/会话/审计/自助密码重置，全部免费） |
| 部署 | server + worker + PG（2025.10 已移除 Redis 依赖） |

### 集成方式

```
APISIX ──jwt-auth 插件──→ Authentik JWKS endpoint（标准 OIDC Discovery）
业务代码 ──ADR-0004 Adapter──→ Authentik API（标准 OIDC/SCIM）
```

### AI 场景覆盖

| AI 需求 | Authentik 覆盖方式 |
|--------|------------------|
| API Key 管理 | Token + Application |
| 租户 token 配额 | 需网关配合（APISIX `ai-rate-limiting`） |
| 模型级权限 | Groups + Policies |
| Agent 工作流服务间认证 | Outpost + Service Account |

### 部署规划

- 编排位置：`components/authentik/docker-compose.yml`
- plane: platform, role: iam
- PG 复用基座实例（`POLARIS_EXTRA_DBS` 新增 `authentik`）
- 通过 APISIX 反代 `/authentik/` 路径暴露

## Consequences

- ADR-0004 Adapter 抽象价值兑现
- IAM 管理 UI 零自建成本
- platform-admin（ADR-0013）通过 OIDC 对接 Authentik SSO 统一登录
- 降级后备：Keycloak（Apache 2.0，功能最全，JVM 资源重）
