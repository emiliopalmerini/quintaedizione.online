SHELL := /bin/sh

# Configuration
PROJECT ?= dnd

.PHONY: help up down logs build build-editor build-parser env-init lint format test test-integration benchmark seed-dump seed-restore seed-dump-dir seed-restore-dir mongo-sh editor-sh

# === Docker Services ===
up:
	docker compose up -d mongo editor parser

down:
	docker compose down

logs:
	docker compose logs -f editor parser

# === Build Commands ===
build:
	docker compose build editor parser

build-editor:
	docker compose build editor

build-parser:
	docker compose build parser

# === Development Setup ===
env-init:
	@test -f .env || cp .env.example .env || true
	@echo "Loaded .env (or created from .env.example)."

# === Go Development ===
lint:
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		go fmt ./...; \
	fi

format:
	go fmt ./...

test:
	go test ./...

test-integration:
	@echo "Running integration tests..."
	./test_go_integration.sh

benchmark:
	@echo "Running performance benchmarks..."
	go test -bench=. -benchmem ./tests/benchmarks

# === Database Management ===
seed-dump:
	@echo "Creating database backup..."
	docker compose exec mongo mongodump --username=admin --password=password --authenticationDatabase=admin --db $(PROJECT) --gzip --archive=/tmp/$(PROJECT).archive.gz
	docker cp $$(docker compose ps -q mongo):/tmp/$(PROJECT).archive.gz ./$(PROJECT)_backup_$$(date +%Y%m%d_%H%M%S).archive.gz
	@echo "Backup created: $(PROJECT)_backup_$$(date +%Y%m%d_%H%M%S).archive.gz"

seed-restore:
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make seed-restore FILE=backup_file.archive.gz"; \
		echo "Available backups:"; \
		ls -1 $(PROJECT)_backup_*.archive.gz 2>/dev/null || echo "No backup files found"; \
	else \
		echo "Restoring from $(FILE)..."; \
		docker cp $(FILE) $$(docker compose ps -q mongo):/tmp/restore.archive.gz; \
		docker compose exec mongo mongorestore --username=admin --password=password --authenticationDatabase=admin --gzip --archive=/tmp/restore.archive.gz --drop; \
		echo "Restore completed from $(FILE)"; \
	fi

seed-dump-dir:
	docker compose exec mongo sh -lc 'rm -rf /seed/dump && mkdir -p /seed/dump && mongodump --username=admin --password=password --authenticationDatabase=admin --db $(PROJECT) --out /seed/dump'

seed-restore-dir:
	docker compose exec mongo sh -lc 'test -d /seed/dump && mongorestore --username=admin --password=password --authenticationDatabase=admin --dir /seed/dump --drop || echo "No /seed/dump directory found"'

# === Container Access ===
mongo-sh:
	docker compose exec mongo sh

editor-sh:
	docker compose exec editor sh

# === Help ===
help:
	@echo "D&D 5e SRD - Available Commands:"
	@echo ""
	@echo "Docker Services:"
	@echo "  make up                    # Start MongoDB + Editor + Parser"
	@echo "  make down                  # Stop all services"
	@echo "  make logs                  # View service logs"
	@echo ""
	@echo "Build Commands:"
	@echo "  make build                 # Build editor and parser images"
	@echo "  make build-editor          # Build only editor image"
	@echo "  make build-parser          # Build only parser image"
	@echo ""
	@echo "Go Development:"
	@echo "  make lint                  # Lint Go code"
	@echo "  make format                # Format Go code"
	@echo "  make test                  # Run Go unit tests"
	@echo "  make test-integration      # Run integration tests"
	@echo "  make benchmark             # Run performance benchmarks"
	@echo ""
	@echo "Database Management:"
	@echo "  make seed-dump             # Create timestamped database backup"
	@echo "  make seed-restore FILE=x   # Restore database from backup file"
	@echo "  make seed-dump-dir         # Dump to directory format"
	@echo "  make seed-restore-dir      # Restore from directory format"
	@echo ""
	@echo "Container Access:"
	@echo "  make mongo-sh              # Access MongoDB shell"
	@echo "  make editor-sh             # Access Editor container shell"
	@echo ""
	@echo "Setup:"
	@echo "  make env-init              # Initialize .env from .env.example"
	@echo "  make help                  # Show this help message"
	@echo ""
	@echo "URLs:"
	@echo "  Editor: http://localhost:8000/"
	@echo "  Parser: http://localhost:8100/"