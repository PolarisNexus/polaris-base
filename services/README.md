# services/ — commons 自研共享服务

本目录存放 `polaris-base-commons` 的**自研**跨产品共享业务模块。与 `components/`（第三方基础设施黑盒）互补。

## 技术栈（见 ADR-0011）

- **后端**：Go（gRPC + buf 工具链）
- **前端**：TypeScript + React + Ant Design Pro（管理后台 UI）
- **组织方式**：monorepo，服务数量 ≤ 10 前不拆仓
- **proto 契约**：统一放在仓库根 `api/proto/<service>/v1/`

## 服务清单

| 服务 | 路径 | 职责 | 状态 |
|------|------|------|------|
| — | — | — | 待立项 |

> 首个候选：会员（membership）或邮件（email）

## 新增服务流程

1. 从 `_template/` 复制到 `services/<new-service>/`
2. 在仓库根 `api/proto/<new-service>/v1/` 定义 gRPC 契约
3. 运行 `buf generate` 生成 Go / TS 桩代码
4. 实现业务逻辑
5. 在 `deploy/docker-compose/docker-compose.yml` 添加 service include
6. 在 `components/apisix/` 路由配置注册对外路径
7. 更新本清单表

## 约束

- **API First**：proto 先行，代码后写
- **IAM 解耦**：不直接调用 Casdoor/后继者私有 API，走 ADR-0004 Adapter
- **observability**：统一使用 OTel SDK（ADR 观测性约束）
- **密钥**：不硬编码，一律走环境变量 / Secret（ADR-0007）

## 退出条件

当某服务满足以下**任一**时，拆出为独立 `polaris-service-<name>` 仓库：

- 团队专职化（独立 owner）
- 发布节奏脱离 commons（要求独立 tag/版本）
- 技术栈超出 commons 约束（引入非 Go/非 React 栈）
