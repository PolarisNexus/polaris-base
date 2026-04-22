# _template — 新服务脚手架

复制本目录到 `services/<your-service>/`，按清单替换占位符：

- `<service>` → 服务名（小写、短横线连接，如 `membership`）
- `<Service>` → Pascal Case（如 `Membership`）
- proto 包名：`polaris.commons.<service>.v1`

## 目录结构

```
<service>/
├── cmd/server/main.go         # 入口
├── internal/
│   ├── server/                # gRPC server 实现
│   ├── repo/                  # 数据访问（连 polaris-base-data）
│   └── service/               # 业务逻辑
├── web/                       # React + AntD Pro 前端
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
├── Dockerfile                 # 后端镜像（multi-stage）
├── Dockerfile.web             # 前端镜像（nginx 静态托管）
├── docker-compose.yml         # 服务 compose 片段（被顶层 include）
├── go.mod
└── README.md
```

## 关键文件说明

### `cmd/server/main.go`
- 初始化 OTel（tracer + meter）
- 启动 gRPC server + 可选 HTTP gateway
- 监听信号优雅关闭

### `docker-compose.yml`
- 声明服务容器 + 端口 + 环境变量
- 挂 `polaris-net` 网络，别名 `base-<service>`
- 健康检查走 gRPC health protocol

### proto 契约位置
**不在**本服务目录下，而在仓库根 `api/proto/<service>/v1/service.proto`（统一管理）。

## 本地开发

```bash
# 生成 proto 桩代码（在仓库根）
buf generate

# 后端
cd services/<service>
go run ./cmd/server

# 前端
cd services/<service>/web
pnpm dev
```

## 注册到基座

1. 顶层 `deploy/docker-compose/docker-compose.yml` 添加：
   ```yaml
   - path: ../../services/<service>/docker-compose.yml
   ```
2. `components/apisix/routes/` 新增 `NN-<service>.yaml`（ADR-0002 Git 源模型）：
   ```yaml
   routes:
     - id: <service>
       uri: /api/<service>/*
       upstream:
         nodes:
           "base-<service>:8080": 1
   ```
   合入后 CI 跑 `scripts/apisix-apply-routes.sh` 写入 etcd
3. 更新 `services/README.md` 服务清单

## 待补充

此骨架目前仅包含 README + 目录约定，具体代码脚手架（main.go、Dockerfile、前端初始化）在首个服务立项时补齐并沉淀回本模板。
