# observability/ — 可观测性栈

平台可观测性基础设施。**OpenTelemetry 是首日约束**——所有业务服务从第一行代码起必须接入 OTel SDK，统一采集 Traces / Metrics / Logs。

## 架构

```
业务服务 ──OTel SDK──▶ OTel Collector ──▶ 后端存储/可视化
```

## 组件

| 子目录 | 服务名 | polaris-net 别名 | 状态 |
|--------|--------|-----------------|------|
| `elasticsearch/` | `elasticsearch-otel` | `base-elasticsearch-otel` | 已接入 |
| `elasticsearch/` | `elasticsearch-otel-init` | — | 已接入（初始化 kibana_system 密码） |
| `kibana/` | `kibana-otel` | `base-kibana-otel` | 已接入 |
| `otel-collector/` | `otel-collector` | `base-otel-collector` | 待接入 |
| `apm-server/` | `apm-server` | `base-apm-server` | 待接入 |
| `prometheus/` | `prometheus` | `base-prometheus` | 待规划 |
| `grafana/` | `grafana` | `base-grafana` | 待规划 |

## 命名规约

与业务 Elastic 栈（`components/elastic/`）区分：

| 域 | ES 服务名 | Kibana 服务名 | polaris-net 别名 |
|----|-----------|---------------|-----------------|
| 业务 | `elasticsearch` | `kibana` | `base-elasticsearch` / `base-kibana` |
| 观测 | `elasticsearch-otel` | `kibana-otel` | `base-elasticsearch-otel` / `base-kibana-otel` |

`docker ps` 一眼可辨归属。

## 凭证变量

业务 ES 与观测 ES 使用独立变量，互不干扰，Compose 变量插值保证 init 与对应 Kibana 自动同步：

| 变量 | 默认值 | 用途 |
|------|--------|------|
| `ELASTIC_OTEL_PASSWORD` | `changeme` | 观测 ES `elastic` 超级用户密码 |
| `KIBANA_OTEL_PASSWORD` | `changeme` | 观测 ES `kibana_system` 用户密码 |
| `ES_OTEL_JAVA_OPTS` | `-Xms512m -Xmx512m` | 观测 ES JVM 配置 |
| `KIBANA_OTEL_PORT` | `5602` | 观测 Kibana 宿主机端口 |

## 约束

- 业务服务**不得**自行选择遥测方案，统一使用 OpenTelemetry SDK
- OTel Collector 是唯一的遥测数据出口，业务服务不直连后端存储
- 日志 ES 集群与业务 ES 集群（`components/elastic/`）物理隔离，独立生命周期
