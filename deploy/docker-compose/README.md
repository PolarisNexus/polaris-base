# deploy/docker-compose/ — Docker Compose 部署入口

## 一行启动

```bash
cd deploy/docker-compose
docker compose up -d
```

所有环境变量都有内置默认值，**零配置即可启动**。默认 `.env` 激活 `data + platform + services` 全部 profile（ADR-0005、ADR-0009）。

## 自定义配置

各组件目录下有 `.env.example`：

```bash
cp ../../components/postgres/.env.example ../../components/postgres/.env
# 编辑 .env 覆盖默认值
```

## 按需启动（按 plane）

通过 `COMPOSE_PROFILES` 选择 plane（shell 环境变量优先于 `.env` 文件）：

```bash
COMPOSE_PROFILES=data docker compose up -d            # 仅 data plane
COMPOSE_PROFILES=platform docker compose up -d        # 仅 platform plane
COMPOSE_PROFILES=data,platform docker compose up -d   # data + platform
```

或通过仓库根 Makefile：

```bash
make up-data / up-platform / up-services
```

## 视图过滤（按 plane / role）

`docker compose ps` 仅支持 status 过滤；plane/role 走 `docker ps` label 过滤：

```bash
docker ps --filter label=com.docker.compose.project=polaris-base \
          --filter label=polaris.plane=data

docker ps --filter label=com.docker.compose.project=polaris-base \
          --filter label=polaris.role=cache

# Makefile 便利
make ps-data / ps-platform / ps-services
```

## 跨项目网络模型

基座启动后自动创建共享网络 `polaris-net`。

**产品仓库接入**：

```yaml
# polaris-alpha/docker-compose.yml
services:
  alpha-api:
    environment:
      DB_HOST: base-postgres
    networks:
      default: {}
      polaris-net:
        aliases:
          - alpha-api

networks:
  polaris-net:
    external: true
    name: polaris-net
```

**别名规则**：基座服务注册 `base-<service>` 别名，产品服务注册 `<project>-<service>` 别名。

## 端口策略

仅网关和管理 UI 映射到宿主机，其他服务内部访问：

| 服务 | 端口 | 说明 |
|------|------|------|
| APISIX | 9080 / 9443 / 9180 | 公网入口 + Admin API |
| Authentik | 9000 | IAM 管理 UI |
| Kibana | 5601 | 业务 ES 管理 |
| otel-Kibana | 5602 | 观测 ES 管理 |
| MinIO Console | 9001 | 对象存储管理 |
| PG / Redis / ES | — | 仅内部访问 |

直连内部服务调试：

```bash
docker compose exec postgres psql -U polaris
docker compose exec redis redis-cli
```

## 数据持久化

卷命名为 `polaris-base_<component>_data`（由 `name: polaris-base` 前缀生成）。`docker compose down` 不删卷；`docker compose down -v` 删除。
