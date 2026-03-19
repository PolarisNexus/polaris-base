# polaris-base

北极星枢纽的基座，一个多语言微服务（Polyglot Microservices）的底座平台。

## 仓库定位

polaris-base 是 PolarisNexus 平台的**共享基座仓库**，存放跨服务契约、部署编排和基础设施配置。业务代码在各产品独立仓库中开发。

## 目录结构

```
polaris-base/
├── CLAUDE.md                          ← AI 全局约定
├── README.md                          ← 本文件
├── api/                               ← 跨服务 API 契约
│   └── proto/                         ← gRPC Protobuf 定义
├── products/                          ← 产品矩阵索引
├── deploy/                            ← 部署编排
│   ├── docker-compose/                ← 开发环境
│   └── k8s/                           ← 生产环境（K8s）
├── infra/                             ← 基础设施组件配置
│   ├── apisix/                        ← API 网关
│   ├── casdoor/                       ← IAM 身份认证
│   ├── safeline/                      ← WAF 防火墙
│   ├── postgres/                      ← 数据库
│   ├── redis/                         ← 缓存
│   ├── elasticsearch/                 ← 搜索引擎
│   └── minio/                         ← 对象存储
└── docs/                              ← 平台文档
    └── dev/
        └── 平台顶层技术栈.md
```

## 快速开始

```bash
# 1. 拉起基础设施（PostgreSQL / Redis / Elasticsearch / MinIO）
cd deploy/docker-compose
cp .env.example .env       # 修改密码等敏感配置
docker compose up -d

# 2. 验证服务状态
docker compose ps
```

## 架构分层

| 层 | 职责 | 核心组件 |
|---|---|---|
| 应用产品层 | 业务功能 | Spring Boot 服务、Python AI 服务、前端应用 |
| 公共基座层 | WAF + 网关 + IAM | SafeLine、APISIX、Casdoor |
| 基础设施层 | 存储 / 缓存 / 搜索 | PostgreSQL、MinIO、Redis、Elasticsearch |

## 相关文档

- [平台顶层技术栈](docs/dev/平台顶层技术栈.md)
