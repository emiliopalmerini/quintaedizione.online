# quintaedizione.online

D&D 5e content website (Quinta Edizione online) - Italian translation of SRD content.

## Build & Run

```bash
# Setup and run (generates templ templates, formats code, runs app)
make run

# Generate templ templates only
make templ-generate

# Format code
make format

# Run tests
make test

# Clean build artifacts
make clean
```

## Prerequisites

- Go 1.25.2
- templ CLI: `go install github.com/a-h/templ/cmd/templ@latest`

## Environment

Copy `.env.example` to `.env` and configure:
- `PORT` - Server port (default: 8000)

## Architecture

Vertical-slice hexagonal architecture with embedded JSON data and in-memory store (no external database). Each bounded context is a self-contained slice with its own domain, application, infrastructure, and web layers:

```
cmd/app/               Entry point (wires all slices)
data/ita/json/         Embedded JSON data files (source of truth)
internal/
  srd/                 SRD content viewer slice
    domain/            Core domain models, ports, value objects
      collections/     Collection definitions & registry
      filters/         Filter types & predicates
      repositories/    Repository interfaces (read-only)
      search/          Search types & interfaces
    application/       Business logic
      filters/         Predicate builder, filter registry
      parsers/         Markdown rendering, glossary linking
      search/          Fuzzy search service & index
      services/        Content & filter orchestration
    infrastructure/    Technical concerns
      config/          Collection metadata loader
      datastore/       JSON loader + in-memory store
      persistence/     Repository implementations
    web/               HTTP handlers, middleware, mappers
      config/          Cache config
      display/         Collection display strategies
      dto/             Data transfer objects
      mappers/         Document mapper
      models/          View models
  combattimenti/       Encounter calculator slice
    domain/            Encounter & monster entities
    application/       XP calculation, queries
    infrastructure/    Persistence & HTTP handlers
  mappe/               Map gallery slice
    domain/            Map entity & repository interface
    infrastructure/    Map repository
    web/               Gallery handlers & templates
  infrastructure/      Shared cross-cutting concerns
    config.go          Server configuration
    cache.go           In-memory cache implementation
pkg/                   Shared packages
  mappers/             Data mappers
  templates/           Template engine
web/
  static/              Static assets
  templates/           Templ template files (.templ)
```

## Tech Stack

- **Framework**: net/http (stdlib)
- **Templates**: a-h/templ
- **Data**: Embedded JSON via `embed.FS` + in-memory store
- **Markdown**: gomarkdown/markdown
- **UI**: due-draghi-design-system (embedded CSS)

## Content Pipeline

```
PDF → Python parser (scripts/parse_srd/) → JSON → Embedded in Go binary → In-memory store
```

JSON files in `data/ita/json/` are embedded at compile time and loaded into an in-memory store at startup. The Python PDF parser extracts structured data from the SRD PDF into JSON. The Go loader maps JSON fields (English) to Italian collection field names and renders markdown descriptions to HTML.

Collections: incantesimi, mostri, animali, classi, backgrounds, equipaggiamenti, oggetti_magici, armi, armature, strumenti, cavalcature_veicoli, servizi, talenti, regole

## Testing

```bash
# Run all tests
make test

# Run with race detector
go test -race ./...

# Run specific package tests
go test -v ./internal/srd/domain/...
```

## CI Workflows

- **Test**: Runs on push/PR to main/develop - templ generate, format check, tests
- **Build and Push**: Builds Docker image on push to main
- **Check Rebase**: Ensures PR branches are rebased on main
