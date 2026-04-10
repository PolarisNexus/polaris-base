# ADR-0004: IAM thin adapter over Casdoor

## Status

Accepted

## Context

Casdoor 作为 IAM 组件提供用户管理、SSO、JWT 颁发。业务代码如果直接调用 Casdoor 的私有 REST API，会在业务逻辑和 IAM 实现之间产生强耦合。未来如果切换 IAM 方案（如 Keycloak、Logto），所有调用点都需要修改。

## Decision

业务代码**不直接调用 Casdoor 私有 API**。通过一层 Adapter/SPI 接口对接，业务侧只依赖抽象接口。

## Consequences

- 业务代码与 IAM 实现解耦，切换 IAM 方案只需实现新 Adapter
- Adapter 层可统一处理 token 刷新、用户信息缓存、权限映射
- 增加一层间接性，初期可能感觉"过度设计"
- JWT 校验仍在网关层（APISIX jwt-auth 插件）完成，Adapter 主要处理用户管理 API

## Alternatives Considered

- **直接调用 Casdoor API**：最快但耦合最紧。一旦 Casdoor API 变更或需要替换，改动面大。
- **完全自建 IAM**：控制力最强但投入最大，不符合"用开源组件快速搭建"的平台策略。
