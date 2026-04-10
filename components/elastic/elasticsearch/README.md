# infra/elasticsearch/ — Elasticsearch 配置

Elasticsearch 8.15 用于全文检索、日志聚合和初期向量检索。

## 预期内容

- Index Template / ILM 策略
- 向量检索（dense_vector）索引映射
- 可观测性相关索引配置（OTel / Elastic APM）

## 注意

- 初期使用 Elasticsearch 内置 dense_vector 满足向量检索需求
- 未来按需切换至 Milvus / Qdrant 专用引擎
