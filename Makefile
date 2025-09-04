SHELL := /bin/sh

# Config
PROJECT ?= dnd

.PHONY: up down up-python logs logs-go seed-dump seed-restore seed-dump-dir seed-restore-dir mongo-sh editor-sh build build-editor build-parser build-python env-init lint lint-go format test test-go test-integration benchmark help

# Go services (default)
up:
	docker compose up -d mongo editor parser

# Python services (legacy)
up-python:
	docker compose --profile python up -d mongo editor-python parser-python

down:
	docker compose down

logs:
	docker compose logs -f editor parser mongo

logs-go:
	docker compose logs -f editor parser mongo

logs-python:
	docker compose --profile python logs -f editor-python parser-python mongo

# Seed helpers (run inside the mongo container)
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

mongo-sh:
	docker compose exec mongo sh

editor-sh:
	docker compose exec editor sh

# Build images
build:
	docker compose build editor parser

build-python:
	docker compose --profile python build editor-python parser-python

build-editor:
	docker compose build editor

build-parser:
	docker compose build parser

# Initialize .env from example (no overwrite)
env-init:
	@test -f .env || cp .env.example .env || true
	@echo "Loaded .env (or created from .env.example)."

# Go tools
lint-go:
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		go fmt ./...; \
	fi

test-go:
	go test ./...

test-integration:
	@echo "Running Go integration tests..."
	./test_go_integration.sh

benchmark:
	@echo "Running Go performance benchmarks..."
	go test -bench=. -benchmem ./tests/benchmarks

# Python tools (legacy)
lint:
	@if command -v ruff >/dev/null 2>&1; then \
		ruff check editor srd_parser; \
	elif command -v pyflakes >/dev/null 2>&1; then \
		pyflakes editor srd_parser; \
	else \
		echo "No linter found (install ruff or pyflakes)"; \
	fi

format:
	@if command -v black >/dev/null 2>&1; then \
		black editor srd_parser; \
	else \
		echo "Black not found; install black to format"; \
	fi

test:
	python test_basic_integration.py

# Show available commands
help:
	@echo "D&D 5e SRD - Available Commands:"
	@echo ""
	@echo "Docker Services (Go by default):"
	@echo "  make up                    # Start MongoDB + Editor + Parser (Go)"
	@echo "  make up-python             # Start MongoDB + Editor + Parser (Python legacy)"
	@echo "  make down                  # Stop all services"
	@echo "  make logs                  # View Go service logs"
	@echo "  make logs-python           # View Python service logs"
	@echo ""
	@echo "Build Commands:"
	@echo "  make build                 # Build Go images"
	@echo "  make build-python          # Build Python images (legacy)"
	@echo "  make build-editor          # Build only editor image"
	@echo "  make build-parser          # Build only parser image"
	@echo ""
	@echo "Development Commands:"
	@echo "  make lint-go               # Lint Go code"
	@echo "  make test-go               # Run Go unit tests"
	@echo "  make test-integration      # Run integration tests"
	@echo "  make benchmark             # Run performance benchmarks"
	@echo "  make lint                  # Lint Python code (legacy)"
	@echo "  make test                  # Run Python tests (legacy)"
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
	@echo "Build Images:"
	@echo "  make build                 # Build Python editor and parser images"
	@echo "  make build-go              # Build Go editor and parser images"
	@echo "  make build-editor          # Build only Python editor image"
	@echo "  make build-parser          # Build only Python parser image"
	@echo ""
	@echo "Development:"
	@echo "  make env-init              # Initialize .env from .env.example"
	@echo "  make lint                  # Run ruff (preferred) or pyflakes"
	@echo "  make format                # Run black formatter"
	@echo "  make help                  # Show this help message"
	@echo ""
	@echo "URLs:"
	@echo "  Editor: http://localhost:8000/"
	@echo "  Parser: http://localhost:8100/"
