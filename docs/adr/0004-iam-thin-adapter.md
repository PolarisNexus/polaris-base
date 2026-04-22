# ADR-0004: IAM thin adapter

## Status

Accepted

## Context

IAM 组件（当前 Authentik，ADR-0010）提供用户管理、SSO、JWT 颁发。业务代码如果直接调用 IAM 私有 API，会产生强耦合，切换实现时改动面大。

## Decision

业务代码**不直接调用 IAM 私有 API**。通过一层 Adapter/SPI 接口对接，业务侧只依赖标准 OIDC 协议和抽象接口。

## Consequences

- 业务代码与 IAM 实现解耦，切换方案只需实现新 Adapter
- Adapter 层可统一处理 token 刷新、用户信息缓存、权限映射
- JWT 校验在网关层（APISIX jwt-auth 插件）完成，Adapter 主要处理用户管理 API
