# polaris-base-commons 便利启动层
#
# 本仓库（commons）与 polaris-base-data 是并列的平台底座，必须成对启动。
# 约定 polaris-base-data 位于 ../polaris-base-data。
#
# 单独启动 commons：docker compose -f deploy/docker-compose/docker-compose.yml up -d
# 单独启动 data：  cd ../polaris-base-data && docker compose up -d

DATA_DIR ?= ../polaris-base-data
COMMONS_COMPOSE := deploy/docker-compose/docker-compose.yml

.PHONY: up down ps logs restart help

## up: 依次启动 commons（创建 polaris-net）+ data
up:
	@echo ">>> [1/2] polaris-base-commons up"
	docker compose -f $(COMMONS_COMPOSE) up -d
	@echo ">>> [2/2] polaris-base-data up"
	@cd $(DATA_DIR) && docker compose up -d
	@echo ">>> done"

## down: 依次停止 data + commons
down:
	@echo ">>> [1/2] polaris-base-data down"
	@cd $(DATA_DIR) && docker compose down
	@echo ">>> [2/2] polaris-base-commons down"
	docker compose -f $(COMMONS_COMPOSE) down
	@echo ">>> done"

## ps: 查看两仓容器状态
ps:
	@echo ">>> polaris-base-commons"
	docker compose -f $(COMMONS_COMPOSE) ps
	@echo ""
	@echo ">>> polaris-base-data"
	@cd $(DATA_DIR) && docker compose ps

## logs: 跟踪两仓日志
logs:
	@echo "use: docker compose -f $(COMMONS_COMPOSE) logs -f <svc>"
	@echo "  or: (cd $(DATA_DIR) && docker compose logs -f <svc>)"

## restart: down + up
restart: down up

## help: 显示本帮助
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
