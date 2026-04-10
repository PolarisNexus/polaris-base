# ADR-0005: Multi-project compose model

## Status

Accepted

## Context

PolarisNexus 的部署模型是"基座 + 多个独立产品仓库"。每个产品仓库有自己的 `docker-compose.yml`。需要决定：

1. 基座和产品如何在 Docker 网络层互通
2. 如何保证 `docker compose up -d` 一行启动（无前置脚本）

## Decision

- 基座使用 compose 默认网络机制，声明 `networks.default.name: polaris-base_polaris-net`
- 基座每个服务注册跨项目稳定别名 `polaris-base-<service>`（如 `polaris-base-postgres`）
- **跨项目复杂度落在产品侧**：产品仓库声明 `external: true, name: polaris-base_polaris-net` 来加入基座网络
- 基座侧零前置步骤，`docker compose up -d` 直接可用
- 数据卷使用 compose 命名卷（非 external），由 compose 自动创建

## Consequences

- 基座启动体验与 Dify/Supabase 对齐：一行命令、零配置
- 产品仓库需先确保基座已启动（网络需先存在），这是合理的依赖方向
- 别名 `polaris-base-<service>` 与 compose 项目名解耦，项目名变化不影响跨项目解析
- 对应 k8s 迁移：别名 ≈ Service DNS、命名卷 ≈ PVC、网络 ≈ Namespace + NetworkPolicy
- 缺点：卷名带 compose 自动前缀（`polaris-base_postgres_data`），不如 external 卷名简洁

## Alternatives Considered

- **External 网络 + external 卷 + bootstrap 脚本**：语义更纯粹（名字不带项目前缀），但需要前置脚本创建资源，违反一行启动原则。
- **单一 compose 项目**：基座和产品都在同一个 compose 项目里。随着产品增多，单项目会膨胀，且产品无法独立启停、独立 CI。
- **Docker Compose `extends`**：产品 compose 继承基座 compose 的服务定义。语义复杂，且 extends 不支持跨文件 volumes/networks 继承。
