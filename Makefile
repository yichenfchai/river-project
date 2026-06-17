.PHONY: help run run-infra stop clean db-init

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

run-infra: ## 启动 PostgreSQL + Redis
	docker compose -f deploy/docker-compose/docker-compose.yml up -d
	@echo "等待 PostgreSQL 就绪..."
	@sleep 3

run: run-infra ## 启动基础设施 + 主服务
	@echo "启动主服务 (单体模式, :8080)..."
	cd services/main-service && go run cmd/main.go

db-init: ## 初始化数据库
	docker compose -f deploy/docker-compose/docker-compose.yml exec -T postgres psql -U canal -d grand_canal -f /docker-entrypoint-initdb.d/init.sql

stop: ## 停止所有容器
	docker compose -f deploy/docker-compose/docker-compose.yml down

clean: stop ## 停止并删除数据卷
	docker compose -f deploy/docker-compose/docker-compose.yml down -v

mod-tidy: ## go mod tidy
	cd pkg && go mod tidy
	cd services/main-service && go mod tidy

test: ## 运行测试
	go test ./...
