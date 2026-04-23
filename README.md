# polaris-base

PolarisNexus 平台**共享基座**——网关、IAM、WAF、可观测性、数据持久化、自研共享业务模块（邮件/会员/支付等）。

单仓 + 单 Compose 项目（`name: polaris-base`）。plane 分层（data / platform / services）通过 Compose profile + label 表达，不拆仓、不分目录（见 ADR-0009）。

## 目录结构

```
polaris-base/
├── api/proto/              跨服务 gRPC 契约
├── components/             第三方基础设施（扁平，按组件）
│   ├── postgres/           relational-db       (plane: data)
│   ├── redis/              cache               (plane: data)
│   ├── elastic/            search              (plane: data)
│   ├── minio/              object-storage      (plane: data)
│   ├── etcd/               config-store        (plane: platform)
│   ├── apisix/             gateway + Coraza WAF (plane: platform，ADR-0002/0012)
│   ├── crowdsec/           waf-bot             (plane: platform，ADR-0012)
│   ├── authentik/          iam                 (plane: platform，ADR-0010)
│   └── observability/      OTel ES + Kibana    (plane: platform)
├── services/               自研共享业务服务（Go + React + AntD Pro，见 ADR-0011）
│   ├── platform-admin/     WAF + Gateway 管理控制台（ADR-0013）
│   ├── crowdsec-bouncer/   APISIX forward-auth 侧车，执行 CrowdSec 决策（ADR-0012）
│   ├── README.md
│   └── _template/          新服务脚手架
├── deploy/docker-compose/  顶层 compose 入口 + .env（默认 profile 激活）
├── products/               产品矩阵索引
├── scripts/
├── docs/adr/               架构决策记录
├── Makefile                便利启动（up / up-data / up-platform / ps-data ...）
├── CLAUDE.md
└── README.md
```

## 快速开始

```bash
# 一行启动全量（默认 .env 激活 data + platform + services）
docker compose -f deploy/docker-compose/docker-compose.yml up -d

# 或通过 Makefile
make up         # 全量
make up-data    # 仅 data plane
make up-platform  # 仅 platform plane
make ps-data    # 过滤 data plane 容器视图
```

## Plane 与 Role

| plane | 组件 | role |
|-------|------|------|
| **data** | postgres / redis / elastic / minio | relational-db / cache / search / object-storage |
| **platform** | etcd / apisix / crowdsec / authentik / observability | config-store / gateway / waf-bot / iam / observability |
| **services** | platform-admin / crowdsec-bouncer | admin-console / waf-bouncer |

每个 compose 服务自带：

```yaml
profiles: ["<plane>"]
labels:
  polaris.plane: <plane>
  polaris.role: <role>
```

按需启动：`COMPOSE_PROFILES=data docker compose up -d`
按 plane 视图：`docker ps --filter label=com.docker.compose.project=polaris-base --filter label=polaris.plane=data`
按 role 视图：`docker ps --filter label=com.docker.compose.project=polaris-base --filter label=polaris.role=cache`

（`docker compose ps` 仅支持 status 过滤，plane/role 用 `docker ps` 按 label 过滤或直接 `make ps-data` / `make ps-platform`。）

## 相关文档

- [架构决策记录](docs/adr/)（核心：ADR-0001、ADR-0002、ADR-0005、ADR-0009、ADR-0010、ADR-0011、ADR-0012、ADR-0013）
- 平台顶层技术栈：`docs/architecture/`
