# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Parser Architecture

The parser follows Clean Architecture principles with Strategy + Registry patterns:

**Domain Layer** (`internal/domain/`):
- Contains a single unified `Document` entity with value objects (DocumentID, DocumentFilters, HTMLContent, MarkdownContent)
- Domain repository interfaces
- NO parsing logic or implementation patterns

**Application Layer** (`internal/application/parsers/`):
- `DocumentStrategy` interface for parsing content
- `ParserRegistry` for managing strategies
- `MarkdownRenderer` for converting markdown to HTML
- Content type definitions and mappings
- All concrete parser implementations (*_document_strategy.go)

**Important Rules:**
1. ALL parsers MUST return `domain.Document` entities, not entity-specific types
2. Parsers MUST render markdown content to HTML using `MarkdownRenderer`
3. Strategy and Registry patterns belong in APPLICATION layer, not domain
4. Each content type has its own strategy implementation
5. Use `BaseDocumentParser` for common functionality

**Adding a New Parser:**
1. Create a new strategy file (e.g., `new_type_document_strategy.go`)
2. Implement the `DocumentStrategy` interface
3. Use `BaseDocumentParser.CreateDocument()` to build the Document with HTML rendering
4. Populate `filters` map with collection-specific metadata
5. Register in `CreateDefaultRegistry()`

**DO NOT:**
- Put parsing logic in the domain layer
- Return entity-specific types or generic maps
- Mix domain entities with parsing strategies
- Skip HTML rendering (always use MarkdownRenderer)

## Repository Architecture

The system implements a unified repository using the Repository pattern with factory injection:

**Repository Structure:**
- `internal/adapters/repositories/factory.go` - Repository factory for dependency injection
- `internal/adapters/repositories/mongodb/document_mongo_repository.go` - Single MongoDB repository implementation
- `internal/domain/repositories/document_repository.go` - Repository interface

**Key Changes from Previous Architecture:**
- Replaced 16+ entity-specific repositories with a single `DocumentRepository`
- All entities are stored and retrieved as `domain.Document` instances
- Collection name is passed as a parameter to repository methods
- ~85% code reduction in repository layer

**Repository Interface Methods:**
- `Create(ctx, doc, collection)` - Insert new document
- `Update(ctx, doc, collection)` - Update existing document
- `Delete(ctx, id, collection)` - Delete by ID
- `FindByID(ctx, id, collection)` - Retrieve single document
- `FindAll(ctx, collection, limit)` - Retrieve all documents
- `FindByFilters(ctx, collection, filters, limit)` - Filter-based query
- `UpsertMany(ctx, collection, documents)` - Bulk upsert
- `Count(ctx, collection)` - Count documents

**Repository Pattern Benefits:**
- Clean separation between domain and data access
- Easy testing with mock repositories
- Consistent CRUD operations across all content types
- Type-safe operations using `domain.Document`
- No code changes needed to add new content types

## Development Commands

**Docker Services (Primary Go Implementation):**
- `make up` - Start MongoDB + Quinta Edizione.online Viewer (port 8000)
- `make down` - Stop all services
- `make logs` - View Go service logs
- `make build` - Build Go viewer image

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

This is a D&D 5e SRD (System Reference Document) management system with a web viewer and CLI parser:

**Clean Architecture Pattern:**
- `cmd/` - Application entry points (viewer, cli-parser)
- `internal/domain/` - Core business logic and entities
- `internal/application/` - Use cases, services, handlers, parsers
- `internal/adapters/` - External interfaces (repository factory, MongoDB repositories, web handlers)
  - `repositories/` - Repository interfaces and implementations
    - `factory.go` - Repository factory for dependency injection
    - `mongodb/` - MongoDB-specific repository implementations
- `internal/infrastructure/` - Configuration and setup
- `internal/shared/` - Common models (BaseEntity, MarkdownContent)
- `pkg/` - Reusable packages (MongoDB client, templates)

**Services:**
1. **Quinta Edizione.online Viewer** (port 8000) - Web interface for viewing D&D content
2. **Parser CLI** - Command-line tool that processes markdown files from `./data/` into MongoDB

**Data Structure:**
- `data/ita/lists/` - **Primary parsing source**: Clean list files containing only D&D entities to be parsed
  - Files: `animali.md`, `armi.md`, `armature.md`, `backgrounds.md`, `classi.md`, `equipaggiamenti.md`, `incantesimi.md`, `mostri.md`, `oggetti_magici.md`, `regole.md`, `talenti.md`, etc.
  - Format: H2 headers for each entity, standardized field formatting, clean structure without descriptive sections
- `data/ita/docs/` - **Backup**: Original SRD documentation files (not used for parsing)
- `data/ita/DIZIONARIO_CAMPI_ITALIANI.md` - Italian field terminology dictionary
- MongoDB collections: `animali`, `armi`, `armature`, `backgrounds`, `cavalcature_e_veicoli`, `classi`, `documenti`, `equipaggiamento`, `incantesimi`, `mostri`, `oggetti_magici`, `regole`, `servizi`, `specie`, `strumenti`, `talenti`

**Key Technologies:**
- Go 1.24 with Gin web framework
- MongoDB 8 for data storage
- Docker Compose for orchestration
- Template-based web rendering with HTMX + Templ

The system uses a CLI parser to process D&D content from markdown files in `data/ita/lists/` into structured MongoDB documents, supporting only Italian SRD content with searchable and renderable formats through the web viewer.

**Database Document Structure:**
Each parsed entity follows a consistent document structure that separates metadata from domain data:

```json
{
  "_id": ObjectId("..."),
  "collection": "armature",
  "source_file": "ita/lists/armature.md",
  "locale": "ita",
  "created_at": "2025-01-10T...",
  "contenuto": "**Costo:** 5 mo\n**Peso:** 3,5 kg\n...",
  "value": {
    "nome": "Armatura Imbottita",
    "slug": "armatura-imbottita",
    "categoria": "Leggera",
    "costo": { "valore": 5, "valuta": "mo" },
    "peso": { "valore": 3.5, "unita": "kg" },
    "classe_armatura": { "base": 11, "modificatore_des": true }
  }
}
```

**Document Structure Explained:**
- **Metadata (root level)**: System and operational data
  - `collection`: Target MongoDB collection name
  - `source_file`: Original markdown file path
  - `locale`: Content locale (always "ita")
  - `created_at`: Parse timestamp
  - `contenuto`: Original markdown source for debugging/audit
- **Domain Data (`value` object)**: All business/domain-specific fields
  - `nome`, `slug`: Entity identification
  - All other fields: Domain-specific structured data

This separation allows for:
- Consistent metadata across all collections
- Clean domain data structure
- Full-text search on original content
- Easy indexing and querying strategies

## Document Standards

All files in `data/ita/lists/` follow standardized formatting:

**Header Hierarchy:**
- H1 (`#`) for file title
- H2 (`##`) for each entity entry
- H3 (`###`) for entity subsections (Tratti, Azioni, etc.)

**Field Format:**
- Bold field names followed by colon: `**Campo:** valore`
- Bullet points for monster stats: `- **Campo:** valore`

**Table Format (Monsters/Animals):**
```markdown
| Caratteristica | Valore | Modificatore | Tiro Salvezza |
|----------------|--------|--------------|---------------|
| FOR | 21 | +5 | +5 |
```

**Metadata Format:**
- Spells: `*Livello X Scuola (Classi)*` or `*Trucchetto di Scuola (Classi)*`
- Magic Items: `*Tipo, Rarità (Requisiti)*`
- Monsters/Animals: `*Tipo Taglia, Allineamento*`
- Feats: `*Categoria Talento*` or `*Talento Categoria (Prerequisiti)*`
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