# ADR-017: Separate Loader Responsibilities

## Status

Accepted

## Context

`loader.go` (~785 lines) has too many responsibilities:

1. JSON deserialization (struct definitions + `readJSON`)
2. Source discovery and metadata loading (`LoadAll`, `loadSource`)
3. Field mapping (English JSON keys → Italian collection field names)
4. Segment-to-markdown conversion (`toMarkdown`, placeholder mechanism)
5. Markdown-to-HTML rendering (`renderContent`, `renderContentInline`)
6. HTML stat-block construction (`buildSpellHTML`, `buildMonsterHTML`, etc.)
7. Raw markdown construction (`buildSpellMarkdown`, `buildMonsterMarkdown`, etc.)
8. Source tagging (`tagDoc`)

This makes the loader hard to test in isolation and tightly couples data loading with rendering.

## Decision

### Split Into Three Concerns

#### 1. `loader.go` — JSON Loading and Source Discovery

Keeps only:
- `Loader` struct, `NewLoader`, `LoadAll`, `loadSource`, `loadSourceData`
- `readJSON`, `tagDoc`
- `loadSpells`, `loadMonsters`, etc. — but these only deserialize JSON and build `map[string]any` with raw fields
- Calls a `ContentRenderer` interface for `content` and `raw_content` fields

#### 2. `content.go` — Segment Types and Markdown Conversion

Moves here from `loader.go`:
- `jsonSegment`, `jsonContent` types
- `toMarkdown(sourceShort)`, `plainText()`
- `segmentCollection`, `segmentTooltipCategory` maps (until ADR-016 moves them to CrossLinker)

This is a pure data transformation layer with no dependencies on rendering.

#### 3. `html_builders.go` — Stat-Block HTML Construction (unchanged for now)

Remains as-is. ADR-018 will address moving stat-blocks to templates.

### ContentRenderer Interface

The loader delegates rendering via an interface rather than holding a direct `*parsers.MarkdownRenderer`:

```go
// ContentRenderer converts markdown content to rendered HTML.
type ContentRenderer interface {
    RenderContent(markdown string) string
    RenderContentInline(markdown string) string
}
```

This decouples the loader from the rendering implementation and makes testing easier (mock the renderer).

## Inputs

- Current `loader.go` (785 lines)
- Current `html_builders.go` (481 lines)

## Outputs

- `loader.go` (~300 lines): source discovery, JSON loading, document assembly
- `content.go` (~100 lines): segment types, markdown conversion
- `html_builders.go` (~450 lines): stat-block HTML construction (unchanged)
- `ContentRenderer` interface

## Edge Cases

- No behavior change — this is a pure refactor
- All existing tests must pass unchanged

## Error Conditions

- None — same error paths as before

## Consequences

- Easier to test loading logic without a real renderer
- Clear separation between "read data" and "transform data"
- Prepares for ADR-016 (CrossLinker can own segment type configuration)
- Prepares for ADR-018 (stat-blocks move to templates)
