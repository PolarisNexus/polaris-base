# routes/ — APISIX 结构配置 Git 源

> ADR-0002 "Git 源 ↔ etcd 同步模型"的 Git 源侧。日常命令见 `components/apisix/README.md`。

## 范围

**结构配置**（需 PR review）：路由 / Upstream / SSL / Service / Consumer / GlobalRule、Coraza WAF CRS 规则启停与路由例外（`95-coraza.yaml`）、全局 logger（`90-access-log.yaml`）。

**不在此**（UI 直改 etcd）：限流阈值、AI Gateway 权重、插件运行时参数。

## 文件命名

`NN-<name>.yaml`，`NN` 两位数前缀可控排序：

```
00-health.yaml
10-auth-login.yaml
20-api-orders.yaml
```

## 格式

顶层字段即 Admin API 资源类型（复数）；`id` 必填：

```yaml
routes:
  - id: orders
    uri: /api/orders/*
    upstream:
      type: roundrobin
      nodes:
        "base-orders:8080": 1

upstreams:
  - id: shared-orders-upstream
    type: roundrobin
    nodes:
      "base-orders:8080": 1
```

apply 脚本用 `PUT /{resource}/{id}` 幂等写入，不清理 UI 改动。
