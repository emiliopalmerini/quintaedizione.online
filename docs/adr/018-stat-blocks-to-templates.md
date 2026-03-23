# ADR-018: Move Stat-Blocks From Go Builders to Templ Templates

## Status

Accepted

## Context

All stat-block HTML is constructed via `fmt.Fprintf` in `html_builders.go` (481 lines of Go code building HTML strings). Each entity type (spell, monster, class, species, background) has a builder function that produces HTML at data-load time.

Problems:
- HTML structure is hardcoded in Go — changing styles requires recompilation
- Duplicated patterns across builders (subtitle, properties, description divs)
- No type-safety on the HTML (just string concatenation)
- Content is rendered once at startup and stored as strings — the template engine has no control over the HTML structure
- Both HTML and markdown versions must be maintained in sync (dual builders)

## Decision

### Store Structured Data, Render at Request Time

Instead of pre-rendering HTML at load time, store structured entity data in the `map[string]any` documents and render via templ templates at request time.

### Phase 1: Entity-Specific View Models

Create typed view models for each entity type in `internal/srd/web/models/`:

```go
type SpellViewModel struct {
    Level          int
    School         string
    CastingTime    string
    Range          string
    Components     string
    Duration       string
    Ritual         bool
    Description    template.HTML  // pre-rendered HTML with crosslinks
    AtHigherLevels template.HTML
    Classes        []string
}
```

The `Description` field stays as pre-rendered HTML (from the markdown/crosslink pipeline). Only the stat-block structure moves to templates.

### Phase 2: Templ Components

Create collection-specific templ components:

```
web/templates/
  stat_blocks/
    spell.templ      → SpellStatBlock(spell SpellViewModel)
    monster.templ     → MonsterStatBlock(monster MonsterViewModel)
    class.templ       → ClassStatBlock(class ClassViewModel)
    species.templ     → SpeciesStatBlock(species SpeciesViewModel)
    background.templ  → BackgroundStatBlock(bg BackgroundViewModel)
    shared.templ      → StatBlockProperty(label, value), StatBlockDivider(), etc.
```

### Phase 3: Update the Loader

The loader stores entity fields directly in `map[string]any` instead of pre-built HTML:

```go
// Before
doc["content"] = l.buildSpellHTML(s)

// After
doc["level"] = s.Level
doc["school"] = s.School
doc["casting_time"] = s.CastingTime
// ...
doc["description_html"] = l.renderContent(s.Description.toMarkdown(src))
```

The `content` field is removed. The web layer's mapper constructs the view model and the template renders the stat-block.

### raw_content Stays

`raw_content` (markdown for copy-to-clipboard) continues to be built at load time by the markdown builders. This doesn't change.

## Inputs

- Current `html_builders.go` builder functions
- Current `map[string]any` document format

## Outputs

- Entity-specific view models in `web/models/`
- Templ stat-block components in `web/templates/stat_blocks/`
- Shared stat-block components (property, divider, section)
- Updated loader: stores structured fields instead of pre-built HTML
- Updated mapper: builds view models from structured fields
- Removed: `buildSpellHTML`, `buildMonsterHTML`, `buildClassHTML`, `buildSpeciesHTML`

## Edge Cases

- Existing `content` field consumers (search indexing, SEO) need to use the structured fields or render on-demand
- Cache strategy may need adjustment since HTML is now rendered per-request (templ is fast, but measure)
- Background entity already uses a unified markdown builder — straightforward migration

## Error Conditions

- Template rendering errors → return 500 with error page (same as current template errors)
- Missing fields in document → template uses zero values (same as current builder behavior)

## Consequences

- HTML structure is now in templates — designers can modify without Go changes
- Type-safe rendering via templ (compile-time errors for mismatched fields)
- Shared components reduce duplication (property, divider, section patterns)
- Slight increase in request-time work (rendering templates instead of serving pre-built strings) — expected to be negligible given templ's performance
- Markdown builders for raw_content remain unchanged
