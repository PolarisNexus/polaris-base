# Casdoor — IAM 身份认证

全局用户目录、SSO 单点登录、JWT 令牌颁发。

- 管理 UI 映射到宿主机（默认 8000）
- 业务代码不直接调用 Casdoor 私有 API，统一通过 Adapter/SPI 薄封装层对接（详见 ADR-0004）
- 使用 PostgreSQL 作为后端存储，通过 `POLARIS_EXTRA_DBS` 自动创建 `casdoor` 数据库
