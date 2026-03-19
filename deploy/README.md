# deploy/ — 部署编排

存放平台所有组件的部署编排文件。

## 目录结构

```
deploy/
├── docker-compose/     ← 开发环境，Docker Compose 一键拉起
│   ├── docker-compose.yml
│   └── .env.example
└── k8s/                ← 生产环境，Kubernetes 编排（待完善）
```

## 开发环境

```bash
cd deploy/docker-compose
cp .env.example .env    # 按需修改配置
docker compose up -d    # 拉起基础设施
```

## 生产环境

生产使用离线 Kubernetes 集群部署，详见 `k8s/README.md`。
