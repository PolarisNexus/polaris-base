# APISIX — API 网关

平台唯一外部入口，接管所有 HTTP/gRPC 流量。

## 运行模式

Standalone 模式（无 etcd），声明式 YAML 配置（详见 ADR-0002）：
- `config.yaml` — APISIX 主配置（角色、配置源）
- `apisix.yaml` — 路由与上游声明

## 职责

JWT 校验、限流熔断、跨域处理、Header 转换（`X-User-ID` / `X-Tenant-ID`）、日志埋点。
