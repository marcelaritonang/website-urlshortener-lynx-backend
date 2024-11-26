.PHONY: dev prod build test clean logs help

# Development
dev: ## Start development environment
	docker-compose up --build

dev-d: ## Start development environment in detached mode
	docker-compose up -d --build

# Production
prod: ## Start production environment
	docker-compose -f docker-compose.prod.yml up --build

prod-d: ## Start production environment in detached mode
	docker-compose -f docker-compose.prod.yml up -d --build

# Build
build: ## Build all containers
	docker-compose build

build-dev: ## Build development containers
	docker-compose build

build-prod: ## Build production containers
	docker-compose -f docker-compose.prod.yml build

# Database
db-migrate: ## Run database migrations
	docker-compose exec api go run cmd/migrate/main.go up

db-rollback: ## Rollback database migrations
	docker-compose exec api go run cmd/migrate/main.go down

db-reset: ## Reset database (drop and recreate)
	docker-compose exec postgres psql -U ${POSTGRES_USER} -d postgres -c "DROP DATABASE IF EXISTS ${POSTGRES_DB};"
	docker-compose exec postgres psql -U ${POSTGRES_USER} -d postgres -c "CREATE DATABASE ${POSTGRES_DB};"
	make db-migrate

# Testing
test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Utilities
clean: ## Clean up containers and volumes
	docker-compose down -v
	rm -rf tmp/

logs: ## View logs
	docker-compose logs -f

redis-cli: ## Access Redis CLI
	docker-compose exec redis redis-cli -a ${REDIS_PASSWORD}

psql: ## Access PostgreSQL CLI
	docker-compose exec postgres psql -U ${POSTGRES_USER} -d ${POSTGRES_DB}

# Help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help