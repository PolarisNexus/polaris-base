# products/ — 产品矩阵索引

记录接入平台基座的所有产品信息。本目录不存放代码，仅维护产品元信息表。

新产品接入时，在下表登记。

## 产品清单

| 产品名 | 仓库 | 技术栈 | 负责人 | 状态 | 说明 |
|--------|------|--------|--------|------|------|
| —      | —    | —      | —      | —    | —    |

## 接入要求

- 在各自仓库根目录维护 `CLAUDE.md` 和 `README.md`
- 对外 API 遵循 OpenAPI 3.x 规范
- 内部通信使用 gRPC，Proto 定义提交至 `api/proto/`
- 通过 APISIX 网关统一暴露外部接口

## 启动链

平台必启基座（ADR-0009）：

```
polaris-base  →  <product>
  (创建 polaris-net)   (加入网络)
```

产品启动前确保基座已 up：`make up` 或 `docker compose -f deploy/docker-compose/docker-compose.yml up -d`。

## 产品注册到基座网关

1. **添加网关路由**：新增 `components/apisix/routes/NN-<product>.yaml`（Git 源，ADR-0002），PR 合入后 CI 跑 `scripts/apisix-apply-routes.sh` 写入 etcd；运行时插件/限流调整走 platform-admin UI
2. **登记产品信息**：在上方产品清单表格中添加一行
3. **配置网络**：产品仓的 `docker-compose.yml` 声明外部网络：
   ```yaml
   networks:
     polaris-net:
       external: true
       name: polaris-net
   ```
4. **访问基座服务**：通过别名 `base-<service>` 接入
   - 网关 / IAM：`base-apisix`、`base-authentik`
   - 数据：`base-postgres`、`base-redis`、`base-elasticsearch`、`base-minio`
   - 观测：`base-otel-elasticsearch`、`base-otel-kibana`
5. **提交 PR**：产品登记提交 PR
