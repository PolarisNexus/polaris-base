# ai-mock — OpenAI 兼容 mock server（dev-only）

> 决策见 ADR-0014 Phase I MVP。

给 AI Gateway 端到端链路做本地冒烟用；**生产不启**（`profiles: ["dev"]`，默认启动命令不带它）。

## 覆盖

| 端点 | 行为 |
|---|---|
| `POST /v1/chat/completions` | 返 OpenAI 兼容响应；`assistant.content` = `"[ai-mock reply for model=<m>] echo: <last user msg>"` |
| `POST /v1/embeddings` | 返 8 维固定向量；`usage.prompt_tokens` 按输入字符数 / 4 估 |
| `GET /v1/models` | 返 3 个假模型列表（`gpt-4o-mini` / `gpt-3.5-turbo` / `text-embedding-3-small`） |
| `GET /healthz` | `ok` |

响应 schema 与 OpenAI 100% 对齐，APISIX `ai-proxy-multi` 把上游当真 OpenAI 即可。

## 不做

- **SSE streaming**：MVP 不验证；请求带 `"stream": true` 返 400
- **image / audio / fine-tuning / function-calling 真实语义**：mock 不是智能体
- **鉴权**：裸开，信任 polaris-net 网络边界 —— 生产环境不启动本服务

## 本地

```bash
# 命令行直跑（占 :8080，避开平台其它服务就改 HTTP_ADDR）
HTTP_ADDR=:18080 go run ./services/ai-mock/cmd/server

# Docker
COMPOSE_PROFILES=dev,services docker compose \
  -f deploy/docker-compose/docker-compose.yml up -d ai-mock
```
