SHELL := /bin/sh

# Config
PROJECT ?= dnd

.PHONY: up down logs seed-dump seed-restore seed-dump-dir seed-restore-dir mongo-sh editor-sh tui tui-local tui-up tui-build

up:
	docker compose up -d mongo editor

down:
	docker compose down

logs:
	docker compose logs -f editor mongo

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

# TUI helpers
tui:
	# Run the TUI inside Docker (interactive)
	docker compose run --rm srd-tui

tui-up:
	# Bring up Mongo and run the TUI
	docker compose up -d mongo
	docker compose run --rm srd-tui

tui-local:
	# Run the TUI locally with your Python
	python -m srd_parser.tui || python3 -m srd_parser.tui

tui-build:
	# Rebuild the srd-tui image to pick up code changes
	docker compose build srd-tui
