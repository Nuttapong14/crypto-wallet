.PHONY: help setup run dev test test-unit test-integration test-coverage lint fmt build clean db-create db-drop db-reset db-seed migrate-up migrate-down migrate-status migration db-console-core db-console-kyc db-console-rates db-console-audit db-ping db-backup db-restore db-schema docker-build docker-up docker-down docker-logs

# Variables
BINARY_NAME=server
BINARY_PATH=bin/$(BINARY_NAME)
GO_FILES=$(shell find . -name '*.go' -type f -not -path "./vendor/*")
GOOSE=goose
MIGRATION_DIR=db/migrations

# Database configuration
CORE_DB_DSN ?= postgresql://user:pass@localhost:5432/core_db?sslmode=disable
KYC_DB_DSN ?= postgresql://user:pass@localhost:5433/kyc_db?sslmode=disable
RATES_DB_DSN ?= postgresql://user:pass@localhost:5434/rates_db?sslmode=disable
AUDIT_DB_DSN ?= postgresql://user:pass@localhost:5435/audit_db?sslmode=disable

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Complete project setup (install deps, create DBs, run migrations, seed)
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy
	@echo "Creating databases..."
	$(MAKE) db-create
	@echo "Running migrations..."
	$(MAKE) migrate-up
	@echo "Seeding databases..."
	$(MAKE) db-seed
	@echo "Setup complete!"

run: ## Start server in production mode
	@echo "Starting server..."
	go run cmd/server/main.go

dev: ## Start server in development mode with hot-reload (requires air)
	@echo "Starting development server with hot-reload..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "air not installed. Install with: go install github.com/air-verse/air@latest"; \
		echo "Falling back to normal run..."; \
		go run cmd/server/main.go; \
	fi

test: ## Run all tests
	@echo "Running all tests..."
	go test -v -race -timeout 30s ./...

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	go test -v -race -short -timeout 30s ./internal/domain/... ./internal/application/... ./pkg/...

test-integration: ## Run integration tests (requires Docker)
	@echo "Running integration tests..."
	go test -v -race -timeout 60s ./tests/integration/...

test-coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run golangci-lint
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run --timeout 5m; \
	else \
		echo "golangci-lint not installed. Install from: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

fmt: ## Format code with gofmt
	@echo "Formatting code..."
	gofmt -s -w $(GO_FILES)
	@echo "Running goimports..."
	@if command -v goimports > /dev/null; then \
		goimports -w $(GO_FILES); \
	else \
		echo "goimports not installed. Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

build: ## Build production binary
	@echo "Building production binary..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $(BINARY_PATH) cmd/server/main.go
	@echo "Binary created at: $(BINARY_PATH)"

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

# Database Management

db-create: ## Create all databases
	@echo "Creating databases..."
	@docker-compose exec -T postgres-core psql -U user -d postgres -c "CREATE DATABASE core_db;" 2>/dev/null || echo "core_db already exists"
	@docker-compose exec -T postgres-kyc psql -U user -d postgres -c "CREATE DATABASE kyc_db;" 2>/dev/null || echo "kyc_db already exists"
	@docker-compose exec -T postgres-rates psql -U user -d postgres -c "CREATE DATABASE rates_db;" 2>/dev/null || echo "rates_db already exists"
	@docker-compose exec -T postgres-audit psql -U user -d postgres -c "CREATE DATABASE audit_db;" 2>/dev/null || echo "audit_db already exists"
	@echo "Databases created!"

db-drop: ## Drop all databases (⚠️ WARNING: DESTRUCTIVE)
	@echo "⚠️  WARNING: This will DELETE all databases!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker-compose exec -T postgres-core psql -U user -d postgres -c "DROP DATABASE IF EXISTS core_db;"; \
		docker-compose exec -T postgres-kyc psql -U user -d postgres -c "DROP DATABASE IF EXISTS kyc_db;"; \
		docker-compose exec -T postgres-rates psql -U user -d postgres -c "DROP DATABASE IF EXISTS rates_db;"; \
		docker-compose exec -T postgres-audit psql -U user -d postgres -c "DROP DATABASE IF EXISTS audit_db;"; \
		echo "Databases dropped!"; \
	else \
		echo "Cancelled."; \
	fi

db-reset: db-drop db-create migrate-up ## Reset databases (drop, create, migrate)

db-seed: ## Seed initial data
	@echo "Seeding databases..."
	@if [ -f scripts/seed.sh ]; then \
		bash scripts/seed.sh; \
	else \
		echo "No seed script found at scripts/seed.sh"; \
	fi

# Database Migrations (using goose)

migrate-up: ## Apply all pending migrations
	@echo "Running migrations..."
	@echo "Core DB..."
	$(GOOSE) -dir $(MIGRATION_DIR)/core_db postgres "$(CORE_DB_DSN)" up
	@echo "KYC DB..."
	$(GOOSE) -dir $(MIGRATION_DIR)/kyc_db postgres "$(KYC_DB_DSN)" up
	@echo "Rates DB..."
	$(GOOSE) -dir $(MIGRATION_DIR)/rates_db postgres "$(RATES_DB_DSN)" up
	@echo "Audit DB..."
	$(GOOSE) -dir $(MIGRATION_DIR)/audit_db postgres "$(AUDIT_DB_DSN)" up
	@echo "All migrations applied!"

migrate-down: ## Rollback last migration
	@echo "Rolling back migrations..."
	@read -p "Rollback core_db? [y/N] " -n 1 -r; echo; [[ $$REPLY =~ ^[Yy]$$ ]] && $(GOOSE) -dir $(MIGRATION_DIR)/core_db postgres "$(CORE_DB_DSN)" down || true
	@read -p "Rollback kyc_db? [y/N] " -n 1 -r; echo; [[ $$REPLY =~ ^[Yy]$$ ]] && $(GOOSE) -dir $(MIGRATION_DIR)/kyc_db postgres "$(KYC_DB_DSN)" down || true
	@read -p "Rollback rates_db? [y/N] " -n 1 -r; echo; [[ $$REPLY =~ ^[Yy]$$ ]] && $(GOOSE) -dir $(MIGRATION_DIR)/rates_db postgres "$(RATES_DB_DSN)" down || true
	@read -p "Rollback audit_db? [y/N] " -n 1 -r; echo; [[ $$REPLY =~ ^[Yy]$$ ]] && $(GOOSE) -dir $(MIGRATION_DIR)/audit_db postgres "$(AUDIT_DB_DSN)" down || true

migrate-status: ## Show migration status for all databases
	@echo "=== Core DB Status ==="
	@$(GOOSE) -dir $(MIGRATION_DIR)/core_db postgres "$(CORE_DB_DSN)" status
	@echo ""
	@echo "=== KYC DB Status ==="
	@$(GOOSE) -dir $(MIGRATION_DIR)/kyc_db postgres "$(KYC_DB_DSN)" status
	@echo ""
	@echo "=== Rates DB Status ==="
	@$(GOOSE) -dir $(MIGRATION_DIR)/rates_db postgres "$(RATES_DB_DSN)" status
	@echo ""
	@echo "=== Audit DB Status ==="
	@$(GOOSE) -dir $(MIGRATION_DIR)/audit_db postgres "$(AUDIT_DB_DSN)" status

migration: ## Create new migration (usage: make migration name=migration_name)
	@if [ -z "$(name)" ]; then \
		echo "Error: name parameter required. Usage: make migration name=migration_name"; \
		exit 1; \
	fi
	@echo "Which database? [core/kyc/rates/audit]"
	@read -p "Database: " db; \
	case $$db in \
		core|kyc|rates|audit) \
			$(GOOSE) -dir $(MIGRATION_DIR)/$${db}_db create $(name) sql; \
			echo "Migration created in $(MIGRATION_DIR)/$${db}_db/"; \
			;; \
		*) \
			echo "Invalid database. Choose: core, kyc, rates, or audit"; \
			exit 1; \
			;; \
	esac

migrate-force: ## Force migration version (usage: make migrate-force version=X db=core)
	@if [ -z "$(version)" ] || [ -z "$(db)" ]; then \
		echo "Error: version and db parameters required"; \
		echo "Usage: make migrate-force version=X db=core"; \
		exit 1; \
	fi
	@case $(db) in \
		core) DSN=$(CORE_DB_DSN);; \
		kyc) DSN=$(KYC_DB_DSN);; \
		rates) DSN=$(RATES_DB_DSN);; \
		audit) DSN=$(AUDIT_DB_DSN);; \
		*) echo "Invalid db. Choose: core, kyc, rates, or audit"; exit 1;; \
	esac; \
	$(GOOSE) -dir $(MIGRATION_DIR)/$(db)_db postgres "$$DSN" version $(version)

# Database Console Access

db-console-core: ## Connect to core database console
	@docker-compose exec postgres-core psql -U user -d core_db

db-console-kyc: ## Connect to KYC database console
	@docker-compose exec postgres-kyc psql -U user -d kyc_db

db-console-rates: ## Connect to rates database console
	@docker-compose exec postgres-rates psql -U user -d rates_db

db-console-audit: ## Connect to audit database console
	@docker-compose exec postgres-audit psql -U user -d audit_db

db-ping: ## Check database connectivity
	@echo "Checking database connections..."
	@docker-compose exec -T postgres-core pg_isready -U user -d core_db && echo "✓ Core DB: OK" || echo "✗ Core DB: FAILED"
	@docker-compose exec -T postgres-kyc pg_isready -U user -d kyc_db && echo "✓ KYC DB: OK" || echo "✗ KYC DB: FAILED"
	@docker-compose exec -T postgres-rates pg_isready -U user -d rates_db && echo "✓ Rates DB: OK" || echo "✗ Rates DB: FAILED"
	@docker-compose exec -T postgres-audit pg_isready -U user -d audit_db && echo "✓ Audit DB: OK" || echo "✗ Audit DB: FAILED"

db-backup: ## Backup database (usage: make db-backup db=core_db)
	@if [ -z "$(db)" ]; then \
		echo "Error: db parameter required. Usage: make db-backup db=core_db"; \
		exit 1; \
	fi
	@timestamp=$$(date +%Y%m%d_%H%M%S); \
	filename="backups/$(db)_$$timestamp.sql"; \
	mkdir -p backups; \
	case $(db) in \
		core_db) container=postgres-core;; \
		kyc_db) container=postgres-kyc;; \
		rates_db) container=postgres-rates;; \
		audit_db) container=postgres-audit;; \
		*) echo "Invalid db. Choose: core_db, kyc_db, rates_db, or audit_db"; exit 1;; \
	esac; \
	docker-compose exec -T $$container pg_dump -U user $(db) > $$filename; \
	echo "Backup created: $$filename"

db-restore: ## Restore database (usage: make db-restore db=core_db file=backup.sql)
	@if [ -z "$(db)" ] || [ -z "$(file)" ]; then \
		echo "Error: db and file parameters required"; \
		echo "Usage: make db-restore db=core_db file=backup.sql"; \
		exit 1; \
	fi
	@if [ ! -f "$(file)" ]; then \
		echo "Error: File $(file) not found"; \
		exit 1; \
	fi
	@case $(db) in \
		core_db) container=postgres-core;; \
		kyc_db) container=postgres-kyc;; \
		rates_db) container=postgres-rates;; \
		audit_db) container=postgres-audit;; \
		*) echo "Invalid db. Choose: core_db, kyc_db, rates_db, or audit_db"; exit 1;; \
	esac; \
	docker-compose exec -T $$container psql -U user -d $(db) < $(file); \
	echo "Database restored from: $(file)"

db-schema: ## Export database schema (usage: make db-schema db=core_db)
	@if [ -z "$(db)" ]; then \
		echo "Error: db parameter required. Usage: make db-schema db=core_db"; \
		exit 1; \
	fi
	@case $(db) in \
		core_db) container=postgres-core;; \
		kyc_db) container=postgres-kyc;; \
		rates_db) container=postgres-rates;; \
		audit_db) container=postgres-audit;; \
		*) echo "Invalid db. Choose: core_db, kyc_db, rates_db, or audit_db"; exit 1;; \
	esac; \
	docker-compose exec -T $$container pg_dump -U user --schema-only $(db)

# Docker Management

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	docker-compose build

docker-up: ## Start all Docker services
	@echo "Starting Docker services..."
	docker-compose up -d
	@echo "Services started!"
	@echo "Backend API: http://localhost:8080"
	@echo "Frontend: http://localhost:3000"
	@echo "pgAdmin: http://localhost:5050"
	@echo "Prometheus: http://localhost:9090"
	@echo "Grafana: http://localhost:3001"

docker-down: ## Stop all Docker services
	@echo "Stopping Docker services..."
	docker-compose down

docker-logs: ## View Docker logs (usage: make docker-logs service=backend)
	@if [ -z "$(service)" ]; then \
		docker-compose logs -f; \
	else \
		docker-compose logs -f $(service); \
	fi

docker-restart: ## Restart Docker services
	@echo "Restarting Docker services..."
	docker-compose restart

docker-ps: ## Show running Docker containers
	@docker-compose ps

# Installation helpers

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@echo "Installing air (hot reload)..."
	go install github.com/air-verse/air@latest
	@echo "Installing goimports..."
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "Installing goose (migration tool)..."
	go install github.com/pressly/goose/v3/cmd/goose@latest
	@echo ""
	@echo "Note: golangci-lint installation instructions: https://golangci-lint.run/usage/install/"
	@echo "Tools installed!"
