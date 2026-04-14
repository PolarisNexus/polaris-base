# ADR-0009: 基座一分为二——polaris-base-commons + polaris-base-data

## Status

Accepted

## Context

ADR-0001 采用"单基座 + 多产品"模型，基座仓库为 `polaris-base`。随着组件扩展，单仓出现以下压力：

- **容器数量**：当前 11 个（业务 ES 栈 3 + OTel ES 栈 3 + PG/Redis/MinIO/APISIX/Casdoor 共 5）；规划中的大数据组件（图 DB、向量 DB、Kafka、Flink 等）将使总数达 20+
- **认知压迫**：同仓库下"业务 ES + OTel ES"等同组件不同职责共存
- **演进解耦需求**：数据层新增（加 Milvus/Neo4j）与网关/IAM/OTel 演进节奏不同
- **K8s 迁移**：近期规划按 plane 分 Helm chart / namespace，仓库结构应提前对齐
- **Commons 扩展**：除基础设施外，还将承载跨产品共享的业务能力（邮件、会员、支付等登录用户相关通用模块）

## Decision

基座一分为二，两仓均使用 `polaris-base-` 前缀以表达"平台底座、必须启动"，与可选产品仓形成视觉区隔：

| 仓库 | 职责 | 组件 |
|------|------|------|
| **`polaris-base-commons`** | 公共能力层（基础设施 + 通用业务模块） | APISIX、IAM、safeline、observability（OTel ES+Kibana）、`services/` 下自研共享服务 |
| **`polaris-base-data`** | 数据持久化底座 | PostgreSQL、Redis、Elasticsearch（业务）、MinIO；未来 Neo4j/Milvus/Kafka 等 |

**命名语义**：
- `commons` = 公共共享能力（Apache Commons 语感），承载行为/能力
- `data` = 持久化状态，承载状态

**别名前缀不变**：两仓内部服务仍注册 `base-<service>`，产品引用零改动。

## Consequences

### 网络

- `polaris-net` 由 **commons 创建**（`networks.polaris-net.name: polaris-net`）
- data 仓以 `external: true, name: polaris-net` 加入
- 产品仓同 data 仓一样 external 加入（与原来完全一致）

### 启动链

```
polaris-base-commons  →  polaris-base-data  →  <product>
     (创建网络)           (加入网络)            (加入网络)
```

commons 仓提供 `Makefile` 便利层（`make up` 依次启动两仓），两仓各自仍支持独立 `docker compose up -d`。

### 跨仓依赖

- IAM（Casdoor/后继者）连 PG：跨 compose `depends_on` 失效，改走应用层重试（主流 IAM 皆支持）
- 应用连数据：通过 `base-postgres` 等别名，无感知

### 版本协调

- 两仓独立打 tag；commons 主仓维护兼容矩阵文档
- 产品引用可锁定 `(commons@vX, data@vY)` 组合

### ADR 与文档

- ADR 集中维护在 commons，data 仓 README 提供指针
- `products/` 产品注册表留 commons
- `api/proto/` 跨服务契约留 commons

### observability 归属

observability（OTel ES + Kibana）留 **commons**——它是"平台观测能力"，不是"被观测的数据存储"。与 APISIX/IAM/WAF 并列为横切基础设施。

## Alternatives Considered

### 单仓 + Compose profiles / 目录分组

profiles 仅减少运行容器数，目录/配置仍共存；未解决认知隔离、独立演进、独立 K8s chart 等核心诉求。

### 单仓 + 延迟拆分

ADR 早期草稿建议"触发条件到达再拆"。由于 K8s 迁移已是近期规划、commons 层即将纳入自研业务服务（邮件/会员/支付），触发条件已命中，延迟拆分只会增加后续迁移成本。

### 三仓（额外拆 observability）

observability 当前仅 3 个容器（OTel ES + init + Kibana），独立成仓收益小。未来 Prometheus/Grafana/Jaeger/Loki 全量引入后可重新评估。

### 基座归为"platform" / "edge" / "core"

- `platform` 过泛且与"polaris 本身是平台"冲突
- `edge` 不覆盖 OTel 观测能力
- `core` 与未来"核心业务代码"语义混淆
- `commons` 最贴合："既有基础能力也有业务能力的共享层"

## 回退路径

两仓合并机制简单：将 `polaris-base-data/components/*` 移回 commons，合并顶层 compose include 即可。
