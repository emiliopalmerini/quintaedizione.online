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
- MongoDB running locally or via `docker-compose up -d`
- templ CLI: `go install github.com/a-h/templ/cmd/templ@latest`

## Environment

Copy `.env.example` to `.env` and configure:
- `MONGO_URI` - MongoDB connection string
- `DB_NAME` - Database name (default: dnd)
- `PORT` - Server port (default: 8000)

## Architecture

Hexagonal architecture with clear separation:

```
cmd/viewer/          Entry point
internal/
  adapters/          External interfaces
    repositories/    Repository implementations
    web/             HTTP handlers, middleware
  application/       Business logic
    events/          Event handling
    filters/         Content filtering
    parsers/         Markdown parsing
    services/        Application services
  domain/            Core domain models
    collections/     Collection definitions
    filters/         Domain filter types
    repositories/    Repository interfaces
  infrastructure/    Technical concerns
    config/          Configuration
    database/        Database setup, indexes
    mongodb/         MongoDB client
pkg/                 Shared packages
  mappers/           Data mappers
  mongodb/           MongoDB utilities
  templates/         Template engine
web/
  static/            Static assets
  templates/         Templ template files (.templ)
```

## Tech Stack

- **Framework**: gin-gonic/gin
- **Templates**: a-h/templ
- **Database**: MongoDB (mongo-driver)
- **Markdown**: gomarkdown/markdown
- **UI**: due-draghi-design-system (embedded CSS)

## Content Pipeline

Markdown files in `data/ita/lists/` are parsed at startup and stored in MongoDB collections:
- incantesimi (spells)
- mostri (monsters)
- classi (classes)
- backgrounds
- equipaggiamenti (equipment)
- oggetti_magici (magic items)
- armi (weapons)
- armature (armor)
- talenti (feats)
- and more...

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
