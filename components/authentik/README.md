# Authentik — IAM

平台 IAM（ADR-0010）。OIDC/OAuth2/SAML 颁发方 + 自带现代化 Admin UI。

## 目录

- `docker-compose.yml` — server + worker（共享 PG，外部依赖 `components/postgres/`）
- `blueprints/` — 声明式资源定义，worker 启动时自动 apply
  - `platform-admin.yaml` — platform-admin 的 OIDC Provider + Application（ADR-0013）

## 首次启动

```bash
make up-platform            # 启动 authentik-server + worker（依赖 postgres 先起）
docker compose logs authentik-worker  # 确认 blueprint 已被 apply（polaris-platform-admin）
```

初次启动 Authentik 会走 Initial Setup：浏览器访问 `http://localhost:9000/if/flow/initial-setup/` 创建管理员账号。完成后所有后续配置（Flows / Policies / Groups）在 Admin UI 进行。

## blueprint 维护

- 结构变更：修改 `blueprints/*.yaml` → 重启 worker 即生效（幂等）
- 运行时配置（用户/组/策略）：走 Authentik Admin UI，不要写回 blueprint，避免与运维 UI 操作冲突
- 详细语法：https://goauthentik.io/docs/developer-docs/blueprints

## 约定

- 业务代码不直接调用 Authentik 私有 API（ADR-0004）；通过标准 OIDC + Adapter 接入
- platform-admin 通过 blueprint 预置的 public client（`polaris-platform-admin`）做 PKCE 登录
