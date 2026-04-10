# deploy/ — 部署编排

存放平台所有组件的部署编排文件。

## 目录结构

```
deploy/
├── docker-compose/             ← Docker Compose 部署入口
│   ├── docker-compose.yml      ← include 聚合所有 components/
│   ├── .env.example            ← 公共变量模板（可选）
│   ├── Makefile                ← 便捷操作
│   └── README.md               ← 部署操作详细说明
└── kubernetes/                 ← K8s 部署（预留）
```

## 开发环境

```bash
cd deploy/docker-compose
docker compose up -d            # 零配置一行启动
```

详细说明见 [docker-compose/README.md](docker-compose/README.md)。

## 生产环境

生产使用离线 Kubernetes 集群部署，详见 `kubernetes/README.md`。
