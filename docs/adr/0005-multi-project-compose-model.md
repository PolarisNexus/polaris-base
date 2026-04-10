# ADR-0005: Multi-project compose model — 单一共享网络 `polaris-net`

## Status

Accepted

## Context

PolarisNexus 的部署模型是"基座 + 多个独立产品仓库"。每个产品仓库有自己的 `docker-compose.yml`。需要决定：

1. 基座和产品如何在 Docker 网络层互通
2. 如何保证 `docker compose up -d` 一行启动（无前置脚本）
3. 是否在 Docker Compose 层面做网络分段（如按域划分 data-net / app-net）

## Decision

采用**单一共享网络 `polaris-net`**，不在 Docker Compose 层面做分层网络。

- 基座顶层 compose 声明 `networks.polaris-net.name: polaris-net`，由 compose 自动创建
- 基座每个服务注册 `base-<service>` 别名（如 `base-postgres`）
- 产品仓库声明 `external: true, name: polaris-net` 加入共享网络
- 各 compose 项目内部通过 `default` 网络短名通信

## Consequences

- 基座 `docker compose up -d` 一行启动，零前置步骤
- 产品仓库接入极简：声明 1 个 external 网络 + 每个服务 1 个别名
- 所有跨项目服务在同一网络，无网络层隔离
- 安全依赖认证（PG 密码、Redis AUTH）而非网络拓扑
- 迁移 k8s 时，NetworkPolicy 从 flat 网络 + 流量观测开始设计（业界最佳实践）

## Alternatives Considered

### 分层网络（polaris-data-net / polaris-app-net）

评估了按域划分的双网络方案：数据层服务（PG/Redis/ES/MinIO）走 `polaris-data-net`，应用层（APISIX/Casdoor）走 `polaris-app-net`。

**否决理由**：

1. **配置膨胀**：典型产品的后端服务同时需要数据层（连 DB）和应用层（被网关路由），几乎每个服务都要挂两个共享网络 + 自身 default = 3 个网络
2. **安全价值有限**：Docker bridge 网络隔离是 L3/L4 层薄壳，不是安全基线。真正的安全在认证、最小权限、TLS，与网络拓扑无关
3. **扩展性差**：新组件（Kafka、Grafana、ML 推理）大多同时需要两个域，"归属哪个网络"的问题会反复出现
4. **K8s 映射精度低**：Docker 网络粒度（二元的在/不在）远粗于 k8s NetworkPolicy（per-pod + per-port），迁移时仍需从零设计策略
5. **业界无先例**：Dify、Supabase、GitLab、n8n 均不在 compose 层面做网络分段

### External 网络 + bootstrap 脚本

基座使用 `external: true` 网络，由前置脚本预创建。语义更纯粹但违反一行启动原则。

## 未来升级触发条件

以下场景出现时，在 k8s 层面引入 NetworkPolicy：
- 平台迁移到 k8s（使用 Cilium Hubble 或 Calico Flow Logs 观测实际流量后生成策略）
- 合规审计要求网络隔离
- 多个独立团队同时开发产品，需要强制依赖边界
