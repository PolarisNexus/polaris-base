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

## 产品注册到基座网关

新产品接入基座网关的 PR 流程：

1. **添加网关路由**：在 `components/apisix/apisix.yaml` 添加产品的路由规则
2. **登记产品信息**：在上方产品清单表格中添加一行
3. **配置网络**：在产品仓库的 `docker-compose.yml` 中声明外部网络：
   ```yaml
   networks:
     polaris-net:
       external: true
       name: polaris-base_polaris-net
   ```
4. **访问基座服务**：通过别名 `polaris-base-<service>` 访问（如 `polaris-base-postgres`）
5. **提交 PR**：路由变更和产品登记在同一个 PR 中提交，经审核后合并
