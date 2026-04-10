# ADR-0007: Secret management evolution

## Status

Accepted

## Context

平台密钥（数据库密码、API Key、JWT Secret）的管理需要一个演进路径。初期用 `.env` 文件足够，但上线后需要更安全的方案。需要在简单性和安全性之间找到适合每个阶段的平衡。

## Decision

采用分阶段演进策略：

### 阶段 1：`.env` 文件（当前）
- 所有密码通过 `${VAR:-default}` 内置默认值，开发环境零配置
- `.env` 文件在 `.gitignore` 中，不入版本库
- `.env.example` 展示变量清单

### 阶段 2：SOPS 加密（上预发/生产前）
- 使用 [mozilla/sops](https://github.com/getsops/sops) 加密 `.env` 文件
- 加密后的文件可安全提交到版本库
- 解密密钥通过 KMS / age key 管理

### 阶段 3：Vault / K8s Secret（生产环境）
- HashiCorp Vault 或 K8s Secret 作为密钥源
- 应用通过 sidecar 或 CSI driver 获取密钥
- 密钥不落盘，运行时注入

## Consequences

- 每个阶段的方案都是前一阶段的超集，迁移路径清晰
- 开发环境始终保持简单（`docker compose up -d` 无需配置密钥）
- 需要在升级到下一阶段前，文档化迁移步骤

## Alternatives Considered

- **初期就上 Vault**：安全性最高，但运维复杂度与当前阶段不匹配。Vault 本身需要高可用部署和 unseal 管理。
- **Docker Secrets**：仅限 Swarm 模式，与 compose standalone 不兼容。
- **只用 `.env` 到底**：开发环境可以，生产环境不可接受。密码明文存储、无审计、无轮换机制。
