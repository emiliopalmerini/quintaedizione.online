# ADR-009: Map Domain Model & Repository

## Status

Accepted

## Context

We want to add a searchable gallery of commercial-licensed maps from Dyson Logos, translated to Italian. The first step is defining the domain model and repository layer so the rest of the feature can build on a stable foundation.

Maps need two organizational dimensions:
- **Category**: broad classification (e.g. dungeon, caverna, città, foresta)
- **Tag**: granular labels for finer filtering (e.g. trappola, acqua, multilivello, boss)

## Decision

### Domain Entity: `Mappa`

```go
type Mappa struct {
    Slug        string   // URL-safe identifier
    Nome        string   // Italian translated name
    NomeOriginale string // Original English name
    Immagine    string   // Image file path (relative to static assets)
    Categoria   string   // Single category
    Tag         []string // Multiple tags
    Autore      string   // Attribution (e.g. "Dyson Logos")
    Licenza     string   // License type
    URLOriginale string  // Link to original source
}
```

### Repository Interface

```go
type MappaRepository interface {
    FindAll() []Mappa
    FindBySlug(slug string) (Mappa, error)
    FindByCategoria(categoria string) []Mappa
    FindByTag(tag string) []Mappa
    Categorie() []string   // All distinct categories
    Tags() []string        // All distinct tags
}
```

### In-Memory Implementation

- JSON file in `data/mappe/mappe.json` embedded via `embed.FS`
- Loaded into an in-memory store at startup, same pattern as existing collections
- Slug is derived from the Italian name if not provided

## Inputs

- `data/mappe/mappe.json`: array of map objects with fields matching the domain entity

## Outputs

- `internal/domain/maps/mappa.go`: domain entity
- `internal/domain/repositories/mappa_repository.go`: repository interface
- `internal/adapters/repositories/inmemory/mappa_repository.go`: in-memory implementation

## Edge Cases

- Duplicate slug → log warning at startup, keep first occurrence
- Missing `categoria` → default to "senza_categoria"
- Empty `tag` array → valid, map simply has no tags
- Missing `immagine` → skip entry, log warning

## Error Conditions

- JSON file fails to parse → fatal at startup (same as existing collections)
- `FindBySlug` with unknown slug → return error

## Consequences

- New domain package `maps` separate from existing SRD collections
- Follows the same embedded JSON + in-memory store pattern, keeping architecture consistent
- Category and tag indexes built at load time for fast filtering
