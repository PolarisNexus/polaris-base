# polaris-base 便利启动层
#
# 单仓 + 单 compose 项目（name: polaris-base）
# plane 通过 compose profile 激活；role 通过 label 过滤（ADR-0009）
#
# 默认 .env 激活 data + platform + services 全部 profile。
# 按 plane 选择性启动时覆盖环境变量：COMPOSE_PROFILES=data make up-raw

COMPOSE := docker compose -f deploy/docker-compose/docker-compose.yml

.PHONY: up up-data up-platform up-services down restart ps ps-data ps-platform ps-services logs help

## up: 启动全量（data + platform + services）
up:
	$(COMPOSE) up -d

## up-data: 仅启动 data plane（PG / Redis / ES / MinIO）
up-data:
	COMPOSE_PROFILES=data $(COMPOSE) up -d

## up-platform: 仅启动 platform plane（APISIX / Casdoor / observability）
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

## logs: 提示日志命令
logs:
	@echo "use: $(COMPOSE) logs -f <svc>"

## help: 显示本帮助
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
