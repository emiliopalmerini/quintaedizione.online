# ADR 0005: Migrazione a Go

Status: Accepted

## Context

Il sistema D&D 5e SRD è diventato più complesso nel tempo e richiede maggiori performance e mantenibilità. L'attuale implementazione in Python, benché funzionale, presenta alcune limitazioni:

- Performance sub-ottimali per operazioni intensive
- Complessità crescente nella gestione delle dipendenze
- Overhead di runtime per un'applicazione che potrebbe beneficiare di compilation ahead-of-time
- Gestione della concorrenza che potrebbe essere migliorata

## Current System Analysis

- **Editor**: FastAPI + HTMX + Jinja2 + MongoDB (async Motor)
- **Parser**: FastAPI + hexagonal architecture + MongoDB (sync PyMongo)
- **Shared Domain**: Domain entities and value objects
- **Architecture**: DDD with hexagonal patterns, event-driven components

## Decisione

Migrare l'intero sistema da Python a Go mantenendo l'architettura esistente e la compatibilità dei dati.

### Go Technology Stack Mapping

| Python Component       | Go Equivalent           |
|------------------------|-------------------------|
| FastAPI                | Gin                     |
| Jinja2 Templates       | templ                   |
| HTMX                   | Same (client-side)      |
| Motor (async MongoDB)  | mongo-go-driver         |
| PyMongo (sync MongoDB) | mongo-go-driver         |
| Pydantic               | Go structs + validation |
| Docker                 | Same                    |

### Proposed Go Architecture

```
cmd/
├── editor/main.go
└── parser/main.go

internal/
├── domain/           # Domain entities, value objects
├── application/      # Use cases, services
├── adapters/         # MongoDB, HTTP handlers
├── infrastructure/   # Configuration, logging
└── shared/          # Shared utilities

pkg/
├── mongodb/         # MongoDB client wrapper
├── validation/      # Struct validation
└── templates/       # Template helpers

web/
├── templates/       # HTML templates
└── static/         # CSS, JS assets
```

### Migration Plan - 4 Phases

**Phase 1: Foundation & Shared Components (Weeks 1-2)**
- Setup Go project structure with modules
- Implement domain entities and value objects in Go
- Create MongoDB connection layer
- Setup Docker containers for Go services
- Migrate shared domain models

**Phase 2: Parser Migration (Weeks 3-4)**
- Port SRD parser logic to Go
- Implement hexagonal architecture patterns
- Migrate all parser modules (spells, monsters, classes, etc.)
- Create Go web interface for parser
- Maintain data compatibility with existing MongoDB schema

**Phase 3: Editor Migration (Weeks 5-6)**
- Port editor HTTP handlers to Go
- Convert Jinja2 templates to Go templates (preserve HTMX)
- Implement query repositories and services
- Migrate admin endpoints and health checks
- Setup cache management and metrics collection

**Phase 4: Integration & Optimization (Weeks 7-8)**
- Performance optimization and benchmarking
- Complete test suite migration
- Update Docker Compose and Makefile
- Documentation updates
- Legacy Python cleanup

### Key Migration Considerations

- **Data Compatibility**: Maintain MongoDB schema compatibility throughout migration
- **Template Strategy**: Convert Jinja2 → Go html/template while preserving HTMX functionality
- **Testing**: Parallel testing during migration to ensure feature parity
- **Deployment**: Blue-green deployment strategy using Docker profiles
- **Performance**: Expect improved performance due to Go's compiled nature and better concurrency

## Conseguenze

**Positive:**
- Migliori performance grazie alla compilazione e alla gestione nativa della concorrenza
- Riduzione del footprint di memoria
- Deployment più semplice con binari statici
- Migliore type safety a compile-time
- Ecosistema Go maturo per applicazioni web e database

**Negative:**
- Sforzo significativo di migrazione (8 settimane stimate)
- Necessità di riapprendimento del team per Go
- Possibili incompatibilità temporanee durante la migrazione
- Rischio di introdurre nuovi bug durante il porting
- Perdita temporale di alcune funzionalità avanzate di Python (es. dynamic typing per casi specifici)