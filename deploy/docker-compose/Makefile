.PHONY: help up up-data down ps logs config release clean

help:
	@echo "polaris-base 基座操作入口"
	@echo "  make up          启动全部组件"
	@echo "  make up-data     仅启动数据层（PG/Redis）"
	@echo "  make down        停止全部"
	@echo "  make ps          查看状态"
	@echo "  make logs        查看日志（S=postgres 指定服务）"
	@echo "  make config      输出完整合并配置"
	@echo "  make release     生成单文件 docker-compose.full.yml"
	@echo "  make clean       停止 + 删除命名卷（需二次确认）"

up:
	docker compose up -d

up-data:
	docker compose up -d postgres redis

down:
	docker compose down

ps:
	docker compose ps

logs:
	docker compose logs -f $(S)

config:
	docker compose config

PROJECT_ROOT := $(shell cd ../.. && pwd)
release:
	docker compose config | sed 's|source: $(PROJECT_ROOT)/|source: ../../|g' > docker-compose.full.yml
	@echo "已生成 docker-compose.full.yml"

clean:
	@read -p "将 down 并删除命名卷（数据不可恢复），输入 YES 确认: " ans && [ "$$ans" = "YES" ]
	docker compose down -v
