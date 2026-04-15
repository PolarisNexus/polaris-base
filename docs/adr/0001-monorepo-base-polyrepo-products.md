# ADR-0001: Monorepo base + polyrepo products

## Status

Accepted

## Context

PolarisNexus 平台包含一个共享基座和多个业务产品。需要决定代码仓库的组织方式：全部 monorepo、全部 polyrepo、还是混合模式。

业务产品之间技术栈可能不同（Java、Python、前端），发布节奏各异，团队边界清晰。但它们共享相同的基础设施（PG、Redis、ES）、API 契约（Proto）和部署编排。

## Decision

采用**混合模式**：基座共享单一仓库 `polaris-base`，每个业务产品各自独立仓库。

**基座（`polaris-base`，单仓）**：
- API 契约（Proto）
- 顶层部署编排（`deploy/docker-compose/`）
- 第三方基础设施（APISIX、IAM、WAF、observability、PostgreSQL、Redis、Elasticsearch、MinIO）
- 自研共享服务（`services/` 下的邮件/会员/支付等）
- 平台文档与 ADR

基座内部通过 **Compose profile + label 表达 plane 分层**（data / platform / services，详见 ADR-0009），不引入仓库拆分。

**产品仓**：业务代码、产品级测试、产品级 CI/CD。

## Consequences

- 基座版本化可控，产品仓库引用基座的 Proto 生成桩代码
- 产品可独立发布，不受其他产品开发节奏影响
- 需要明确"产品如何注册到基座网关"的 PR 流程
- 跨仓库的变更（如 Proto 修改）需要协调多仓库更新
- 基座 plane 分层通过元数据承载，目录保持扁平，K8s 迁移时 label 可直接复用

## Alternatives Considered

- **全 monorepo**：所有代码在一个仓库。优点是原子提交，缺点是多语言 CI 复杂、仓库膨胀、团队权限难隔离。
- **全 polyrepo**：基座也拆成多个仓库（网关一个、IAM 一个，或 commons + data 双仓）。过于碎片化，配置关联性强的组件分开管理成本高；commons + data 双仓方案在 ADR-0009 详述并已否决。
