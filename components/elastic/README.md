# elastic/ — 业务 Elastic 栈

服务于产品的全文搜索、向量检索等**业务功能**的 Elastic 组件。安全默认开启（`xpack.security.enabled=true`）。

> 可观测性场景（日志/APM/Traces）的 Elastic 栈放在 `components/observability/` 下，使用独立 ES 集群，避免日志洪峰影响业务搜索。

| 服务名 | polaris-net 别名 | 说明 |
|--------|-----------------|------|
| `elasticsearch` | `base-elasticsearch` | 业务全文检索 |
| `elasticsearch-init` | — | 一次性初始化（配置 kibana_system 密码） |
| `kibana` | `base-kibana` | ES 管理与可视化 |
