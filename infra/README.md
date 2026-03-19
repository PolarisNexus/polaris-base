# infra/ — 基础设施组件配置

存放平台各基础设施组件的**自身配置文件**（非部署编排，编排在 `deploy/`）。

## 子目录

| 目录 | 组件 | 用途 |
|------|------|------|
| `apisix/` | Apache APISIX | API 网关 — 路由、JWT 校验、限流熔断 |
| `casdoor/` | Casdoor | IAM — 用户目录、SSO、JWT 颁发 |
| `safeline/` | SafeLine | WAF — 恶意流量拦截 |
| `postgres/` | PostgreSQL 16 | 核心业务数据库 |
| `redis/` | Redis 7 | 缓存、分布式锁、简单异步任务 |
| `elasticsearch/` | Elasticsearch 8.15 | 全文检索、日志聚合、初期向量检索 |
| `minio/` | MinIO | 统一对象存储（S3 协议） |
