# infra/postgres/ — PostgreSQL 配置

PostgreSQL 16 作为平台核心业务数据库。

## 预期内容

- `postgresql.conf` 调优参数
- 初始化 SQL 脚本（建库、建角色）
- 多租户数据隔离策略配置

## 注意

- 代码层屏蔽 SQL 方言，支持未来切换至达梦、人大金仓等国产数据库
- 密码通过环境变量注入，不硬编码
