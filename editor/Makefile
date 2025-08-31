# Makefile for D&D 5e SRD Editor

.PHONY: help install test test-unit test-integration test-e2e test-coverage lint format type-check clean dev

# Default target
help:
	@echo "Available targets:"
	@echo "  install          Install dependencies"
	@echo "  install-test     Install test dependencies"
	@echo "  test             Run all tests"
	@echo "  test-unit        Run unit tests only"
	@echo "  test-integration Run integration tests only"
	@echo "  test-e2e         Run e2e tests only"
	@echo "  test-coverage    Run tests with coverage report"
	@echo "  lint             Run code linting"
	@echo "  format           Format code"
	@echo "  type-check       Run type checking"
	@echo "  clean            Clean build artifacts"
	@echo "  dev              Start development server"

# Installation
install:
	pip install -r requirements.txt

install-test: install
	pip install -r requirements-test.txt

# Testing
test: install-test
	pytest

test-unit: install-test
	pytest tests/unit -v

test-integration: install-test
	pytest tests/integration -v -m "not slow"

test-e2e: install-test
	pytest tests/e2e -v

test-coverage: install-test
	pytest --cov=. --cov-report=html --cov-report=term-missing

test-slow: install-test
	pytest -m slow -v

# Code quality
lint: install-test
	flake8 . --count --select=E9,F63,F7,F82 --show-source --statistics
	flake8 . --count --exit-zero --max-complexity=10 --max-line-length=127 --statistics

format: install-test
	black .
	isort .

type-check: install-test
	mypy . --ignore-missing-imports

# Development
dev:
	uvicorn main:app --reload --host 0.0.0.0 --port 8000

dev-debug:
	uvicorn main:app --reload --host 0.0.0.0 --port 8000 --log-level debug

# Cleanup
clean:
	find . -type f -name "*.pyc" -delete
	find . -type d -name "__pycache__" -delete
	find . -type d -name "*.egg-info" -exec rm -rf {} +
	rm -rf htmlcov/
	rm -rf .coverage
	rm -rf .pytest_cache/
	rm -rf dist/
	rm -rf build/

# CI/CD helpers
ci-test: install-test
	pytest --junitxml=test-results.xml --cov=. --cov-report=xml

pre-commit-install: install-test
	pre-commit install

pre-commit-run: install-test
	pre-commit run --all-files