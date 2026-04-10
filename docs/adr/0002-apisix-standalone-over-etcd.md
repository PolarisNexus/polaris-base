# ADR-0002: APISIX standalone over etcd

## Status

Accepted

## Context

Apache APISIX 支持三种部署模式：传统模式（依赖 etcd 集群）、standalone 模式（YAML 声明式）、混合模式。初期平台规模小（< 10 路由），需要在运维复杂度和动态能力之间取舍。

## Decision

初期使用 **standalone 模式**（`deployment.role: data_plane` + `config_provider: yaml`），路由配置以 YAML 文件挂载，APISIX 自动热加载。

## Consequences

- 零依赖：不需要部署和维护 etcd 集群（至少 3 节点）
- 配置即代码：路由变更走 git + PR 审核，有完整审计轨迹
- 热加载：修改 YAML 后 APISIX 自动重新加载，无需重启
- 缺点：不支持 Admin API 动态下发路由，不适合需要高频动态变更的场景
- 缺点：水平扩展多实例时需要共享配置文件（NFS / ConfigMap）

## Alternatives Considered

- **传统模式（etcd）**：适合大规模动态路由。初期 < 10 路由，etcd 集群的运维成本远超收益。
- **混合模式**：control plane + data plane 分离。适合多集群，初期单集群无此需求。
