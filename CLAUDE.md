# polaris-base — AI 上下文

## 本层职责

polaris-base 是 PolarisNexus 平台的**共享基座**（单仓），存放：

- 跨服务 API 契约（gRPC Proto、未来的 OpenAPI 聚合）
- 部署编排（Docker Compose / K8s）
- 组件配置与编排（PG、Redis、ES、MinIO、APISIX、Casdoor、可观测性栈）
- 自研共享业务服务（`services/`，详见 ADR-0011）
- 产品矩阵索引（`products/`）
- 平台级文档与架构决策记录（`docs/adr/`）

**不存放产品业务代码。** 各产品在独立仓库中开发（详见 ADR-0001）。

## 关键约定

- **API First** — 先定义契约（Proto / OpenAPI），再实现代码
- **gRPC 集中管理** — 所有 `.proto` 文件统一放在 `api/proto/`
- **IAM 薄封装** — 业务代码不直接调用 Casdoor 私有 API，通过 Adapter/SPI 对接（ADR-0004；ADR-0010 重选型中）
- **不引入微服务全家桶** — 无 Spring Cloud / Dubbo，服务间通过 gRPC 直连
- **声明式网关配置** — APISIX Standalone 模式（YAML），不依赖 etcd（ADR-0002）
- **环境变量管理敏感信息** — 密码、密钥一律走环境变量 / Secret（ADR-0007）
- **一行启动** — `docker compose up -d` 零配置启动全部组件，默认 `.env` 激活所有 profile（ADR-0005）
- **组件即包** — 每个组件在 `components/<name>/` 下自包含 compose + 配置 + 文档，`components/` 扁平，不按 plane/role 分子目录
- **Plane / Role 走元数据** — 每个 compose 服务声明 `profiles: ["<plane>"]` + `labels.polaris.plane` + `labels.polaris.role`（ADR-0009）

## 禁止事项

- 禁止在此仓库放置产品业务源码（`services/` 下的共享业务除外）
- 禁止在配置文件中硬编码密码、密钥、连接串
- 禁止引入额外的服务发现组件（初期用 Docker Compose 直连 / K8s CoreDNS）
- 禁止引入消息队列（ADR-0003）
- 禁止使用 `latest` 镜像 tag，必须锁定具体版本
- 禁止按 plane/role 在 `components/` 下分子目录（plane/role 走 label，不走目录）

## 上下文指针

- 平台顶层技术栈：`docs/architecture/平台顶层技术栈.md`
- 部署编排入口：`deploy/docker-compose/docker-compose.yml`
- 默认 profile 激活：`deploy/docker-compose/.env`
- 网关路由配置：`components/apisix/apisix.yaml`
- 网关主配置：`components/apisix/config.yaml`
- 架构决策记录：`docs/adr/`（核心：ADR-0001、ADR-0005、ADR-0009、ADR-0011）
