# polaris-base — AI 上下文

## 本层职责

polaris-base 是 PolarisNexus 平台的**共享基座**（单仓），存放：

- 跨服务 API 契约（gRPC Proto、未来的 OpenAPI 聚合）
- 部署编排（Docker Compose / K8s）
- 组件配置与编排（PG、Redis、ES、MinIO、APISIX+Coraza、etcd、CrowdSec、Authentik、可观测性栈）
- 自研共享业务服务（`services/`，详见 ADR-0011）
- 产品矩阵索引（`products/`）
- 平台级文档与架构决策记录（`docs/adr/`）

**不存放产品业务代码。** 各产品在独立仓库中开发（详见 ADR-0001）。

## 关键约定

- **API First** — 先定义契约（Proto / OpenAPI），再实现代码
- **gRPC 集中管理** — 所有 `.proto` 文件统一放在 `api/proto/`
- **IAM 薄封装** — 业务代码不直接调用 Authentik 私有 API，通过标准 OIDC + Adapter/SPI 对接（ADR-0004、ADR-0010）
- **不引入微服务全家桶** — 无 Spring Cloud / Dubbo，服务间通过 gRPC 直连
- **网关 etcd 模式** — APISIX 传统模式（etcd 存储），通过 Admin API 管理路由和插件（ADR-0002）
- **路由 Git 源 ↔ etcd 同步** — 结构配置（Route/Upstream/SSL、Coraza CRS 调整、全局 logger）走 `components/apisix/routes/*.yaml` + CI apply；限流阈值、插件运行时参数走 platform-admin UI 直改 etcd（ADR-0002、ADR-0013）
- **WAF 分层** — 请求层 Coraza（APISIX Wasm 插件）+ 行为层 CrowdSec（独立容器）+ 人机验证 Turnstile SaaS + 指纹 FingerprintJS Pro SaaS（ADR-0012）
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
- 网关主配置：`components/apisix/config.yaml`
- 路由 Git 源：`components/apisix/routes/*.yaml`
- IAM 编排：`components/authentik/docker-compose.yml`
- WAF 行为层：`components/crowdsec/docker-compose.yml`
- 管理控制台指引：`services/platform-admin/README.md`
- 架构决策记录：`docs/adr/`（核心：ADR-0001、ADR-0002、ADR-0005、ADR-0009、ADR-0010、ADR-0011、ADR-0012、ADR-0013）
