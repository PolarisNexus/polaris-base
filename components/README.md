# components/ — 组件即包

每个子目录是一个**自包含的组件包**（compose + 配置 + 文档）。总入口 `deploy/docker-compose/docker-compose.yml` 通过 `include` 聚合所有组件。

## 组件清单

| 目录 | 组件 | 用途 |
|------|------|------|
| `postgres/` | PostgreSQL 16 | 核心业务数据库 |
| `redis/` | Redis 7 | 缓存、分布式锁 |
| `elastic/` | Elasticsearch 8.15 + Kibana | 业务全文检索 |
| `minio/` | MinIO | 统一对象存储（S3 协议） |
| `observability/` | Elasticsearch-otel + Kibana-otel | 可观测性数据存储与可视化 |
| `apisix/` | Apache APISIX | API 网关（Standalone 模式） |
| `casdoor/` | Casdoor | IAM — 用户目录、SSO、JWT |
| `safeline/` | SafeLine | WAF（暂缓接入，详见 ADR-0006） |

## 启动

```bash
cd deploy/docker-compose
docker compose up -d                          # 全量启动
docker compose up -d postgres redis           # 按需裁剪
```

## 镜像版本锁定

所有镜像**禁止 `latest`**，锁定具体版本。升级流程：本地单组件验证 → 更新 image tag → 提交 commit → CI/预发验证。

## 初始化职责

不做统一初始化容器，各组件用原生机制：

| 组件 | 方式 |
|------|------|
| PostgreSQL | `init/*.sh` 挂载到 `/docker-entrypoint-initdb.d`，首次启动幂等执行 |
| Elasticsearch | `elasticsearch-init` 一次性服务，配置 `kibana_system` 密码后退出 |
| APISIX | 声明式配置挂载即生效 |
| Casdoor | 应用自举（连 PG 后自动建表） |
