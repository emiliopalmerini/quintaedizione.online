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

Hexagonal architecture with embedded JSON data and in-memory store (no external database):

```
cmd/viewer/          Entry point
data/ita/json/       Embedded JSON data files (source of truth)
internal/
  adapters/          External interfaces
    repositories/
      inmemory/      In-memory repository implementations
    web/             HTTP handlers, middleware
  application/       Business logic
    filters/         Predicate-based content filtering
    parsers/         Markdown rendering, keyword linking
    services/        Application services
  domain/            Core domain models
    collections/     Collection definitions
    filters/         Domain filter types
    repositories/    Repository interfaces (read-only)
  infrastructure/    Technical concerns
    config/          Configuration
    datastore/       JSON loader + in-memory store
pkg/                 Shared packages
  mappers/           Data mappers
  templates/         Template engine
web/
  static/            Static assets
  templates/         Templ template files (.templ)
```

## Tech Stack

- **Framework**: gin-gonic/gin
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
go test -v ./internal/domain/...
```

## CI Workflows

- **Test**: Runs on push/PR to main/develop - templ generate, format check, tests
- **Build and Push**: Builds Docker image on push to main
- **Check Rebase**: Ensures PR branches are rebased on main
