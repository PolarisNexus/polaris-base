# ADR-0003: No message queue initially

## Status

Accepted

## Context

微服务架构中消息队列（Kafka、RabbitMQ）常用于异步解耦、事件驱动。但初期业务服务尚未上线，异步场景不明确，过早引入 MQ 会增加运维负担和架构复杂度。

## Decision

初期**不引入消息队列**。异步需求通过 Redis（List/Stream）或数据库轮询满足。

## Consequences

- 减少基础设施组件数量，降低运维和理解成本
- Redis Stream 可满足简单的生产者-消费者模式
- 当出现以下信号时引入 MQ：消息量级 > Redis 承载能力、需要消息回溯/持久化保证、多消费者组独立消费
- 迁移时需要替换 Redis Stream 为 Kafka/RabbitMQ consumer

## Alternatives Considered

- **初期引入 Kafka**：功能强大但运维沉重（ZooKeeper/KRaft、分区管理、topic 治理），初期无量级支撑需求。
- **初期引入 RabbitMQ**：比 Kafka 轻量，但仍增加一个有状态服务。当前没有明确的异步场景证明其必要性。
