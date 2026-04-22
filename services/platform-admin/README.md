# platform-admin — WAF + Gateway 统一管理控制台

> 决策见 ADR-0013。Go BFF + React + AntD。

IAM 管理走 Authentik 自带 Admin UI，本服务不涉及。

## 后端依赖

| 依赖 | 用途 | 凭据 |
|------|------|------|
| APISIX Admin API `:9180` | 运行时策略 CRUD（结构配置走 Git，ADR-0002） | `APISIX_ADMIN_KEY` → X-API-KEY |
| Elasticsearch `:9200` | Coraza 攻击日志查询（index `apisix-access`） | `ELASTIC_USERNAME/PASSWORD` → basic auth |
| CrowdSec LAPI `:8080` | 决策读 + 告警读 / 写 + 决策删除 | machine：`CROWDSEC_USERNAME/PASSWORD` → JWT<br>bouncer：`CROWDSEC_BOUNCER_KEY` → X-Api-Key |
| Authentik `:9000` | OIDC SSO | public client + PKCE，无 secret |

> CrowdSec 的双凭据不是冗余：LAPI 按角色区分权限 —— `machines` 写入 alerts，`bouncers` 只读 decisions。platform-admin 同时扮演两种角色，故两份凭据都需要。

## 目录结构

```
services/platform-admin/
├── cmd/server/main.go          后端入口
├── internal/                   业务实现
│   ├── config/                 环境变量加载
│   ├── gateway/                APISIX Admin API 封装
│   ├── waf/                    Coraza 攻击日志查询（ES）
│   └── bot/                    CrowdSec 决策 / 告警
├── web/                        React + AntD 前端
├── Dockerfile                  Go 多阶段镜像
└── docker-compose.yml          聚合进 deploy/docker-compose
```

生成的 gRPC/Connect Go 桩位于仓库根 `api/gen/go/`（独立 Go module，由 `go.work` 聚合）。

## 开发

```bash
# 1. 一键启动全栈（含 Authentik + APISIX + ES + CrowdSec + platform-admin 本身）
docker compose -f deploy/docker-compose/docker-compose.yml up -d

# 2. 首次使用 Authentik：浏览器访问 http://localhost:9000/if/flow/initial-setup/
#    创建管理员账号（Authentik blueprint 已自动装好 polaris-platform-admin OIDC Provider）

# 3. apply APISIX 结构路由到 etcd（含 Coraza global_rule + elasticsearch-logger）
bash scripts/apisix-apply-routes.sh

# 4. 前端开发服务器（:5173，Vite 代理 /polaris.* 到 BFF :8080）
cd services/platform-admin/web && npm install && npm run dev

# 5. 浏览器登录：http://localhost:5173 → 跳 Authentik → 回调后进入仪表盘

# 6. 重新生成 proto 桩（仅在修改 api/proto/polaris/platform_admin/**.proto 后）
docker run --rm -v "$PWD/api:/work" -w /work/proto bufbuild/buf:1.47.2 generate
```

若需 `go run` 热迭代后端（不走镜像），先 `docker compose stop platform-admin`，再用 `services/platform-admin/.env.example` 对齐的环境变量启动。

## 认证（ADR-0013 SSO）

- Authentik blueprint：`components/authentik/blueprints/platform-admin.yaml` 自动创建 OAuth2/OIDC Provider（public client + PKCE）和 Application，slug `platform-admin`，client_id `polaris-platform-admin`
- 前端：PKCE authorization code flow，`/auth/callback` 处理回调，ID token 存 `localStorage`
- BFF：从 `.well-known/openid-configuration` 拉 JWKS，RS256 本地校验，未通过返回 401
- 跳过路径：`/healthz`
- 生产部署需把前端同源部署到 APISIX 后（避免 CORS 暴露 token）

## 开发分期（见 ADR-0013）

- **P1（已完成）** — SSO 登录；Gateway 路由只读；WAF 攻击日志（ES `apisix-access`，按 `response.status >= 400` 口径）；Bot 决策（bouncer）+ 告警（machine）；IAM 菜单外链 Authentik Admin
- **P1 明确不做** — WAF 规则启停 UI；走 `components/apisix/routes/95-coraza.yaml` Git PR
- **P2** — AI Gateway（`ai-proxy-multi` / token 配额）+ Turnstile & FingerprintJS Pro 联动；APISIX CrowdSec bouncer 生效路径；WAF 规则 UI（如届时证明有需求）
- **P3** — Prometheus 面板、灰度发布、插件市场

结构配置（路由/Upstream/SSL、Coraza CRS、全局 logger）在 UI 中只读，编辑引导到 Git PR 流程。

## 约束

- API First：BFF API 定义先于前端（`api/proto/polaris/platform_admin/v1/`）
- 密钥走环境变量 / Secret（见 `.env.example`）
- 只经 Admin API 与 APISIX 交互，禁止直写 etcd
