# infra/casdoor/ — Casdoor IAM 配置

Casdoor 负责全局用户目录、SSO 单点登录和 JWT 令牌颁发。

## 预期内容

- 应用（Application）和组织（Organization）初始化配置
- 多租户 Realm 配置
- RBAC 角色与权限模板

## 注意

业务代码不直接调用 Casdoor 私有 API，统一通过 Adapter/SPI 薄封装层对接。
