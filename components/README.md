# components/ — 组件即包（Self-contained Component Unit）

每个子目录是一个**自包含的组件包**，包含该组件的：
- `docker-compose.yml` — 编排定义（可独立 `docker compose up -d`）
- `.env.example` — 环境变量清单与默认值
- 配置文件、init 脚本、README

总入口 `deploy/docker-compose/docker-compose.yml` 通过 `include` 指令聚合所有组件，实现一键启动。

## 子目录

| 目录 | 组件 | 用途 |
|------|------|------|
| `apisix/` | Apache APISIX | API 网关 — 路由、JWT 校验、限流熔断 |
| `casdoor/` | Casdoor | IAM — 用户目录、SSO、JWT 颁发 |
| `safeline/` | SafeLine | WAF — 恶意流量拦截（暂未纳入编排，详见 ADR-0006） |
| `postgres/` | PostgreSQL 16 | 核心业务数据库 |
| `redis/` | Redis 7 | 缓存、分布式锁、简单异步任务 |
| `elastic/elasticsearch/` | Elasticsearch 8.15 | 全文检索、日志聚合、初期向量检索 |
| `minio/` | MinIO | 统一对象存储（S3 协议） |
| `observability/` | 可观测性栈 | 占位，未来接入 OTel / Prometheus / Grafana |

## 快速启动

```bash
cd deploy/docker-compose
docker compose up -d       # 零配置，所有变量都有内置默认值
```

按需裁剪：`COMPOSE_PROFILES=data docker compose up -d` 仅启动数据层。

## 镜像版本锁定策略

所有镜像**禁止使用 `latest`**，必须锁定到具体版本或日期 tag，避免同一份 compose 在不同时间拉到不同镜像。

| 组件 | 当前版本 | tag 风格 | 升级窗口 |
|------|---------|---------|---------|
| PostgreSQL | `postgres:16.6` | `<major>.<minor>` | 跟随上游 minor 发布，评审后升 |
| Redis | `redis:7.4.1` | `<major>.<minor>.<patch>` | 同上 |
| Elasticsearch | `docker.elastic.co/elasticsearch/elasticsearch:8.15.0` | `<major>.<minor>.<patch>` | 避免跨 minor 升级 |
| MinIO | `minio/minio:RELEASE.2024-11-07T00-52-20Z` | `RELEASE.<日期>` | 季度评审 |
| APISIX | `apache/apisix:3.11.0-debian` | `<x>.<y>.<z>-debian` | 跟随上游 |
| Casdoor | `casbin/casdoor:v1.843.0` | `v<x>.<y>.<z>` | 跟随上游 |

### 升级流程

1. 本地单组件验证：`cd components/<name> && docker compose up -d`，跑健康检查
2. 更新对应 `components/<component>/docker-compose.yml` 的 image tag
3. 提交 commit，commit message 注明旧版本 → 新版本和主要变更
4. 在 CI/预发环境验证后再进生产

### 安全基线（按环境分层）

当前默认值是**本地开发基线**：
- `ES_SECURITY_ENABLED=false`
- 所有组件端口直接暴露到宿主机
- 默认密码 `changeme`

上更高环境（staging/prod）时**必须**通过 `docker-compose.override.yml` 覆盖：
- ES 开启 `xpack.security.enabled=true`，通过专用初始化容器生成证书和重置密码
- 除 APISIX 网关外，所有组件端口**只绑 `127.0.0.1`**，统一从网关入
- 所有密码走密钥管理（Vault / K8s Secret），不落盘在 `.env`

详见 ADR-0007（密钥管理演进路径）。

### 初始化职责分配

**不做"统一初始化容器"**——每个组件的初始化用其原生机制：

| 组件 | 初始化方式 |
|------|----------|
| PostgreSQL | `components/postgres/init/*.sh` 挂载到 `/docker-entrypoint-initdb.d`，首次启动幂等执行 |
| Elasticsearch | 启用安全时单独加一个**一次性 init 服务**（同 ES 镜像 + `restart: "no"`），完成后退出 |
| APISIX | 声明式配置挂载即生效，无需初始化 |
| Casdoor | 应用自举（连 PG 后自动建表） |
