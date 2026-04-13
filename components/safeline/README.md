# SafeLine — WAF 防火墙（暂缓接入）

SafeLine 需要 7 个容器、独立子网和 host 网络模式，与平台编排差异较大，暂不纳入 compose 编排（详见 ADR-0006）。

## 接入方式

按照 [SafeLine 官方文档](https://docs.waf-ce.chaitin.cn/) 独立部署，在 APISIX 上游指向 SafeLine 的 Tengine 端口即可串联。
