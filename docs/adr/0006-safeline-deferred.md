# ADR-0006: SafeLine deferred

## Status

Accepted

## Context

SafeLine（雷池 WAF）是计划中的 Web Application Firewall 组件。评估后发现其部署复杂度远超其他组件：

- 需要 7 个容器（tengine、detector、mario、bridge、management、log-processor、postgres）
- 要求 host 网络模式或独立子网，与基座网络模型冲突
- 自带独立 PostgreSQL 实例，与基座 PG 需要隔离管理
- 管理界面需要额外的端口和安全配置

## Decision

SafeLine **暂不纳入基座编排**。`components/safeline/` 仅保留 README 占位文件，说明延后理由和未来接入条件。

## Consequences

- 基座编排保持简洁，6 个核心组件均已健康运行
- 需要独立部署 SafeLine（直接用官方 docker-compose 或物理机部署）
- 网关层（APISIX）暂时缺少 WAF 防护，需通过其他手段弥补（如 APISIX 内置的基础安全插件）

## Alternatives Considered

- **立即纳入编排**：技术上可行，但会显著增加编排复杂度（+7 容器），且 host 网络模式与 compose 网络模型不兼容，需要额外的网络桥接方案。
- **用 APISIX 安全插件替代**：APISIX 有 `ip-restriction`、`ua-restriction`、`cors` 等插件，但不具备 SafeLine 的深度语义分析能力（SQL 注入、XSS 检测等）。长期仍需专业 WAF。
