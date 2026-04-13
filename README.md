# polaris-base

北极星枢纽的基座，一个多语言微服务（Polyglot Microservices）的底座平台。

## 仓库定位

polaris-base 是 PolarisNexus 平台的**共享基座仓库**，存放跨服务契约、部署编排和组件配置。业务代码在各产品独立仓库中开发。

## 目录结构

```
polaris-base/
├── api/                               ← 跨服务 API 契约（gRPC Proto）
├── components/                        ← 组件配置与编排（组件即包）
│   ├── postgres/                      ← 数据库
│   ├── redis/                         ← 缓存
│   ├── elastic/                       ← 业务 Elasticsearch + Kibana
│   ├── minio/                         ← 对象存储
│   ├── observability/                 ← 可观测性 Elasticsearch + Kibana
│   ├── apisix/                        ← API 网关
│   ├── casdoor/                       ← IAM 身份认证
│   └── safeline/                      ← WAF（暂缓接入）
├── deploy/                            ← 部署编排
│   ├── docker-compose/                ← Docker Compose 入口
│   └── kubernetes/                    ← K8s 部署（预留）
├── products/                          ← 产品矩阵索引
├── scripts/                           ← 运维脚本
└── docs/                              ← 架构文档与 ADR
```

## 快速开始

```bash
cd deploy/docker-compose
docker compose up -d
```

零配置即可启动。自定义参数见各组件目录下 `.env.example`。

## 架构分层

| 层 | 职责 | 核心组件 |
|---|---|---|
| 应用产品层 | 业务功能 | Spring Boot、Python AI、前端应用 |
| 公共基座层 | 网关 + IAM | APISIX、Casdoor |
| 可观测性层 | 遥测与监控 | OTel、Elasticsearch-otel、Kibana-otel |
| 基础设施层 | 存储 / 缓存 / 搜索 | PostgreSQL、Redis、Elasticsearch、MinIO |

## 相关文档

- [平台顶层技术栈](docs/architecture/平台顶层技术栈.md)
- [架构决策记录](docs/adr/)
