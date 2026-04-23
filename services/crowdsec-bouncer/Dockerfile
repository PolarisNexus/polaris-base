FROM golang:1.23.4-alpine AS build

WORKDIR /src

# 复制仓库根的 go.work 与其引用的 module。
# 本 bouncer 当前仅依赖标准库，但保留 go.work 聚合入口便于未来共享 proto / pkg。
COPY go.work go.work.sum* ./
COPY api/gen/go ./api/gen/go
COPY services/platform-admin ./services/platform-admin
COPY services/crowdsec-bouncer ./services/crowdsec-bouncer

WORKDIR /src/services/crowdsec-bouncer
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/crowdsec-bouncer ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/crowdsec-bouncer /crowdsec-bouncer
EXPOSE 8080
ENTRYPOINT ["/crowdsec-bouncer"]
