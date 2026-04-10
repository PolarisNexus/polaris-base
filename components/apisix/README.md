# infra/apisix/ — APISIX 网关配置

Apache APISIX 是平台唯一的外部入口，接管所有 HTTP/gRPC 流量。

## 运行模式

初期使用 **Standalone 模式**（DB-less），路由配置以声明式 YAML 维护：

- `config.yaml` — APISIX 声明式路由配置（`apisix.yaml` 格式）

未来可迁移至 etcd 动态路由模式。

## 职责

- JWT 校验（离线公钥 / 实时穿透双模式）
- 限流熔断（`limit-req` / `api-breaker` 插件）
- 跨域处理
- Header 转换（提取 Token Payload 注入 `X-User-ID`、`X-Tenant-ID`）
- 日志埋点
