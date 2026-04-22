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

- **`polaris-base` 创建 `polaris-net`**：顶层 compose 声明 `networks.polaris-net.name: polaris-net`，由 compose 自动创建
- **产品仓库加入 `polaris-net`**：声明 `external: true, name: polaris-net`
- 基座每个服务注册 `base-<service>` 别名（如 `base-postgres`、`base-apisix`）
- 各 compose 项目内部通过 `default` 网络短名通信

**启动链**（base → product）：基座必须先启动以创建网络；产品仓随后加入。基座内部 plane 分层走 Compose profile + label（见 ADR-0009）。

## Consequences

- 基座 `docker compose up -d` 一行启动，零前置步骤
- 产品仓库接入极简：声明 1 个 external 网络 + 每个服务 1 个别名
- 所有跨项目服务在同一网络，无网络层隔离
- 安全依赖认证（PG 密码、Redis AUTH）而非网络拓扑
- 迁移 k8s 时，NetworkPolicy 从 flat 网络 + 流量观测开始设计（业界最佳实践）

## Alternatives Considered

### 分层网络（polaris-data-net / polaris-app-net）

评估了按域划分的双网络方案：数据层服务（PG/Redis/ES/MinIO）走 `polaris-data-net`，应用层（APISIX/Authentik）走 `polaris-app-net`。

**否决理由**：

1. **配置膨胀**：典型产品的后端服务同时需要数据层（连 DB）和应用层（被网关路由），几乎每个服务都要挂两个共享网络 + 自身 default = 3 个网络
2. **安全价值有限**：Docker bridge 网络隔离是 L3/L4 层薄壳，不是安全基线。真正的安全在认证、最小权限、TLS，与网络拓扑无关
3. **扩展性差**：新组件（Kafka、Grafana、ML 推理）大多同时需要两个域，"归属哪个网络"的问题会反复出现
4. **K8s 映射精度低**：Docker 网络粒度（二元的在/不在）远粗于 k8s NetworkPolicy（per-pod + per-port），迁移时仍需从零设计策略
5. **业界无先例**：Dify、Supabase、GitLab、n8n 均不在 compose 层面做网络分段

### External 网络 + bootstrap 脚本

基座使用 `external: true` 网络，由前置脚本预创建。语义更纯粹但违反一行启动原则。

## 多机部署与当前架构的关系

当前 Compose 方案是**单机模型**——bridge 网络、named volume、`include` 聚合均不跨主机。这是有意选择：搭建期追求最低摩擦，多机需求出现时直接迁移 K8s，不经过 Swarm 等过渡方案。

### K8s 多机策略概述

K8s 集群由多个 Node（物理机/虚拟机）组成，调度器自动决定每个 Pod（≈容器）运行在哪台机器上：

- **按需调度**：默认由调度器根据各 Node 剩余资源自动分配。也可通过 nodeSelector / affinity 规则将特定组件钉到特定机器（如 ES 钉到大内存节点、GPU 服务钉到 GPU 节点）
- **弹性扩缩容**：HPA（Horizontal Pod Autoscaler）根据 CPU/内存/自定义指标自动增减 Pod 副本数。流量高峰时自动扩出更多副本分散到不同机器，低谷时缩回——粒度是单个服务，不是整机
- **有状态 vs 无状态**：无状态服务（APISIX、业务 API）天然支持多副本水平扩展；有状态服务（PG、ES、Redis）通过 StatefulSet + PVC 管理，扩容需要考虑数据分片/主从复制

### Compose → K8s 映射

| Compose 概念 | K8s 对应 | 说明 |
|-------------|---------|------|
| `polaris-net` + alias | Namespace + Service DNS | `base-postgres` → `base-postgres.polaris.svc.cluster.local` |
| named volume | PersistentVolumeClaim | 存储与 Pod 生命周期解耦，跨节点可用 |
| `include` 聚合 | Helm chart / Kustomize | 声明式编排，支持环境差异化 |
| healthcheck | livenessProbe / readinessProbe | 语义一致，语法不同 |
| `ports` 映射 | Service type / Ingress | 内部通信走 ClusterIP，外部入口走 Ingress |

### 迁移时机不变

当前 Compose 编排中的别名、卷、healthcheck、认证模式可 1:1 映射到 K8s 资源，无需为迁移预埋额外抽象。

## 未来升级触发条件

以下场景出现时，在 k8s 层面引入 NetworkPolicy：
- 平台迁移到 k8s（使用 Cilium Hubble 或 Calico Flow Logs 观测实际流量后生成策略）
- 合规审计要求网络隔离
- 多个独立团队同时开发产品，需要强制依赖边界
