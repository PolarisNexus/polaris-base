# ADR-0011: Commons 自研共享服务技术栈

## Status

Accepted

## Context

ADR-0009 拆分后，`polaris-base-commons` 承载两类内容：

- `components/`：第三方基础设施黑盒（APISIX、IAM、OTel 等）
- `services/`：**自研**跨产品共享业务模块（邮件、会员、支付、未来的工作流引擎包装等）

自研服务需要确定：

1. 后端语言/框架
2. 前端语言/框架（管理后台 UI）
3. 服务组织方式（monorepo vs polyrepo）
4. 新服务创建流程

## Decision

### 目录结构

```
polaris-base-commons/
├── api/proto/                      # gRPC 契约（跨所有服务）
├── components/                     # 第三方基础设施
└── services/                       # 自研共享服务（monorepo）
    ├── README.md                   # 服务清单 + 开发约定
    ├── _template/                  # 新服务脚手架
    │   ├── cmd/server/main.go
    │   ├── internal/
    │   ├── web/                    # React + AntD Pro 前端
    │   ├── Dockerfile
    │   ├── docker-compose.yml
    │   └── README.md
    ├── membership/                 # （未来）会员服务
    ├── email/                      # （未来）邮件服务
    └── payment/                    # （未来）支付服务
```

### 后端：Go

**理由**：
- gRPC 生态最佳（CLAUDE.md 约定"服务间通过 gRPC 直连"）
- K8s 友好：秒级启动、小镜像（multi-stage 构建后通常 20-30MB）
- IAM 候选（Zitadel、Ory）同为 Go，语言/工具链一致
- 中国支付生态（支付宝、微信支付）有官方 Go SDK
- 学习曲线可控（简单语法、强标准库）

**工具链**：
- 模块管理：Go modules
- gRPC：`grpc-go` + `buf` + `protoc-gen-go` + `protoc-gen-go-grpc`
- HTTP 网关（按需）：`grpc-gateway` 或 `connect-go`
- 日志：`slog`（标准库）
- 观测：OpenTelemetry Go SDK
- 测试：标准 `testing` + `testify`

### 前端：TypeScript + React + Ant Design Pro

服务前端均为**管理/配置型后台 UI**（会员管理、支付配置、邮件模板编辑），非终端用户界面。

**理由**：
- Ant Design Pro 是后台脚手架业界标配（表格/表单/权限/国际化开箱即用）
- 中文生态最强，团队上手快
- 与未来 IAM 管理控制台风格一致

**工具链**：
- 构建：Vite
- 路由/状态：React Router + Zustand 或 React Query
- 组件：Ant Design + Ant Design Pro Components
- API 对接：基于 proto 生成 TS 客户端（`protoc-gen-ts` 或 `connect-es`）

### 服务组织：Monorepo

所有自研服务平铺在 `polaris-base-commons/services/`：

**理由**：
- 共享 proto 定义、CI 模板、构建脚本
- 跨服务改动单 PR 完成（如 proto 契约演进）
- 服务数量 ≤ 10 前仓库规模可控

**退出条件**：某服务膨胀为独立业务域（团队、发布节奏、技术栈脱离平台约束）→ 拆为独立 `polaris-service-<name>` 仓库。

### 新服务流程

1. 从 `services/_template/` 复制到 `services/<new-service>/`
2. 在 `api/proto/<new-service>/v1/` 定义 gRPC 契约
3. 生成桩代码（`buf generate`）
4. 实现业务逻辑
5. 在顶层 `deploy/docker-compose/docker-compose.yml` 添加 include
6. 在 APISIX 路由配置注册对外路径
7. 更新 `services/README.md` 服务清单

## Consequences

- 技术栈统一降低维护成本，新服务启动速度快
- proto 先行：接口契约在代码之前定稿，符合 CLAUDE.md "API First"
- 前端脚手架一次性投入，后续服务复用
- 当某服务业务复杂度超过 commons 承载范围时，按"退出条件"拆仓，成本可控

## Alternatives Considered

### 后端：Java Spring Boot

- 企业级生态最成熟、支付 SDK 最全
- 否决理由：JVM 启动慢、内存占用大，与 K8s 秒级调度、Go 为主的平台语言栈不契合

### 后端：TypeScript (Node)

- 与前端栈统一
- 否决理由：gRPC 支持不如 Go 流畅；高并发场景性能劣势

### 前端：Vue 3 + Element Plus

- 与 React + AntD Pro 对等，中文生态同样强
- 选型差异不大，最终以 AntD Pro 后台脚手架生态更深为决定因素

### 每服务独立仓库

- 独立权限/版本/CI
- 否决理由：初期服务数少、共享资产多；monorepo 效率高，拆仓按需触发
