# deploy/k8s/ — Kubernetes 生产部署

生产环境使用离线 Kubernetes 集群编排，当前为占位目录。

## 规划

- 离线部署，零外部依赖
- 兼容 ARM 架构（麒麟/统信 UOS）
- HPA 自动扩容
- CoreDNS 服务发现

## TODO

- [ ] 基础设施组件 Helm Chart / Kustomize 清单
- [ ] 产品服务 Deployment 模板
- [ ] Ingress / APISIX 路由配置
