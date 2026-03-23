# ADR-016: Unify Cross-Link System

## Status

Accepted

## Context

The content rendering pipeline has two independent cross-link systems that produce inconsistent HTML output:

1. **Data-driven references** (in `loader.go`): `jsonSegment` types (spell, condition, ability, skill, etc.) are converted to markdown links or Unicode placeholder tooltips at load time. Tooltips produce `<span class="glossary-term" data-term-id data-term-cat data-term-link>`.

2. **Curated glossary** (in `glossary_linker.go`): Regex-based post-processing on rendered HTML. Produces `<span class="glossary-term" data-term-id data-term-def data-term-cat>`.

Problems:
- Same CSS class (`glossary-term`) but different data attributes (`data-term-link` vs `data-term-def`)
- Client-side JS must handle both attribute shapes
- Two separate loading paths and processing phases
- The curated glossary parses HTML into a full DOM tree for every document, then re-renders it
- The placeholder mechanism (U+F8FF) is fragile — it depends on markdown/HTML renderers not consuming Private Use Area characters

## Decision

### Unified CrossLinker

Replace both systems with a single `CrossLinker` in `application/parsers/` that:

1. Accepts a glossary terms list (from `glossary.json`) AND the segment-to-collection mapping (currently hardcoded in `loader.go`)
2. Produces consistent HTML output for all cross-link types

### Consistent HTML Output

All cross-linked terms produce the same `<span>` shape:

```html
<span class="glossary-term"
      data-term-id="accecato"
      data-term-cat="condizione"
      data-term-link="/srd/glossario/5.5e/accecato"
      data-term-def="Non è in grado di vedere..."
      tabindex="0">accecato</span>
```

- `data-term-id`: always present (term identifier)
- `data-term-cat`: always present (Italian category label)
- `data-term-link`: present when the term has a dedicated page (conditions, spells, equipment, rules) — the `<span>` is wrapped in an `<a>` tag
- `data-term-def`: present when a glossary definition exists (from `glossary.json`)

### Processing Flow

The new flow eliminates the placeholder mechanism:

```
jsonContent segments
    ↓
toMarkdown(sourceShort): entity references → markdown links (unchanged)
    ↓
MarkdownRenderer.Render(): markdown → HTML (unchanged)
    ↓
CrossLinker.EnrichTerms(html): finds glossary terms in rendered HTML,
    wraps first occurrence with <span> + data attributes
    (skips inside <a>, <code>, <pre>, <h1-h6>, existing <span>)
```

For segment types that currently use tooltip placeholders (damage_type, creature_type, ability, skill):
- Move them to `segmentCollection` so they become markdown links like spells/conditions
- OR keep them as plain text and let the CrossLinker's glossary matching handle them

### Migration

1. Move `segmentCollection` and `segmentTooltipCategory` maps into the CrossLinker as configuration
2. Remove `resolveTooltipPlaceholders()` and the U+F8FF mechanism from `loader.go`
3. Remove `GlossaryLinker` — its functionality is absorbed into `CrossLinker`
4. Update `MarkdownRenderer` to use `CrossLinker` instead of `GlossaryLinker`

## Inputs

- `glossary.json` terms (id, term, category, definition)
- Segment type → collection mapping
- Segment type → tooltip info mapping (category, rule page)
- Source short name for URL building

## Outputs

- `CrossLinker` struct in `application/parsers/`
- Consistent `<span class="glossary-term">` HTML across all cross-link types
- Removed: `GlossaryLinker`, `resolveTooltipPlaceholders`, tooltip placeholder mechanism

## Edge Cases

- Term appears inside an existing `<a>` tag (data-driven link) → skip, don't double-wrap
- Term appears inside `<code>` or `<pre>` → skip
- Multiple glossary terms in same text node → only first occurrence of each term is linked
- Glossary term overlaps with a data-driven link text → data-driven link takes precedence (it's already an `<a>`)
- Term with both a dedicated page AND a glossary definition → both `data-term-link` and `data-term-def` present

## Error Conditions

- Malformed HTML input → return input unchanged (same as current GlossaryLinker behavior)
- Empty glossary → no-op, pass through
- Regex compilation failure for a term → skip that term (log warning)

## Consequences

- Client-side JS simplified: one attribute shape to handle
- Single point of configuration for all cross-link behavior
- Removes fragile U+F8FF placeholder mechanism
- Glossary definitions available on data-driven terms too (richer tooltips)
