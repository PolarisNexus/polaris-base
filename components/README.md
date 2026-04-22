# components/ — 组件包目录

每个子目录是一个**自包含的组件包**（compose + 配置 + 文档），由顶层 `deploy/docker-compose/docker-compose.yml` 通过 `include` 聚合。

**范围**：第三方基础设施黑盒（我们编排不改代码）。自研共享服务见 `../services/`。

## 扁平结构 + plane/role 标签

目录按组件扁平存放，**不按 plane/role 分子目录**。plane/role 由每个 compose 服务的 `labels` 和 `profiles` 表达（ADR-0009）。

每个服务必须声明：

```yaml
services:
  <svc>:
    profiles: ["<plane>"]        # data | platform | services
    labels:
      polaris.plane: <plane>
      polaris.role: <role>
```

## 组件 → role/plane 映射

| 目录 | 组件 | plane | role |
|------|------|-------|------|
| `postgres/` | PostgreSQL | data | relational-db |
| `redis/` | Redis | data | cache |
| `elastic/` | Elasticsearch + Kibana（业务域） | data | search |
| `minio/` | MinIO | data | object-storage |
| `etcd/` | etcd（APISIX 配置存储） | platform | config-store |
| `apisix/` | Apache APISIX（etcd 模式，ADR-0002）+ 内嵌 Coraza WAF（ADR-0012） | platform | gateway |
| `crowdsec/` | CrowdSec 行为层 Bot/恶意 IP 检测（ADR-0012） | platform | waf-bot |
| `authentik/` | Authentik（IAM，ADR-0010） | platform | iam |
| `observability/` | OTel Elasticsearch + OTel Kibana | platform | observability |

## 按 role 查询

```bash
# 当前 cache 实现是哪个？
docker ps --filter label=com.docker.compose.project=polaris-base --filter label=polaris.role=cache

# 当前 platform 全景？
docker ps --filter label=com.docker.compose.project=polaris-base --filter label=polaris.plane=platform

# 或直接用 Makefile 便利
make ps-data / ps-platform / ps-services
```

## 启动

```bash
# 全量（默认 .env 激活 data + platform + services）
docker compose -f deploy/docker-compose/docker-compose.yml up -d

# 按 plane 选择性启动
COMPOSE_PROFILES=data docker compose -f deploy/docker-compose/docker-compose.yml up -d

# Makefile 便利
make up / up-data / up-platform / up-services
make ps / ps-data / ps-platform / ps-services
```

## 镜像版本锁定

所有镜像**禁止 `latest`**，锁定具体版本。

## 初始化职责

各组件用原生机制：

| 组件 | 方式 |
|------|------|
| elasticsearch / otel-elasticsearch | `*-init` 一次性服务，配置 `kibana_system` 密码后退出 |
| APISIX | etcd 模式，结构配置走 `routes/*.yaml` Git 源 → `apisix-apply-routes.sh` 写 etcd |
| CrowdSec | 启动时自动装 collections（env `COLLECTIONS`），从 apisix 容器 stdout 读日志 |
| Authentik | 应用自举（连 `postgres` 后自动迁移；同 compose 项目，直连短名） |
| etcd | 数据持久化到卷，单节点开发 / 3 节点生产 HA |

## 命名规约

- 业务域 ES/Kibana：裸名 `elasticsearch` / `kibana`，别名 `base-elasticsearch` / `base-kibana`
- 观测域 ES/Kibana：一律 `otel-` 前缀——`otel-elasticsearch` / `otel-kibana`，别名 `base-otel-elasticsearch` / `base-otel-kibana`，卷 `otel_elasticsearch_data`
- 对外别名统一 `base-<service>` 前缀，产品仓通过别名接入
