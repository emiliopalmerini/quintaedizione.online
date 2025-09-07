# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Parser Architecture

The parser follows Clean Architecture principles with Strategy + Registry patterns:

**Domain Layer** (`internal/domain/`):
- Contains ONLY pure business entities (Incantesimo, Mostro, Documento, etc.)
- Domain interfaces that other layers implement (e.g., ParserRepository)
- NO parsing logic or implementation patterns

**Application Layer** (`internal/application/parsers/`):
- ParsingStrategy interface and implementations
- ParserRegistry for managing strategies
- Content type definitions and mappings
- All concrete parser implementations

**Important Rules:**
1. Parsers MUST return domain objects, not `map[string]interface{}`
2. Strategy and Registry patterns belong in APPLICATION layer, not domain
3. Each content type should have its own strategy implementation
4. Use the BaseParser for common functionality

**Adding a New Parser:**
1. Create a new strategy file (e.g., `monsters_strategy.go`)
2. Implement the `ParsingStrategy` interface
3. Return proper domain objects
4. Register in `CreateDefaultRegistry()`

**DO NOT:**
- Put parsing logic in the domain layer
- Return generic maps from parsers
- Mix domain entities with parsing strategies

## Development Commands

**Docker Services (Primary Go Implementation):**
- `make up` - Start MongoDB + Editor (port 8000) + Parser (port 8100)
- `make down` - Stop all services
- `make logs` - View Go service logs
- `make build` - Build Go editor and parser images

**Go Development:**
- `make lint` - Run go vet and golangci-lint (or go fmt as fallback)
- `make test` - Run Go unit tests with `go test ./...`
- `make test-integration` - Run integration tests via `./test_go_integration.sh`
- `make benchmark` - Run performance benchmarks

**Database Management:**
- `make seed-dump` - Create timestamped database backup
- `make seed-restore FILE=backup.archive.gz` - Restore from backup
- `make mongo-sh` - Access MongoDB container shell
- Database: `dnd`, Credentials: `admin:password`

## Architecture

This is a D&D 5e SRD (System Reference Document) management system with two main services:

**Clean Architecture Pattern:**
- `cmd/` - Application entry points (editor, parser)
- `internal/domain/` - Core business logic and entities
- `internal/application/` - Use cases, services, handlers, parsers
- `internal/adapters/` - External interfaces (MongoDB repository, web handlers)
- `internal/infrastructure/` - Configuration and setup
- `internal/shared/` - Common models (BaseEntity, MarkdownContent)
- `pkg/` - Reusable packages (MongoDB client, templates)

**Services:**
1. **Editor** (port 8000) - Web interface for viewing/editing D&D content
2. **Parser** (port 8100) - Processes markdown files from `./data/` into MongoDB

**Data Structure:**
- `data/eng/` - English SRD 5.2 markdown files
- `data/ita/` - Italian SRD markdown files
- MongoDB collections: `documenti`, `classi`, `backgrounds`, `documenti_en`

**Key Technologies:**
- Go 1.24 with Gin web framework
- MongoDB 8 for data storage
- Docker Compose for orchestration
- Template-based web rendering with HTMX + Templ

The system parses D&D content from markdown files into structured MongoDB documents, supporting only Italian SRD content with searchable and renderable formats.
- usa l'italiano per i termini di dominio
- Use any instead of interface
- Use .Contains instead of loop. Like this
```go
     for _, valid := range validTypes {
        if contentType == valid {
            return true
        }
    }
```
Use this:
```go
    return slices.Contains(validTypes, contentType)
```
- Don't call file with name relative to recent command. Name file relative to domain and beheviour. Like: don't use test_improvments but test_parser_strategies