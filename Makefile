SHELL := /bin/sh

# Config
PROJECT ?= dnd

.PHONY: up down up-go down-go logs seed-dump seed-restore seed-dump-dir seed-restore-dir mongo-sh editor-sh build build-editor build-parser build-go env-init lint format help

up:
	docker compose up -d mongo editor srd-parser

# Go services
up-go:
	docker compose --profile editor-go --profile parser-go up -d mongo editor-go srd-parser-go

down:
	docker compose down

down-go:
	docker compose --profile editor-go --profile parser-go down

logs:
	docker compose logs -f editor srd-parser mongo

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
	docker compose build editor srd-parser

build-go:
	docker compose --profile editor-go --profile parser-go build editor-go srd-parser-go

build-editor:
	docker compose build editor

build-parser:
	docker compose build srd-parser

# Initialize .env from example (no overwrite)
env-init:
	@test -f .env || cp .env.example .env || true
	@echo "Loaded .env (or created from .env.example)."

# Lint/format helpers (optional: ruff/black if installed; fallback: pyflakes)
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

# Show available commands
help:
	@echo "D&D 5e SRD - Available Commands:"
	@echo ""
	@echo "Docker Services:"
	@echo "  make up                    # Start MongoDB + Editor + SRD Parser (Python)"
	@echo "  make up-go                 # Start MongoDB + Editor + SRD Parser (Go)"
	@echo "  make down                  # Stop Python services"
	@echo "  make down-go               # Stop Go services"
	@echo "  make logs                  # View logs from all services"
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
