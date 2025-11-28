SHELL := /bin/sh

PROJECT ?= dnd
GO_VERSION ?= 1.24
DOCKER_COMPOSE := docker compose
MONGO_USER := admin
MONGO_PASS := password
MONGO_DB := $(PROJECT)

RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

.PHONY: help logs build templ-generate test clean

.DEFAULT_GOAL := help

build: format templ-generate
	@echo -e "$(BLUE)Building viewer image...$(NC)"
	$(DOCKER_COMPOSE) build viewer
	@echo -e "$(GREEN)Build completed successfully!$(NC)"

clean:
	@echo -e "$(BLUE)Cleaning up Docker resources...$(NC)"
	$(DOCKER_COMPOSE) down --volumes --remove-orphans
	docker system prune -f
	@echo -e "$(GREEN)Cleanup completed!$(NC)"

templ-generate:
	@echo -e "$(BLUE)Generating Templ templates...$(NC)"
	@command -v templ >/dev/null 2>&1 || (echo -e "$(RED)Error: templ is not installed. Run 'make install-deps'$(NC)" && exit 1)
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

help:
	@echo -e "$(BLUE) quintaedizione.online - Available Commands:$(NC)"
	@echo ""
	@echo -e "$(YELLOW)Build Commands:$(NC)"
	@echo "  make build                 # Build viewer image (with templ generation)"
	@echo "  make templ-generate        # Generate Go code from Templ templates"
	@echo "  make clean                 # Clean up Docker resources"
	@echo ""
	@echo -e "$(YELLOW)Go Development:$(NC)"
	@echo "  make format                # Format Go code"
	@echo "  make test                  # Run Go unit tests"
	@echo ""
	@echo -e "$(YELLOW)URLs:$(NC)"
	@echo -e "  Viewer: $(GREEN)http://localhost:8000/$(NC)"
