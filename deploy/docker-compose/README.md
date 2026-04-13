# deploy/docker-compose/ — Docker Compose 部署入口

## 一行启动

```bash
cd deploy/docker-compose
docker compose up -d
```

所有环境变量都有内置默认值，**零配置即可启动**。

## 自定义配置

各组件目录下有 `.env.example`：

```bash
cp ../../components/postgres/.env.example ../../components/postgres/.env
# 编辑 .env 覆盖默认值
```

## 按需启动

```bash
docker compose up -d postgres redis              # 仅数据层
docker compose up -d postgres redis apisix        # 数据层 + 网关
```

## 生成单文件分发

```bash
make release
# 生成 docker-compose.full.yml，可独立分发部署
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
| APISIX | 9080 / 9443 | 唯一公网入口 |
| Casdoor | 8000 | IAM 管理 UI |
| Kibana | 5601 | 业务 ES 管理 |
| Kibana-otel | 5602 | 观测 ES 管理 |
| MinIO Console | 9001 | 对象存储管理 |
| PG / Redis / ES | — | 仅内部访问 |

直连内部服务调试：

```bash
docker compose exec postgres psql -U polaris
docker compose exec redis redis-cli
```

## 数据持久化

卷命名为 `polaris-base_<component>_data`。`docker compose down` 不删卷；`docker compose down -v` 删除。
