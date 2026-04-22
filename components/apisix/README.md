# APISIX — API 网关

平台唯一外部入口。etcd 模式（ADR-0002）；内嵌 Coraza WAF，行为层对接 CrowdSec（ADR-0012）。

## 目录

- `config.yaml` — APISIX 主配置（Admin API、etcd 连接、Wasm 插件注册）
- `routes/` — 结构配置 Git 源，PR review → CI apply（模型见 ADR-0002）
  - `90-access-log.yaml` — 访问日志 → ES（WAF 攻击日志查询源）
  - `95-coraza.yaml` — Coraza WAF 全局启用 + CRS 规则集

## Coraza WAF

- 二进制由 `apisix-coraza-init` 容器首次启动时下载（coraza-proxy-wasm v0.6.0，~18MB）至 `apisix_wasm` volume
- CRS 规则嵌入 wasm 自带，通过 `Include @owasp_crs/*.conf` 启用
- 启停 / 路由例外编辑 `routes/95-coraza.yaml` 走 PR（ADR-0013 P1 不做 UI 编辑器，P2 再评估）

## 日常命令

```bash
# Git 源 → etcd（幂等 PUT）
./scripts/apisix-apply-routes.sh

# etcd → YAML（UI 改动回流 Git 做 diff）
./scripts/apisix-export-routes.sh > /tmp/snap.yaml

# Admin API 直调
curl -H "X-API-KEY: ${APISIX_ADMIN_KEY}" http://localhost:9180/apisix/admin/routes
```
