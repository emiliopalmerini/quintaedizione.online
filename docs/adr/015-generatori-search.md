# ADR-015: Generatori Table Search

## Status

Accepted

## Context

The generatori home page displays 64 tables organized in 8 groups. Users need to quickly find a specific table by name or description without scrolling through all groups.

## Decision

### Text Search

- Search bar at top of generatori home page, above the group cards
- Case-insensitive substring match on table `Name` and `Description` fields
- HTMX-powered: `hx-get="/generatori/"` with `?q=` param, debounced keyup (200ms)
- Partial response: HTMX requests return just the grid, full requests return the whole page

### Application Layer

- Add `SearchGroups(query string) []domain.Group` to `application.Service`
- Empty query returns all groups (same as `ListGroups()`)
- Filters tables within each group; omits groups with zero matching tables
- Preserves group ordering from `GroupRegistry`

### URL State

- Query reflected in `?q=...` param
- `hx-push-url="true"` for shareable URLs and back-button support

### Template Structure

- Extract group grid into `GeneratoriGrid(groups, query)` component (like mappe's `GalleryGrid`)
- Search input inside `<form>` with HTMX attributes
- Empty state: "Nessun generatore trovato" with clear filters link

## Inputs

- `application.Service` table list
- User input: search text (`q` query param)

## Outputs

- `SearchGroups` method on Service
- Updated `handleHome` handler accepting `?q=` and returning HTMX partial
- Updated `home.templ` with search input and extractable grid

## Edge Cases

- Empty query string → show all groups (unfiltered)
- No matching tables → empty state with "Cancella filtri" link
- Query matches description but not name → table still shown
- Group with all tables filtered out → group hidden entirely

## Error Conditions

- None expected — in-memory substring match cannot fail

## Consequences

- Follows established mappe search pattern (ADR-011)
- No new dependencies or packages
- Search stays within the generatori vertical slice
