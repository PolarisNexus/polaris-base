FROM golang:1.23.4-alpine AS build

WORKDIR /src
COPY go.work go.work.sum* ./
COPY api/gen/go ./api/gen/go
COPY services/platform-admin ./services/platform-admin
COPY services/crowdsec-bouncer ./services/crowdsec-bouncer
COPY services/ai-mock ./services/ai-mock

WORKDIR /src/services/ai-mock
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/ai-mock ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/ai-mock /ai-mock
EXPOSE 8080
ENTRYPOINT ["/ai-mock"]
