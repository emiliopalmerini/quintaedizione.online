SHELL := /bin/sh

# Config
PROJECT ?= dnd

.PHONY: up down logs seed-dump seed-restore seed-dump-dir seed-restore-dir mongo-sh editor-sh build build-editor build-parser env-init lint format

up:
	docker compose up -d mongo editor srd-parser

down:
	docker compose down

logs:
	docker compose logs -f editor srd-parser mongo

# Seed helpers (run inside the mongo container)
seed-dump:
	docker compose exec mongo sh -lc 'mkdir -p /seed && mongodump --db $(PROJECT) --gzip --archive=/seed/$(PROJECT).archive.gz'

seed-restore:
	docker compose exec mongo sh -lc 'test -f /seed/$(PROJECT).archive.gz && mongorestore --gzip --archive=/seed/$(PROJECT).archive.gz --drop || echo "No /seed/$(PROJECT).archive.gz found"'

seed-dump-dir:
	docker compose exec mongo sh -lc 'rm -rf /seed/dump && mkdir -p /seed/dump && mongodump --db $(PROJECT) --out /seed/dump'

seed-restore-dir:
	docker compose exec mongo sh -lc 'test -d /seed/dump && mongorestore --dir /seed/dump --drop || echo "No /seed/dump directory found"'

mongo-sh:
	docker compose exec mongo sh

editor-sh:
	docker compose exec editor sh

# Build images
build:
	docker compose build editor srd-parser

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
