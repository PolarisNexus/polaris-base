# ADR-0009: 单仓 + Profile + Label 的 plane 表达

## Status

Accepted（2026-04-15，取代 2026-04-14 的"双仓拆分"决策）

## Context

ADR-0001 采用"单基座 + 多产品"模型，基座仓库为 `polaris-base`。随组件扩展，单仓出现以下压力：

- **容器数量**：当前 9 个（业务 ES 栈 3 + OTel ES 栈 3 + PG/Redis/MinIO/APISIX/Casdoor 共 6 非 ES 容器），规划中图 DB/向量 DB/Kafka/Flink 等将推向 20+
- **认知压迫**：业务 ES 与观测 ES 等同组件不同职责共存
- **演进解耦需求**：数据层新增与网关/IAM/OTel 演进节奏不同
- **K8s 迁移对齐**：按 plane 分 Helm chart / namespace 的规划需要仓库/编排结构提前对齐
- **Commons 扩展**：除基础设施外，还将承载跨产品共享业务能力（邮件/会员/支付）

### 决策演进

- **v1–v3**：单仓 `polaris-base`，组件在 `components/` 扁平存放
- **v4（2026-04-14 一度实施并提交）**：拆分为 `polaris-base-commons` + `polaris-base-data` 双仓
- **v5（本 ADR，2026-04-15）**：撤回 v4，回到单仓 + 扁平 `components/` + Compose profile/label

v4 实施当日评估为不妥：双仓引入跨 compose `depends_on` 失效、`external: true` 网络语义、两套 CI、版本兼容矩阵等固有成本；而上述压力（容器数量、认知隔离、独立演进、K8s 对齐）**不需要通过"结构"（仓库/目录）表达分类，用"元数据"（profile + label）更成熟优雅**。

### 业界参照

Supabase、Dify、Sentry self-hosted、Grafana LGTM Stack、Temporal self-hosted 均为**单仓 + 单 compose + profile/label 模型**。目录按组件扁平，plane/role 通过 compose profile 与标签表达。

## Decision

**单仓 `polaris-base`；`components/` 扁平；plane 通过 `profiles:` 激活；plane/role 通过 `labels:` 过滤视图。**

### Plane 与 Role 归属

| 组件 | plane | role |
|------|-------|------|
| postgres | data | relational-db |
| redis | data | cache |
| elastic（elasticsearch + init + kibana） | data | search |
| minio | data | object-storage |
| apisix | platform | gateway |
| casdoor | platform | iam |
| observability（otel-elasticsearch + init + otel-kibana） | platform | observability |
| safeline（占位，ADR-0006 暂缓） | platform | waf |

每个 compose 服务声明：

```yaml
profiles: ["<plane>"]
labels:
  polaris.plane: <plane>
  polaris.role: <role>
```

### Compose 项目

- 单一 `name: polaris-base`、单一 `deploy/docker-compose/docker-compose.yml`（`include` 聚合所有组件）
- `deploy/docker-compose/.env` 默认 `COMPOSE_PROFILES=data,platform,services` 保证 `docker compose up -d` 一行启动
- 网络 `polaris-net` 由本 compose 创建；产品仓通过 `external: true, name: polaris-net` 接入

### 按需启动 / 视图过滤

```bash
# 按 plane 选择性启动
COMPOSE_PROFILES=data docker compose up -d

# 按 plane 过滤视图（docker compose ps 仅支持 status 过滤，故走 docker ps）
docker ps --filter label=com.docker.compose.project=polaris-base --filter label=polaris.plane=data

# 按 role 过滤视图
docker ps --filter label=com.docker.compose.project=polaris-base --filter label=polaris.role=cache
```

### 便利层

Makefile：`make up` / `up-data` / `up-platform` / `up-services` / `ps-data` / `ps-platform` / `ps-services`。

### OTel 命名规约

观测域组件一律 `otel-` 前缀（服务名、别名、卷名）：`otel-elasticsearch`、`otel-elasticsearch-init`、`otel-kibana`；别名 `base-otel-elasticsearch` / `base-otel-kibana`；卷 `otel_elasticsearch_data`；环境变量 `OTEL_ELASTIC_PASSWORD` / `OTEL_KIBANA_PASSWORD` / `OTEL_ES_JAVA_OPTS` / `OTEL_KIBANA_PORT`。`elasticsearch` 不缩写（对齐 Elastic 官方、Bitnami、Helm 图表惯例）。

### observability 归属

observability 归 `platform`——当前仅 OTel ES+Kibana，轻量，与网关/IAM 并列为横切平台能力。未来扩到 Prometheus/Grafana/Jaeger/Loki 再考虑独立 plane。

### 类型嵌套（gateway/apisix/、iam/casdoor/）

**不加**。业界扁平为主（Supabase、Dify、Sentry、Grafana 均扁平）；角色用 label 表达，同一角色出现多实现时再考虑嵌套。顶层 `components/README.md` 维护角色→实现映射表作为"类型视角"入口。

## Consequences

- `docker compose up -d` 一行启动不变
- Casdoor 同 compose 直连 `postgres` 短名 + `depends_on` 恢复，无跨仓重试负担
- plane/role 语义以元数据承载，目录结构与仓库边界保持简单
- 未来 K8s 迁移时 label 可直接复用为 `app.kubernetes.io/component` / `polaris.plane` 等标签，`namespace` 按 plane 切分无需重新发明分类
- 取消 v4 的仓库/网络/版本矩阵 overhead

## Alternatives Considered

### 双仓拆分（v4，已否决）

`polaris-base-commons` + `polaris-base-data` 双仓。**问题**：跨 compose `depends_on` 失效（Casdoor ↔ PG）、`external: true` 网络耦合、两仓 CI/版本矩阵成本、与业界单仓模型背离。v4 当日回退。

### 按 plane 分子目录（`components/data/`、`components/platform/`）

结构上清晰，但与 `profiles + labels` 方案重复表达同一信息；且增加路径深度、include 语义噪音，目录重命名牵扯面大。业界主流是扁平。

### 按 role 分子目录（`components/gateway/apisix/`、`components/iam/casdoor/`）

同一 role 多实现时才有价值。当前 1:1 映射，直接用 `polaris.role` label 更轻。

### 延迟决策

v4 已证明"结构性拆分"是错误方向；推迟只会累积更多结构性债务。单仓 + profile/label 是终态，不需要再等。

## 回退路径

单仓内全回退：删除每个 compose 的 `profiles` 与 `labels`，合并默认 `.env`。所有组件仍可独立启动。
