# infra/safeline/ — SafeLine WAF 配置

SafeLine 部署在 APISIX 前，作为反向代理 WAF 拦截恶意流量。

## 防护能力

- OWASP Top 10（SQL 注入、XSS、CSRF、路径遍历等）
- CC 攻击防护 / 速率限制
- 爬虫识别与人机验证
- IP 黑白名单

## 预期内容

- 自定义防护规则集
- 白名单配置
- 与 APISIX 的联动配置
