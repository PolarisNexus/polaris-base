# polaris-base-commons — AI 上下文

> 仓库目录名保留 `polaris-base`；逻辑职责已按 ADR-0009 拆分为 commons，远端仓库重命名待手动操作。

## 本层职责

polaris-base-commons 是 PolarisNexus 平台的**公共能力层**，存放：

- 跨服务 API 契约（gRPC Proto）
- 部署编排顶层入口（Docker Compose / K8s）
- 第三方基础设施组件：APISIX、IAM（Casdoor，待 ADR-0010 重选型）、safeline（WAF）、observability（OTel ES+Kibana）
- 自研共享业务服务：`services/` 下的 Go + React 模块（邮件/会员/支付等，见 ADR-0011）
- 产品矩阵索引（products/）
- 平台级文档与架构决策记录（docs/adr/，主 ADR 仓）

**配套数据持久化底座：`polaris-base-data`**（PG/Redis/ES/MinIO 等，另一仓库）。

## 关键约定

- **API First** — 先定义契约（Proto），再实现代码
- **gRPC 集中管理** — 所有 `.proto` 统一放在 `api/proto/`，按服务域名建子目录
- **IAM 薄封装** — 业务代码不直接调 IAM 私有 API，走 Adapter/SPI（ADR-0004）
- **不引入微服务全家桶** — 无 Spring Cloud / Dubbo，服务间 gRPC 直连
- **声明式网关** — APISIX Standalone 模式（ADR-0002）
- **环境变量敏感信息** — 密码密钥走 env/Secret，禁硬编码（ADR-0007）
- **网络模型** — 本仓创建 `polaris-net`，data 仓和产品仓以 `external` 加入（ADR-0005、ADR-0009）
- **一行启动** — 本仓 `docker compose up -d` 独立启动；`make up` 组合启动 commons + data
- **组件即包** — `components/<name>/` 自包含 compose + env + README
- **自研服务技术栈** — Go 后端 + React/AntD Pro 前端 + monorepo（ADR-0011）

## 禁止事项

- 禁止在本仓放置产品业务代码（但 `services/` 下**跨产品共享**的通用能力可放）
- 禁止在本仓放置有状态数据引擎（归 data 仓）
- 禁止硬编码密码、密钥、连接串
- 禁止额外服务发现组件（用 Docker Compose / K8s CoreDNS）
- 禁止消息队列在本仓（ADR-0003；未来 Kafka 等归 data 仓）
- 禁止 `latest` 镜像 tag

## 上下文指针

- 顶层 compose：`deploy/docker-compose/docker-compose.yml`
- 组合启动：`Makefile`
- 网关路由：`components/apisix/apisix.yaml`
- 网关主配置：`components/apisix/config.yaml`
- 自研服务约定：`services/README.md`
- ADR：`docs/adr/`（本仓是 ADR 主仓）
- 数据仓：`../polaris-base-data/`
