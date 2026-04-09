# ADR-022: Multi-Version Document Switcher

## Status

Accepted

## Context

The site serves D&D SRD content from multiple editions (5e and 5.5e). Each edition's version of a document (spell, monster, etc.) is stored as a separate entry keyed by `{source_short}/{slug}` and served at its own URL: `/srd/{collection}/{source}/{slug}`. However, users have no way to discover that a document exists in another edition, or to compare versions. The only cross-edition feature is the edition filter on collection list pages.

The Notion task "Documento singolo multi-versione (preferita/piu usata + ultima)" calls for a unified document page that lets users switch between editions of the same content.

## Decision

Add a version switcher (tab bar) to the item detail page. Each edition keeps its own URL; tabs navigate between them.

### Inputs

- `collection` (string): path parameter, e.g. "incantesimi"
- `source` (string): path parameter, edition short name, e.g. "5.5e"
- `slug` (string): path parameter, document ID, e.g. "palla-di-fuoco"

### Outputs

- The existing item detail page, augmented with a version switcher when the same slug exists in more than one source
- The switcher renders between the title row and the article content

### Behavior

1. When the item detail handler receives a request for `/{collection}/{source}/{slug}`, it looks up all documents in `collection` that share the same `slug` across sources.
2. If more than one version exists:
   - A `<nav class="version-switcher">` tab bar renders below the title, with one tab per edition.
   - The tab matching the current `source` is marked active (`aria-selected="true"`).
   - Inactive tabs are links to the other edition's URL, using HTMX (`hx-get`, `hx-target="#page-root"`, `hx-select="#page-root"`, `hx-swap="outerHTML"`, `hx-push-url="true"`) for seamless in-page switching.
3. If only one version exists, no switcher is rendered (current behavior preserved).
4. Prev/next navigation continues to work within the current sorted document list (unchanged).

### Data Flow

```
Store.GetBySlug(collection, slug)
  -> []map[string]any (all docs sharing this slug)
  -> Repository.FindVersions(collection, slug) -> []VersionInfo
  -> ContentService.GetItemVersions(collection, slug) -> []VersionInfo
  -> Handler builds []VersionTab (URL, label, active state)
  -> Template renders VersionSwitcher component
```

### New Types

- `domain.VersionInfo{SourceShort string, CompositeID string}` - value object
- `repositories.DocumentVersions` interface - new port (follows ISP pattern alongside DocumentReader, DocumentStatistics, DocumentNavigation)
- `models.VersionTab{SourceShort, URL, IsCurrent, Label string}` - view model

### Store Change

Add a secondary index `slugIndex map[string]map[string][]string` (collection -> bare slug -> []compositeIDs) built at startup in `NewStore`. Add `GetBySlug(collection, slug string) []map[string]any` for O(1) lookups.

### Edge Cases

- **Single source:** `GetBySlug` returns one entry; handler sets `Versions` to nil; template renders nothing new.
- **Slug exists in only one source despite multi-source mode:** same as single source; no switcher.
- **Legacy URL** (`/{collection}/{slug}`): existing redirect to `/{collection}/{defaultSource}/{slug}` lands on the default edition with the switcher visible.
- **Document title differs between editions:** each tab shows the edition label (e.g. "5.5e"), not the title. The page title updates when switching.

### Error Conditions

- `GetItemVersions` fails: handler logs a warning and renders the page without the switcher (graceful degradation). The primary content still displays correctly.
- `GetBySlug` returns empty for a slug that was successfully fetched by `Get`: should not happen since the same slug must exist in the index. If it does, treat as single-version (no switcher).

## Consequences

- Users can discover and compare editions of the same content
- Each edition retains its own URL (SEO preserved, deep links work)
- No changes to URL routing, data loading, search, or collection list behavior
- Negligible memory overhead: one extra map entry per document for the slug index
- The switcher degrades gracefully to invisible when irrelevant
