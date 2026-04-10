# deploy/docker-compose/ — Docker Compose 部署入口

## 一行启动

```bash
cd deploy/docker-compose
docker compose up -d
```

所有环境变量都有内置默认值，**零配置即可启动**。

## 自定义配置

各组件的 `.env.example` 在对应目录下：

```bash
# 示例：修改 PG 密码
cp ../../components/postgres/.env.example ../../components/postgres/.env
# 编辑 .env 修改 POSTGRES_PASSWORD
```

## 按需启动

不想启动全部组件时，指定服务名：

```bash
docker compose up -d postgres redis              # 仅数据层
docker compose up -d postgres redis apisix        # 数据层 + 网关
```

## 生成单文件分发

```bash
make -C ../../scripts release
# 生成 docker-compose.full.yml，可独立分发部署
```

## 跨项目网络模型

基座启动后会自动创建共享网络 `polaris-net`。

**产品仓库接入基座**：在产品仓库的 `docker-compose.yml` 中声明外部网络：

```yaml
# polaris-alpha/docker-compose.yml
services:
  alpha-api:
    environment:
      DB_HOST: base-postgres            # 基座服务别名
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

**服务别名规则**：基座服务在 `polaris-net` 上注册 `base-<service>` 别名（如 `base-postgres`、`base-redis`）。产品服务注册 `<project>-<service>` 别名（如 `alpha-api`）。

**注意**：产品仓库需要在基座 `docker compose up -d` 之后才能启动（网络需先存在）。

## 生产覆盖

```bash
cp docker-compose.override.yml.prod.example docker-compose.override.yml
docker compose up -d
```

生产覆盖会：开启 ES 安全、要求强密码等。

## 端口策略

**只有 APISIX 网关暴露端口到宿主机**（9080/9443），其他所有服务（PG、Redis、ES、MinIO、Casdoor）均不映射端口。所有访问统一从网关入。

需要直连内部服务调试时：

```bash
docker compose exec postgres psql -U polaris        # 连 PG
docker compose exec redis redis-cli                  # 连 Redis
```

## 数据持久化

卷命名为 `polaris-base_<component>_data`（如 `polaris-base_postgres_data`），由 compose 自动创建。`docker compose down` 不删卷；`docker compose down -v` 会删除。
