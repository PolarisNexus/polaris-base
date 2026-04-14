# components/ — commons 第三方基础设施

每个子目录是一个**自包含的组件包**（compose + 配置 + 文档），由顶层 `deploy/docker-compose/docker-compose.yml` 通过 `include` 聚合。

**范围**：本目录只放第三方基础设施黑盒（我们编排不改代码）。自研共享服务见 `../services/`。数据持久化引擎（PG/Redis/ES/MinIO 等）在 `polaris-base-data` 仓。

## 组件清单

| 目录 | 组件 | 用途 |
|------|------|------|
| `apisix/` | Apache APISIX | API 网关（Standalone 模式，ADR-0002） |
| `casdoor/` | Casdoor | IAM — 用户目录、SSO、JWT（待 ADR-0010 重选型） |
| `observability/` | Elasticsearch-otel + Kibana-otel | OTel 数据存储与可视化 |
| `safeline/` | SafeLine | WAF（暂缓接入，ADR-0006） |

## 启动

```bash
# 仅 commons（独立）
cd deploy/docker-compose && docker compose up -d

# commons + data（组合，从仓库根）
make up
```

## 镜像版本锁定

所有镜像**禁止 `latest`**，锁定具体版本。

## 初始化职责

各组件用原生机制：

| 组件 | 方式 |
|------|------|
| Elasticsearch-otel | `elasticsearch-otel-init` 一次性服务，配置 `kibana_system` 密码后退出 |
| APISIX | 声明式配置挂载即生效 |
| Casdoor | 应用自举（连 `base-postgres` 后自动建表；PG 在 polaris-base-data 仓） |
