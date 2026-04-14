# polaris-base-commons

PolarisNexus 平台**公共能力层**——网关、IAM、WAF、可观测性、自研共享业务模块（邮件/会员/支付等）。

与 `polaris-base-data`（数据持久化底座）并列，共同构成平台必启底座（见 ADR-0009）。

> 仓库历史名为 `polaris-base`；拆分后承担 commons 职责，远端仓库重命名由用户在 GitHub/GitLab 后台完成。

## 目录结构

```
polaris-base-commons/
├── api/proto/              跨服务 gRPC 契约
├── components/             第三方基础设施（黑盒编排）
│   ├── apisix/             API 网关
│   ├── casdoor/            IAM（待 ADR-0010 重选型）
│   ├── observability/      OTel Elasticsearch + Kibana
│   └── safeline/           WAF（占位）
├── services/               自研共享业务服务（Go + React + AntD Pro，见 ADR-0011）
│   ├── README.md
│   └── _template/          新服务脚手架
├── deploy/docker-compose/  顶层 compose 入口
├── products/               产品矩阵索引
├── scripts/
├── docs/adr/               架构决策记录
├── Makefile                便利启动（同时启动 commons + data）
├── CLAUDE.md
└── README.md
```

## 快速开始

```bash
# 同时启动 commons + data（推荐）
make up

# 或仅启动 commons
cd deploy/docker-compose && docker compose up -d
```

**前置**：同级目录存在 `polaris-base-data` 仓库。

## 架构分层

| 层 | 仓库 | 核心 |
|---|---|---|
| 产品层 | `polaris-<product>` | 业务代码 |
| Commons（本仓） | `polaris-base-commons` | 网关 + IAM + WAF + 可观测性 + 自研共享服务 |
| Data | `polaris-base-data` | PG / Redis / ES / MinIO / 未来图&向量&消息 |

## 相关文档

- [架构决策记录](docs/adr/)（核心：ADR-0001、ADR-0005、ADR-0009、ADR-0011）
- 平台顶层技术栈：`docs/architecture/`
