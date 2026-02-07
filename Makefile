SHELL := /bin/sh

PROJECT ?= dnd

RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

.PHONY: help setup build run templ-generate test clean pdf-setup pdf-convert

.DEFAULT_GOAL := help

setup: templ-generate
	@echo -e "$(GREEN)Setup completed!$(NC)"

build: format
	@echo -e "$(GREEN)Build completed!$(NC)"

run: setup build
	@echo -e "$(BLUE)Running application...$(NC)"
	go run ./cmd/viewer/main.go

templ-generate:
	@echo -e "$(BLUE)Generating Templ templates...$(NC)"
	@command -v templ >/dev/null 2>&1 || (echo -e "$(RED)Error: templ is not installed. Run 'go install github.com/a-h/templ/cmd/templ@latest'$(NC)" && exit 1)
	@if [ ! -d "web/templates" ]; then \
		echo -e "$(RED)Error: web/templates directory not found$(NC)"; \
		exit 1; \
	fi
	cd web/templates && templ generate
	@echo -e "$(GREEN)Templ templates generated successfully!$(NC)"

format:
	@echo -e "$(BLUE)Formatting Go code...$(NC)"
	go fmt ./...
	@echo -e "$(GREEN)Code formatted successfully!$(NC)"

test: format
	@echo -e "$(BLUE)Running Go unit tests...$(NC)"
	go test -v ./...
	@echo -e "$(GREEN)Unit tests completed!$(NC)"

clean:
	@echo -e "$(BLUE)Cleaning up Go build artifacts...$(NC)"
	go clean
	@echo -e "$(GREEN)Cleanup completed!$(NC)"

pdf-setup:
	@echo -e "$(BLUE)Setting up PDF pipeline environment...$(NC)"
	@command -v uv >/dev/null 2>&1 || (echo -e "$(RED)Error: uv is not installed. Visit https://docs.astral.sh/uv/$(NC)" && exit 1)
	@if [ ! -d "scripts/pdf-pipeline" ]; then \
		echo -e "$(RED)Error: scripts/pdf-pipeline directory not found$(NC)"; \
		exit 1; \
	fi
	cd scripts/pdf-pipeline && uv sync
	@echo -e "$(GREEN)PDF pipeline setup completed!$(NC)"

pdf-convert: pdf-setup
	@if [ -z "$(PDF)" ]; then \
		echo -e "$(RED)Error: PDF argument required. Usage: make pdf-convert PDF=path/to/file.pdf$(NC)"; \
		exit 1; \
	fi
	@echo -e "$(BLUE)Running PDF conversion pipeline...$(NC)"
	cd scripts/pdf-pipeline && uv run pdf-pipeline run ../../$(PDF) --output-dir ../../data/ita/lists/
	@echo -e "$(GREEN)PDF conversion completed!$(NC)"

help:
	@echo -e "$(BLUE) quintaedizione.online - Available Commands:$(NC)"
	@echo ""
	@echo -e "$(YELLOW)Build Commands:$(NC)"
	@echo "  make setup                 # Setup: generate Templ templates"
	@echo "  make build                 # Build: format Go code"
	@echo "  make run                   # Setup, build, and run application"
	@echo "  make templ-generate        # Generate Go code from Templ templates"
	@echo "  make clean                 # Clean Go build artifacts"
	@echo ""
	@echo -e "$(YELLOW)Go Development:$(NC)"
	@echo "  make format                # Format Go code"
	@echo "  make test                  # Run Go unit tests"
	@echo ""
	@echo -e "$(YELLOW)PDF Pipeline:$(NC)"
	@echo "  make pdf-setup             # Setup PDF pipeline environment (uv sync)"
	@echo "  make pdf-convert PDF=<path> # Convert PDF to markdown (includes setup)"
