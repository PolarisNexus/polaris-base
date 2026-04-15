# ADR-0008: CI validation strategy

## Status

Accepted

## Context

polaris-base 作为共享基座，其配置变更（compose 文件、网关路由、Proto 定义）会影响所有产品。需要 CI 验证保证变更不会破坏基座完整性。但当前处于平台搭建初期，尚无跨仓库联动场景，CI 投入的 ROI 不高。

## Decision

**本轮先不实现 CI workflow**，但明确未来 CI 需要覆盖的检查项：

### 静态检查（每次 PR 触发）
- `docker compose config`：验证 compose 文件语法正确、include 路径存在、变量替换无误
- `buf lint`：Proto 文件规范检查
- `buf breaking`：Proto 向后兼容性检查
- `.editorconfig` 检查：文件格式一致性

### 集成检查（合并到 main 后触发）
- `docker compose up -d` + healthcheck：验证所有服务能正常启动
- 网关 → Casdoor 链路验通：`curl /casdoor/api/health`
- 跨项目别名可达性：验证 `base-postgres` 等 `base-<service>` 别名解析

### 触发条件
- 首个跨仓库联动场景出现时，启动 CI 建设
- 优先实现静态检查（成本低、收益高），再逐步加集成检查

## Consequences

- 当前依赖人工验证（`docker compose config` + `make up` + 手动检查）
- 未来 CI 建设有明确 checklist，不会遗漏关键检查项
- 延迟建设 CI 的风险：手动操作可能遗漏验证步骤

## Alternatives Considered

- **立即建 CI**：保证最高质量，但当前只有一个开发者、变更频率低，CI 维护成本（runner、镜像缓存、flaky test）与收益不匹配。
- **只做静态检查不做集成检查**：成本最低，但 compose 语法正确不代表能启动成功（如镜像不存在、端口冲突）。
