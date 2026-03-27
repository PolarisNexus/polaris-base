# infra/ — 基础设施组件配置与编排

每个子目录维护该组件的独立 `docker-compose.yml` 和相关配置文件。
总入口 `deploy/docker-compose/docker-compose.yml` 通过 `include` 指令引用各组件，实现一键启动。

## 子目录

| 目录 | 组件 | 用途 |
|------|------|------|
| `apisix/` | Apache APISIX | API 网关 — 路由、JWT 校验、限流熔断 |
| `casdoor/` | Casdoor | IAM — 用户目录、SSO、JWT 颁发 |
| `safeline/` | SafeLine | WAF — 恶意流量拦截（暂未纳入编排，独立部署） |
| `postgres/` | PostgreSQL 16 | 核心业务数据库 |
| `redis/` | Redis 7 | 缓存、分布式锁、简单异步任务 |
| `elasticsearch/` | Elasticsearch 8.15 | 全文检索、日志聚合、初期向量检索 |
| `minio/` | MinIO | 统一对象存储（S3 协议） |

## 快速启动

```bash
cd deploy/docker-compose
cp .env.example .env   # 编辑 .env 设置实际密码
docker compose up -d
```

按需裁剪：注释掉 `deploy/docker-compose/docker-compose.yml` 中对应的 include 行即可跳过某个组件。
