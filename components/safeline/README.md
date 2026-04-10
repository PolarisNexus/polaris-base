# SafeLine WAF

SafeLine 需要 7 个容器（PostgreSQL、管理端、检测引擎、Tengine、Luigi、FVM、Chaos），使用独立子网和 host 网络模式，与平台组件编排差异较大。

## 防护能力

- OWASP Top 10（SQL 注入、XSS、CSRF、路径遍历等）
- CC 攻击防护 / 速率限制
- 爬虫识别与人机验证
- IP 黑白名单

## 接入方式

建议按照 [SafeLine 官方文档](https://docs.waf-ce.chaitin.cn/) 独立部署，部署完成后在 APISIX 上游指向 SafeLine 的 Tengine 端口即可串联。

## 后续计划

- 独立部署 SafeLine 后，在 `infra/apisix/config.yaml` 中将流量链路调整为：客户端 → APISIX → SafeLine → 后端服务
- 或反过来：客户端 → SafeLine → APISIX → 后端服务（取决于安全策略）
- 自定义防护规则集、白名单配置
