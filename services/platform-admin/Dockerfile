FROM golang:1.23.4-alpine AS build

WORKDIR /src

# 复制 go.work 与其引用的所有 module（否则 go.work 解析会失败）。
# 本服务实际只依赖 api/gen/go 与自己，但 go.work 需要所有 use 路径存在。
COPY go.work go.work.sum* ./
COPY api/gen/go ./api/gen/go
COPY services/platform-admin ./services/platform-admin
COPY services/crowdsec-bouncer ./services/crowdsec-bouncer
COPY services/ai-mock ./services/ai-mock

WORKDIR /src/services/platform-admin
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/platform-admin ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/platform-admin /platform-admin
EXPOSE 8080
ENTRYPOINT ["/platform-admin"]
