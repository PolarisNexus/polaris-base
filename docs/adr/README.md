# Architecture Decision Records

本目录记录 polaris-base 平台的关键架构决策，采用 [Michael Nygard](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions) 格式。

## 索引

| 编号 | 标题 | 状态 |
|------|------|------|
| [0001](0001-monorepo-base-polyrepo-products.md) | Monorepo base + polyrepo products | Accepted |
| [0002](0002-apisix-standalone-over-etcd.md) | APISIX etcd 模式 | Accepted |
| [0003](0003-no-message-queue-initially.md) | No message queue initially | Accepted |
| [0004](0004-iam-thin-adapter.md) | IAM thin adapter | Accepted |
| [0005](0005-multi-project-compose-model.md) | Multi-project compose model | Accepted |
| [0007](0007-secret-management-evolution.md) | Secret management evolution | Accepted |
| [0008](0008-ci-validation-strategy.md) | CI validation strategy | Accepted |
| [0009](0009-plane-split.md) | 单仓 + Profile + Label plane 表达 | Accepted |
| [0010](0010-iam-reselection.md) | IAM 重选型——Authentik | Accepted |
| [0011](0011-common-services-stack.md) | Commons 自研共享服务技术栈 | Accepted |
| [0012](0012-waf-coraza-in-apisix.md) | WAF——Coraza-in-APISIX | Accepted |
| [0013](0013-platform-admin-console.md) | platform-admin 统一管理控制台 | Accepted |
| [0014](0014-ai-gateway.md) | AI Gateway—— `ai-proxy-multi` + 按用户配额 | Accepted（Phase I MVP） |

## 新增 ADR

复制 [template.md](template.md)，编号递增，填写后提交 PR。
