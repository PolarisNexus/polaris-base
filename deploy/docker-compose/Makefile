.PHONY: init up down ps logs config release clean

ENV_FILE := .env

# 首次初始化：从模板生成 .env
init:
	@test -f $(ENV_FILE) || (cp .env.example $(ENV_FILE) && echo "已生成 .env，请修改密码后再启动")
	@test -f $(ENV_FILE) && echo ".env 已存在"

# 启动全部组件
up: init
	docker compose up -d

# 停止全部组件
down:
	docker compose down

# 停止并删除数据卷（危险）
clean:
	@echo "即将删除所有数据卷，Ctrl+C 取消"; sleep 3
	docker compose down -v

# 查看状态
ps:
	docker compose ps

# 查看日志（可追加 S=postgres 指定服务）
logs:
	docker compose logs -f $(S)

# 输出完整合并配置（排查用）
config: init
	docker compose config

# 生成可发布的单文件 docker-compose.full.yml（绝对路径转相对）
PROJECT_ROOT := $(shell cd ../.. && pwd)
release: init
	docker compose config | sed 's|source: $(PROJECT_ROOT)/|source: ../../|g' > docker-compose.full.yml
	@echo "已生成 docker-compose.full.yml"
