SHELL := /bin/sh

# Configuration
PROJECT ?= dnd
GO_VERSION ?= 1.24
DOCKER_COMPOSE := docker compose
MONGO_USER := admin
MONGO_PASS := password
MONGO_DB := $(PROJECT)

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

.PHONY: help up down restart logs build build-viewer templ-generate env-init lint format test test-integration benchmark seed-dump seed-restore seed-dump-dir seed-restore-dir mongo-sh viewer-sh cli-parser cli-build cli-install clean status health check-deps install-deps

# === Docker Services ===
up: check-deps
	@echo -e "$(BLUE)Starting MongoDB and Viewer services...$(NC)"
	$(DOCKER_COMPOSE) up -d mongo viewer
	@echo -e "$(GREEN)Services started successfully!$(NC)"
	@echo -e "$(YELLOW)Viewer: http://localhost:8000/$(NC)"

down:
	@echo -e "$(BLUE)Stopping all services...$(NC)"
	$(DOCKER_COMPOSE) down
	@echo -e "$(GREEN)All services stopped.$(NC)"

restart: down up

logs:
	@echo -e "$(BLUE)Following viewer logs...$(NC)"
	$(DOCKER_COMPOSE) logs -f viewer

status:
	@echo -e "$(BLUE)Service status:$(NC)"
	$(DOCKER_COMPOSE) ps

# === Build Commands ===
build: templ-generate
	@echo -e "$(BLUE)Building viewer image...$(NC)"
	$(DOCKER_COMPOSE) build viewer
	@echo -e "$(GREEN)Build completed successfully!$(NC)"

build-viewer: templ-generate
	@echo -e "$(BLUE)Building viewer image...$(NC)"
	$(DOCKER_COMPOSE) build viewer
	@echo -e "$(GREEN)Viewer build completed!$(NC)"

clean:
	@echo -e "$(BLUE)Cleaning up Docker resources...$(NC)"
	$(DOCKER_COMPOSE) down --volumes --remove-orphans
	docker system prune -f
	@echo -e "$(GREEN)Cleanup completed!$(NC)"



# === Template Generation ===
templ-generate: check-deps
	@echo -e "$(BLUE)Generating Templ templates...$(NC)"
	@if [ ! -d "web/templates" ]; then \
		echo -e "$(RED)Error: web/templates directory not found$(NC)"; \
		exit 1; \
	fi
	cd web/templates && templ generate
	@echo -e "$(GREEN)Templ templates generated successfully!$(NC)"

# === Development Setup ===
env-init:
	@echo -e "$(BLUE)Initializing environment...$(NC)"
	@test -f .env || cp .env.example .env || true
	@echo -e "$(GREEN)Environment initialized (.env file ready)$(NC)"

check-deps:
	@echo -e "$(BLUE)Checking dependencies...$(NC)"
	@command -v docker >/dev/null 2>&1 || (echo -e "$(RED)Error: Docker is not installed$(NC)" && exit 1)
	@command -v docker compose >/dev/null 2>&1 || command -v docker-compose >/dev/null 2>&1 || (echo -e "$(RED)Error: Docker Compose is not installed$(NC)" && exit 1)
	@command -v go >/dev/null 2>&1 || (echo -e "$(RED)Error: Go is not installed$(NC)" && exit 1)
	@command -v templ >/dev/null 2>&1 || (echo -e "$(YELLOW)Warning: templ is not installed. Run 'make install-deps' to install it$(NC)")

install-deps:
	@echo -e "$(BLUE)Installing Go dependencies...$(NC)"
	go mod download
	@echo -e "$(BLUE)Installing templ...$(NC)"
	go install github.com/a-h/templ/cmd/templ@latest
	@echo -e "$(GREEN)Dependencies installed successfully!$(NC)"

# === Go Development ===
lint:
	@echo -e "$(BLUE)Running Go linting...$(NC)"
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo -e "$(BLUE)Running golangci-lint...$(NC)"; \
		golangci-lint run; \
	else \
		echo -e "$(YELLOW)golangci-lint not found, running go fmt...$(NC)"; \
		go fmt ./...; \
	fi
	@echo -e "$(GREEN)Linting completed!$(NC)"

format:
	@echo -e "$(BLUE)Formatting Go code...$(NC)"
	go fmt ./...
	@echo -e "$(GREEN)Code formatted successfully!$(NC)"

test:
	@echo -e "$(BLUE)Running Go unit tests...$(NC)"
	go test -v ./...
	@echo -e "$(GREEN)Unit tests completed!$(NC)"

test-integration:
	@echo -e "$(BLUE)Running integration tests...$(NC)"
	@if [ -f "./test_go_integration.sh" ]; then \
		./test_go_integration.sh; \
	else \
		echo -e "$(RED)Error: test_go_integration.sh not found$(NC)"; \
		exit 1; \
	fi
	@echo -e "$(GREEN)Integration tests completed!$(NC)"

benchmark:
	@echo -e "$(BLUE)Running performance benchmarks...$(NC)"
	@if [ -d "./tests/benchmarks" ]; then \
		go test -bench=. -benchmem ./tests/benchmarks; \
	else \
		echo -e "$(YELLOW)Warning: ./tests/benchmarks directory not found$(NC)"; \
	fi

health:
	@echo -e "$(BLUE)Checking service health...$(NC)"
	@curl -f http://localhost:8000/health 2>/dev/null && echo -e "$(GREEN)Viewer service is healthy$(NC)" || echo -e "$(RED)Viewer service is not responding$(NC)"

# === Database Management ===
seed-dump:
	@echo -e "$(BLUE)Creating database backup...$(NC)"
	@TIMESTAMP=$$(date +%Y%m%d_%H%M%S); \
	BACKUP_FILE="$(PROJECT)_backup_$$TIMESTAMP.archive.gz"; \
	$(DOCKER_COMPOSE) exec mongo mongodump --username=$(MONGO_USER) --password=$(MONGO_PASS) --authenticationDatabase=admin --db $(MONGO_DB) --gzip --archive=/tmp/$(PROJECT).archive.gz && \
	docker cp $$($(DOCKER_COMPOSE) ps -q mongo):/tmp/$(PROJECT).archive.gz ./$$BACKUP_FILE && \
	echo -e "$(GREEN)Backup created: $$BACKUP_FILE$(NC)" || \
	echo -e "$(RED)Backup failed!$(NC)"

seed-restore:
	@if [ -z "$(FILE)" ]; then \
		echo -e "$(YELLOW)Usage: make seed-restore FILE=backup_file.archive.gz$(NC)"; \
		echo -e "$(BLUE)Available backups:$(NC)"; \
		ls -1 $(PROJECT)_backup_*.archive.gz 2>/dev/null || echo -e "$(YELLOW)No backup files found$(NC)"; \
	else \
		echo -e "$(BLUE)Restoring from $(FILE)...$(NC)"; \
		if [ ! -f "$(FILE)" ]; then \
			echo -e "$(RED)Error: File $(FILE) not found$(NC)"; \
			exit 1; \
		fi; \
		docker cp $(FILE) $$($(DOCKER_COMPOSE) ps -q mongo):/tmp/restore.archive.gz && \
		$(DOCKER_COMPOSE) exec mongo mongorestore --username=$(MONGO_USER) --password=$(MONGO_PASS) --authenticationDatabase=admin --gzip --archive=/tmp/restore.archive.gz --drop && \
		echo -e "$(GREEN)Restore completed from $(FILE)$(NC)" || \
		echo -e "$(RED)Restore failed!$(NC)"; \
	fi

seed-dump-dir:
	@echo -e "$(BLUE)Dumping database to directory format...$(NC)"
	$(DOCKER_COMPOSE) exec mongo sh -lc 'rm -rf /seed/dump && mkdir -p /seed/dump && mongodump --username=$(MONGO_USER) --password=$(MONGO_PASS) --authenticationDatabase=admin --db $(MONGO_DB) --out /seed/dump'
	@echo -e "$(GREEN)Directory dump completed!$(NC)"

seed-restore-dir:
	@echo -e "$(BLUE)Restoring from directory format...$(NC)"
	$(DOCKER_COMPOSE) exec mongo sh -lc 'test -d /seed/dump && mongorestore --username=$(MONGO_USER) --password=$(MONGO_PASS) --authenticationDatabase=admin --dir /seed/dump --drop || echo "No /seed/dump directory found"'
	@echo -e "$(GREEN)Directory restore completed!$(NC)"

# === Container Access ===
mongo-sh:
	@echo -e "$(BLUE)Accessing MongoDB container...$(NC)"
	$(DOCKER_COMPOSE) exec mongo sh

viewer-sh:
	@echo -e "$(BLUE)Accessing Viewer container...$(NC)"
	$(DOCKER_COMPOSE) exec viewer sh

# === CLI Parser ===
cli-build:
	@echo -e "$(BLUE)Building CLI parser...$(NC)"
	@mkdir -p bin
	go build -o bin/cli-parser ./cmd/cli-parser
	@echo -e "$(GREEN)CLI parser built successfully!$(NC)"

cli-install: cli-build
	@echo -e "$(BLUE)Installing CLI parser to /usr/local/bin...$(NC)"
	sudo mkdir -p /usr/local/bin
	sudo cp bin/cli-parser /usr/local/bin/dnd-parser
	@echo -e "$(GREEN)CLI parser installed as 'dnd-parser'$(NC)"

cli-parser: cli-build
	@echo -e "$(BLUE)Running CLI parser...$(NC)"
	./bin/cli-parser $(ARGS)


# === Help ===
help:
	@echo -e "$(BLUE)D&D 5e SRD - Available Commands:$(NC)"
	@echo ""
	@echo -e "$(YELLOW)Docker Services:$(NC)"
	@echo "  make up                    # Start MongoDB + Viewer services"
	@echo "  make down                  # Stop all services"
	@echo "  make restart               # Restart all services"
	@echo "  make logs                  # View viewer service logs"
	@echo "  make status                # Show service status"
	@echo "  make health                # Check service health"
	@echo ""
	@echo -e "$(YELLOW)Build Commands:$(NC)"
	@echo "  make build                 # Build viewer image (with templ generation)"
	@echo "  make build-viewer          # Build only viewer image (with templ generation)"
	@echo "  make templ-generate        # Generate Go code from Templ templates"
	@echo "  make clean                 # Clean up Docker resources"
	@echo ""
	@echo -e "$(YELLOW)Go Development:$(NC)"
	@echo "  make lint                  # Lint Go code (vet + golangci-lint)"
	@echo "  make format                # Format Go code"
	@echo "  make test                  # Run Go unit tests"
	@echo "  make test-integration      # Run integration tests"
	@echo "  make benchmark             # Run performance benchmarks"
	@echo ""
	@echo -e "$(YELLOW)Database Management:$(NC)"
	@echo "  make seed-dump             # Create timestamped database backup"
	@echo "  make seed-restore FILE=x   # Restore database from backup file"
	@echo "  make seed-dump-dir         # Dump to directory format"
	@echo "  make seed-restore-dir      # Restore from directory format"
	@echo ""
	@echo -e "$(YELLOW)Container Access:$(NC)"
	@echo "  make mongo-sh              # Access MongoDB container shell"
	@echo "  make viewer-sh             # Access Viewer container shell"
	@echo ""
	@echo -e "$(YELLOW)CLI Parser:$(NC)"
	@echo "  make cli-build             # Build CLI parser binary"
	@echo "  make cli-install           # Install CLI parser system-wide"
	@echo "  make cli-parser ARGS='...' # Run CLI parser with arguments"
	@echo ""
	@echo -e "$(YELLOW)Setup & Dependencies:$(NC)"
	@echo "  make env-init              # Initialize .env from .env.example"
	@echo "  make check-deps            # Check required dependencies"
	@echo "  make install-deps          # Install Go dependencies"
	@echo "  make help                  # Show this help message"
	@echo ""
	@echo -e "$(YELLOW)URLs:$(NC)"
	@echo -e "  Viewer: $(GREEN)http://localhost:8000/$(NC)"