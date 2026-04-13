# PostgreSQL 16 — 核心业务数据库

- 代码层屏蔽 SQL 方言，支持未来切换国产数据库
- `init/01-create-databases.sh` 通过 `POLARIS_EXTRA_DBS` 自动创建额外数据库
- 密码通过环境变量注入，不硬编码
