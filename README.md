# polaris-base

北极星枢纽的基座，一个多语言微服务（Polyglot Microservices）的底座平台。

## 仓库定位

polaris-base 是 PolarisNexus 平台的**共享基座仓库**，存放跨服务契约、部署编排和组件配置。业务代码在各产品独立仓库中开发。

## 目录结构

```
polaris-base/
├── CLAUDE.md                          ← AI 全局约定
├── README.md                          ← 本文件
├── api/                               ← 跨服务 API 契约
│   └── proto/                         ← gRPC Protobuf 定义
├── products/                          ← 产品矩阵索引
├── deploy/                            ← 部署编排
│   ├── docker-compose/                ← Docker Compose 部署入口
│   └── kubernetes/                    ← K8s 部署（预留）
├── components/                        ← 组件配置与编排（组件即包）
│   ├── apisix/                        ← API 网关
│   ├── casdoor/                       ← IAM 身份认证
│   ├── safeline/                      ← WAF 防火墙（暂未纳入编排）
│   ├── postgres/                      ← 数据库
│   ├── redis/                         ← 缓存
│   ├── elastic/                       ← Elastic 家族
│   │   └── elasticsearch/             ← 搜索引擎
│   ├── minio/                         ← 对象存储
│   └── observability/                 ← 可观测性栈（占位）
├── scripts/                           ← 运维助手脚本
├── docs/                              ← 平台文档
│   ├── architecture/                  ← 架构文档
│   └── adr/                           ← 架构决策记录
```

## 快速开始

```bash
cd deploy/docker-compose
docker compose up -d

# 验证服务状态
docker compose ps
```

所有环境变量都内置默认值，**零配置即可启动**。如需自定义密码等参数，各组件目录下有 `.env.example` 可参考。

## 架构分层

| 层 | 职责 | 核心组件 |
|---|---|---|
| 应用产品层 | 业务功能 | Spring Boot 服务、Python AI 服务、前端应用 |
| 公共基座层 | WAF + 网关 + IAM | SafeLine、APISIX、Casdoor |
| 基础设施层 | 存储 / 缓存 / 搜索 | PostgreSQL、MinIO、Redis、Elasticsearch |

## 相关文档

- [平台顶层技术栈](docs/architecture/平台顶层技术栈.md)
- [架构决策记录](docs/adr/)
