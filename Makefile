# polaris-base 便利启动层
#
# 单仓 + 单 compose 项目（name: polaris-base）
# plane 通过 compose profile 激活；role 通过 label 过滤（ADR-0009）
#
# 默认 .env 激活 data + platform + services 全部 profile。
# 按 plane 选择性启动时覆盖环境变量：COMPOSE_PROFILES=data make up

COMPOSE_FILE := deploy/docker-compose/docker-compose.yml
COMPOSE      := docker compose -f $(COMPOSE_FILE)

.PHONY: up up-data up-platform up-services down restart ps ps-data ps-platform ps-services logs config release clean help

## up: 启动全量（data + platform + services）
up:
	$(COMPOSE) up -d

## up-data: 仅启动 data plane（PG / Redis / ES / MinIO）
up-data:
	COMPOSE_PROFILES=data $(COMPOSE) up -d

## up-platform: 仅启动 platform plane（etcd / APISIX / Authentik / observability）
up-platform:
	COMPOSE_PROFILES=platform $(COMPOSE) up -d

## up-services: 仅启动 services plane（自研共享服务）
up-services:
	COMPOSE_PROFILES=services $(COMPOSE) up -d

## down: 停止并清理全部容器
down:
	$(COMPOSE) down

## restart: down + up
restart: down up

## ps: 查看全部容器
ps:
	$(COMPOSE) ps

## ps-data: 按 label 过滤 data plane 容器
ps-data:
	docker ps --filter label=com.docker.compose.project=polaris-base --filter label=polaris.plane=data

## ps-platform: 按 label 过滤 platform plane 容器
ps-platform:
	docker ps --filter label=com.docker.compose.project=polaris-base --filter label=polaris.plane=platform

## ps-services: 按 label 过滤 services plane 容器
ps-services:
	docker ps --filter label=com.docker.compose.project=polaris-base --filter label=polaris.plane=services

## logs: 查看日志（S=<svc> 指定服务）
logs:
	$(COMPOSE) logs -f $(S)

## config: 输出合并后的完整配置
config:
	$(COMPOSE) config

## release: 生成单文件 deploy/docker-compose/docker-compose.full.yml（离线分发用）
release:
	$(COMPOSE) config | sed 's|source: $(CURDIR)/|source: ../../|g' > deploy/docker-compose/docker-compose.full.yml
	@echo "已生成 deploy/docker-compose/docker-compose.full.yml"

## clean: down + 删除命名卷（数据不可恢复，需输入 YES 确认）
clean:
	@read -p "将 down 并删除命名卷（数据不可恢复），输入 YES 确认: " ans && [ "$$ans" = "YES" ]
	$(COMPOSE) down -v

## help: 显示本帮助
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
