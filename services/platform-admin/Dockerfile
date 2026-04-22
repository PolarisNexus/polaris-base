FROM golang:1.23.4-alpine AS build

WORKDIR /src

# 复制整仓（需要 go.work + 两个 module 才能构建）。
COPY go.work go.work.sum* ./
COPY api/gen/go ./api/gen/go
COPY services/platform-admin ./services/platform-admin

WORKDIR /src/services/platform-admin
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/platform-admin ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/platform-admin /platform-admin
EXPOSE 8080
ENTRYPOINT ["/platform-admin"]
