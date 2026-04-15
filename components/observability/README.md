# observability/ — 可观测性栈

**OpenTelemetry 是首日约束**——所有业务服务从第一行代码起必须接入 OTel SDK。

```
业务服务 ──OTel SDK──▶ OTel Collector ──▶ 后端存储/可视化
```

## 已接入

| 服务名 | polaris-net 别名 | 说明 |
|--------|-----------------|------|
| `otel-elasticsearch` | `base-otel-elasticsearch` | 日志 / APM 数据存储 |
| `otel-elasticsearch-init` | — | 一次性初始化（配置 kibana_system 密码） |
| `otel-kibana` | `base-otel-kibana` | 日志查询 / APM 可视化 |

## 待接入

| 服务名 | 说明 |
|--------|------|
| `otel-collector` | 统一遥测数据接收、处理、导出 |
| `otel-apm-server` | Elastic APM 数据接收 |
| `otel-prometheus` | 指标存储与告警 |
| `otel-grafana` | 指标仪表盘 |

## 命名规约

- 观测域组件一律 `otel-` 前缀（服务名、别名、卷名）
- 业务域 ES/Kibana 在 `components/elastic/`，裸名 `elasticsearch` / `kibana`

| 域 | ES | Kibana | 别名 |
|----|-----|--------|------|
| 业务 | `elasticsearch` | `kibana` | `base-elasticsearch` / `base-kibana` |
| 观测 | `otel-elasticsearch` | `otel-kibana` | `base-otel-elasticsearch` / `base-otel-kibana` |

## 约束

- 业务服务统一使用 OTel SDK，不得自选遥测方案
- OTel Collector 是唯一遥测出口，业务服务不直连后端存储
- 观测 ES 与业务 ES（`components/elastic/`）物理隔离
