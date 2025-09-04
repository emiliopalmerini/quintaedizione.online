# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a D&D 5e SRD (System Reference Document) viewer and parser system with two main components:
- **Editor**: FastAPI web application with HTMX for viewing SRD content
- **SRD Parser**: Web-based parser for ingesting SRD content into MongoDB

The application uses Italian for UI text and documentation but code is in English.

## Development Commands

### Docker Setup (Primary Development Method)
```bash
make up                    # Start MongoDB + Editor + SRD Parser
make down                  # Stop all services
make logs                  # View logs from all services
make build                 # Build both editor and parser images
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

### Code Quality
```bash
make lint                  # Run ruff (preferred) or pyflakes
make format                # Run black formatter
```

### Local Development (without Docker)
**Editor:**
```bash
cd editor
pip install -r requirements.txt
export MONGO_URI="mongodb://admin:password@localhost:27017/?authSource=admin"
export DB_NAME="dnd"
uvicorn main:app --reload --port 8000
```

**SRD Parser:**
```bash
cd srd_parser
pip install -r requirements.txt
# Same MONGO_URI and DB_NAME as above
python web.py  # or uvicorn web:app --reload --port 8100
```

## Architecture

### Editor Application (`/editor`)
- **Framework**: FastAPI + Jinja2 templates + HTMX for progressive enhancement
- **Database**: MongoDB via Motor (async driver)
- **Styling**: Tailwind CSS (CDN) + custom CSS
- **Structure**:
  - `main.py`: FastAPI app setup
  - `routers/pages.py`: Main routes and view logic
  - `core/`: Database connections, template environment, utilities
  - `templates/`: Jinja2 templates with HTMX integration
  - `services/`: Business logic services

### SRD Parser Application (`/srd_parser`)
- **Framework**: FastAPI with web interface for parsing operations
- **Architecture**: Hexagonal Architecture with Domain-Driven Design
- **Parser Structure**:
  - `domain/`: Domain entities, value objects, and services (DDD implementation)
  - `parsers/*.py`: Domain-specific parsers (spells, monsters, classes, etc.)
  - `work.py`: Configuration of collections and source files
  - `application/`: Service layer with ingest runner and service
  - `adapters/`: MongoDB persistence adapter and external interfaces
  - `web.py`: FastAPI web interface (primary)
  - `web_hexagonal.py`: Alternative hexagonal architecture web interface
- **Key Features**:
  - Dry-run mode for analysis without database writes
  - Upsert operations for data ingestion
  - Rich domain model with proper validation
  - Class parser generates structured data (`features_by_level`, `spellcasting_progression`)

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