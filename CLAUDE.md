# polaris-base — AI 上下文

## 本层职责

polaris-base 是 PolarisNexus 平台的**共享基座仓库**，只存放：

- 跨服务 API 契约（gRPC Proto、未来的 OpenAPI 聚合）
- 部署编排（Docker Compose / K8s）
- 组件配置与编排（APISIX、Casdoor、SafeLine、PG、Redis、ES、MinIO）
- 产品矩阵索引（products/）
- 平台级文档

**不存放任何业务代码。** 各产品/服务在独立仓库中开发。

## 关键约定

- **API First** — 先定义契约（Proto / OpenAPI），再实现代码
- **gRPC 集中管理** — 所有 `.proto` 文件统一放在 `api/proto/`，按服务域名建子目录，各服务仓库引用生成桩代码
- **IAM 薄封装** — 业务代码不直接调用 Casdoor 私有 API，通过 Adapter/SPI 对接，便于未来切换
- **不引入微服务全家桶** — 无 Spring Cloud / Dubbo，服务间通过 gRPC 直连
- **声明式网关配置** — APISIX 初期使用 Standalone 模式（YAML），不依赖 etcd
- **环境变量管理敏感信息** — 密码、密钥一律走环境变量 / Secret，不硬编码

## 禁止事项

- 禁止在此仓库放置业务服务源码
- 禁止在配置文件中硬编码密码、密钥、连接串
- 禁止引入额外的服务发现组件（初期用 Docker Compose 直连 / K8s CoreDNS）
- 禁止引入消息队列（初期用 Redis / 数据库轮询满足异步需求）

## 上下文指针

- 平台顶层技术栈：`docs/architecture/平台顶层技术栈.md`
- 部署编排入口：`deploy/docker-compose/docker-compose.yml`
- 网关路由配置：`components/apisix/apisix.yaml`
- 网关主配置：`components/apisix/config.yaml`
- 架构决策记录：`docs/adr/`
