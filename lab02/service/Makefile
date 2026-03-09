-include .env
export $(shell sed 's/=.*//' .env)

GOOSE_DRIVER = postgres
MIGRATIONS_DIR = migrations
DB_DSN := "host=$(POSTGRES_HOST) port=$(POSTGRES_PORT) dbname=$(POSTGRES_DB) user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) sslmode=disable"

TARGET ?= store

.PHONY: help
help:
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: db-up
db-up:
	docker-compose up -d

.PHONY: db-down
db-down:
	docker-compose down -v

.PHONY: migrate-up
migrate-up:
	@echo "Applying migrations from $(MIGRATIONS_DIR)..."
	@goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) $(DB_DSN) up

.PHONY: migrate-down
migrate-down:
	@echo "Rolling back last migration..."
	@goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) $(DB_DSN) down

.PHONY: migrate-status
migrate-status:
	@echo "Migration status in $(MIGRATIONS_DIR):"
	@goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) $(DB_DSN) status

.PHONY: migrate-create
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=your_migration_name"; \
		exit 1; \
	fi
	@goose -dir $(MIGRATIONS_DIR) create $(NAME) sql

.PHONY: migrate-reset
migrate-reset:
	@echo "Resetting database..."
	@goose -dir $(MIGRATIONS_DIR) postgres $(DB_DSN) reset
	@echo "Reapplying migrations in $(MIGRATIONS_DIR)..."
	@goose -dir $(MIGRATIONS_DIR) postgres $(DB_DSN) up

.PHONY: build
build:
	@echo "Выполняется go build для таргета ${TARGET}"
	@mkdir -p .bin
	@go build -o ./bin/${TARGET} ./cmd/${TARGET}

.PHONY: up
up:
	docker compose up -d --build

# Run full stack via Docker Compose
.PHONY: run
run: up migrate-up

# Stop and remove containers
.PHONY: down
down:
	docker compose down -v

# Follow service logs
.PHONY: logs
logs:
	docker compose logs -f service-products

.PHONY: lint
lint:
	@gofumpt -l -w . || true
	@golangci-lint run || true