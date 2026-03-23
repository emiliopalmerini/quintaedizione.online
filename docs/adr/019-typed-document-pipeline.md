# ADR-019: Typed Document Pipeline

## Status

Accepted

## Context

The entire pipeline from loader to web layer passes data as `map[string]any`. The domain has a `Document` struct, but it's rarely used — most code paths work with raw maps:

- `Loader` produces `[]map[string]any`
- `Store` holds `map[string]map[string]map[string]any`
- `DocumentRepository` exposes `FindMapByID` and `FindMaps` (raw maps)
- `ContentService` returns `map[string]any` to handlers
- `DocumentMapper` manually extracts fields from maps with type assertions
- `persistence/document_repository.go` converts maps to `Document` in `FindByID`/`FindAll` but these typed methods are barely used

This means:
- No compile-time safety on field access
- Silent bugs when field names change or types mismatch
- Every layer does its own `doc["field"].(string)` with `ok` checks
- The domain model is decorative — the real data model is the map

## Decision

### Keep `map[string]any` for the Store Layer

The in-memory store remains generic (`map[string]any`). It's a general-purpose key-value store that doesn't need to know about entity types. This is appropriate for its role.

### Typed Conversion at the Repository Boundary

The `DocumentRepository` becomes the boundary where `map[string]any` is converted to typed domain objects. Remove `FindMapByID` and `FindMaps` — all queries return typed results.

### Richer Domain Document

Extend `Document` with the fields that are currently only in maps:

```go
type Document struct {
    ID              DocumentID
    Title           string
    Source          string       // source short name
    Collection      string       // collection name
    Content         HTMLContent
    RawContent      MarkdownContent
    Fields          map[string]any  // collection-specific fields (livello, scuola, etc.)
}
```

The `Fields` map holds collection-specific data (spell level, monster CR, etc.) that varies per entity type. This is a pragmatic middle ground — we get type safety on the common fields while keeping flexibility for collection-specific ones.

### Updated Service Layer

`ContentService` returns `*Document` (or `[]Document`) instead of `map[string]any`. The web mapper works with `Document` instead of raw maps.

### Migration Path

1. Add `Source` and `Collection` fields to `Document`
2. Update `DocumentRepository` implementations to populate the new fields
3. Deprecate `FindMapByID` and `FindMaps` — add typed alternatives
4. Update `ContentService` to return `*Document`
5. Update `DocumentMapper` to accept `*Document`
6. Remove deprecated map-based methods

## Inputs

- Current `Document` struct (5 fields)
- Current `map[string]any` flow throughout the pipeline

## Outputs

- Extended `Document` struct with `Source`, `Collection`, `Fields`
- Updated `DocumentRepository` interface — typed query methods
- Updated `ContentService` returning typed documents
- Updated `DocumentMapper` accepting typed documents
- Removed: `FindMapByID`, `FindMaps`, direct map access in web layer

## Edge Cases

- Collection-specific fields (spell level, monster CR) remain in `Fields` map — not fully typed
- Filter/aggregation operations still need raw field access → `Fields` map preserves this
- Search indexing works on `Title` + `RawContent` — already typed
- Display strategies extract fields for list views → work with `Fields` map

## Error Conditions

- Field not found in `Fields` map → zero value (same behavior as current map access, but explicit)
- Type mismatch in `Fields` → caught at repository boundary during conversion

## Consequences

- Compile-time safety on common fields (ID, Title, Source, Content)
- Single conversion point (repository) instead of scattered type assertions
- Web layer works with typed data — fewer runtime errors
- `Fields` map is a pragmatic escape hatch for collection-specific data
- Prepares for ADR-018 (typed view models can be built from typed documents)
