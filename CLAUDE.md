# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a D&D 5e SRD (System Reference Document) viewer and parser system with two main components:
- **Editor**: Go web application with Gin and HTMX for viewing SRD content
- **SRD Parser**: Go web-based parser for ingesting SRD content into MongoDB

The system has been migrated from Python to Go for better performance and maintainability. The application uses Italian for UI text and documentation but code is in English.

**Note**: Python versions are still available as legacy components but Go versions are now the default.

## Development Commands

### Docker Setup (Primary Development Method)
```bash
make up                    # Start MongoDB + Editor + Parser (Go)
make up-python             # Start MongoDB + Editor + Parser (Python legacy)
make down                  # Stop all services
make logs                  # View Go service logs
make build                 # Build Go images
make build-python          # Build Python images (legacy)
make build-editor          # Build only editor image
make build-parser          # Build only parser image
```

### Database Management
```bash
make seed-dump             # Create database backup to /seed/dnd.archive.gz
make seed-restore          # Restore database from backup
make seed-dump-dir         # Dump to directory format
make seed-restore-dir      # Restore from directory format
```

### Code Quality & Testing
```bash
# Go (primary)
make lint-go               # Lint Go code with go vet and golangci-lint
make test-go               # Run Go unit tests
make test-integration      # Run integration tests with Go services
make benchmark             # Run performance benchmarks

# Python (legacy)
make lint                  # Run ruff (preferred) or pyflakes
make format                # Run black formatter
make test                  # Run Python integration tests
```

### Local Development (without Docker)

**Go Editor:**
```bash
export MONGO_URI="mongodb://admin:password@localhost:27017/?authSource=admin"
export DB_NAME="dnd"
export PORT="8000"
go run cmd/editor/main.go
```

**Go Parser:**
```bash
export MONGO_URI="mongodb://admin:password@localhost:27017/?authSource=admin"
export DB_NAME="dnd"
export PORT="8100"
go run cmd/parser/main.go
```

**Python Editor (legacy):**
```bash
cd editor
pip install -r requirements.txt
export MONGO_URI="mongodb://admin:password@localhost:27017/?authSource=admin"
export DB_NAME="dnd"
uvicorn main:app --reload --port 8000
```

**Python SRD Parser (legacy):**
```bash
cd srd_parser
pip install -r requirements.txt
# Same MONGO_URI and DB_NAME as above
python web.py  # or uvicorn web:app --reload --port 8100
```

## Architecture

### Go Editor Application (`/cmd/editor`, `/internal`, `/web`)
- **Framework**: Gin + Go templates + HTMX for progressive enhancement
- **Database**: MongoDB via mongo-go-driver
- **Styling**: Tailwind CSS (CDN) + custom CSS (preserved from Python version)
- **Performance**: Built-in metrics collection, caching, and monitoring
- **Structure**:
  - `cmd/editor/main.go`: Application entry point with graceful shutdown
  - `internal/adapters/web/`: HTTP handlers and routing
  - `internal/application/services/`: Business logic services with caching
  - `internal/infrastructure/`: Configuration, performance monitoring, caching
  - `pkg/mongodb/`: MongoDB client wrapper with common operations
  - `pkg/templates/`: Template engine with helper functions
  - `web/templates/`: Go templates with HTMX integration (converted from Jinja2)
  - `web/static/`: CSS and static assets

### Go SRD Parser Application (`/cmd/parser`, `/internal/application/parsers`)
- **Framework**: Gin with web interface for parsing operations
- **Architecture**: Hexagonal Architecture with Domain-Driven Design (preserved)
- **Parser Structure**:
  - `internal/domain/`: Domain entities and value objects
  - `internal/application/parsers/*.go`: Italian-only parsers (spells, monsters, classes, etc.)
  - `internal/application/parsers/work.go`: Configuration of collections and source files
  - `internal/application/services/`: Service layer with ingest operations
  - `internal/adapters/mongodb/`: MongoDB persistence adapter
- **Key Features**:
  - Italian-only content parsing (as requested)
  - Upsert operations for data ingestion
  - Hexagonal architecture with proper separation of concerns
  - All original parser functionality preserved

### Legacy Python Applications (available via `make up-python`)
- **Editor** (`/editor`): Original FastAPI + Jinja2 implementation
- **Parser** (`/srd_parser`): Original FastAPI parser implementation
- **Note**: Maintained for reference but Go versions are recommended

### Shared Domain (`/shared_domain`)
- **Purpose**: Domain entities and value objects shared between editor and parser
- **Structure**: Common business logic and domain rules
- **Benefits**: Consistency across applications, single source of truth for domain concepts

### Database Schema
- **MongoDB Database**: `dnd` (configurable via `DB_NAME`)
- **Collections**: spells, magic_items, monsters, classes, backgrounds, etc.
- **Authentication**: Uses `admin/password` credentials (dev environment)

## Development Guidelines

### Code Conventions
- Python: Use existing patterns from the codebase
- Templates: Follow HTMX patterns for progressive enhancement
- Database: Use Motor (async) for editor, PyMongo (sync) for parser
- Error Handling: Keep error messages generic in UI, detailed logs in backend

### Testing
- **Integration tests**: Located in root directory (`test_basic_integration.py`, `test_curl_integration.sh`, `test_domain_model.py`)
- **Editor tests**: Located in `editor/tests/`
- **Parser tests**: Located in `srd_parser/tests/`
- Run with pytest (configure via `pytest.ini`)
- Integration tests require running services (MongoDB, Editor, Parser)

### Deployment Considerations
- Uses Docker Compose for local development
- MongoDB runs with authentication enabled
- Environment variables: `MONGO_URI`, `DB_NAME`, `INPUT_DIR`, `SOURCE_LABEL`
- Ports: Editor (8000), Parser (8100), MongoDB (27017)

### Commit Conventions
Use Conventional Commits format: feat, fix, docs, chore, build, refactor, perf, test

## Key Files to Know

- `Makefile`: All development commands
- `docker-compose.yml`: Service definitions
- `editor/main.py`: Editor app entry point
- `srd_parser/web.py`: Parser web interface
- `srd_parser/work.py`: Parser configuration
- Template files use Italian for UI text

## URLs
- Editor: http://localhost:8000/
- Parser Web UI: http://localhost:8100/