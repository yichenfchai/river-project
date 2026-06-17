.PHONY: help run-all run-infra run-services stop clean db-init

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run-infra: ## Start PostgreSQL + Redis
	docker compose -f deploy/docker-compose/docker-compose.yml up -d postgres redis
	@echo "Waiting for PostgreSQL..."
	@sleep 3
	@echo "Infrastructure ready!"

run-all: run-infra ## Start infrastructure + all services
	@echo "Starting all services..."
	@cd services/user-service && go run cmd/main.go &
	@cd services/content-service && go run cmd/main.go &
	@cd services/quiz-service && go run cmd/main.go &
	@cd services/api-gateway && go run cmd/main.go &
	@echo "All services started. Gateway: http://localhost:8080"

db-init: ## Initialize database with seed data
	docker compose -f deploy/docker-compose/docker-compose.yml exec -T postgres psql -U canal -d grand_canal -f /docker-entrypoint-initdb.d/init.sql

stop: ## Stop all Docker containers
	docker compose -f deploy/docker-compose/docker-compose.yml down

clean: stop ## Stop and remove volumes
	docker compose -f deploy/docker-compose/docker-compose.yml down -v

mod-tidy: ## Run go mod tidy for all modules
	cd pkg && go mod tidy
	cd services/user-service && go mod tidy
	cd services/content-service && go mod tidy
	cd services/quiz-service && go mod tidy
	cd services/api-gateway && go mod tidy

test: ## Run tests
	go test ./...

proto-gen: ## Generate protobuf code
	protoc --go_out=proto/gen/go --go_opt=paths=source_relative proto/*.proto
