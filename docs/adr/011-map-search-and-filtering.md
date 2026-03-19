# ADR-011: Map Search & Filtering

## Status

Proposed

## Context

The gallery page (ADR-010) displays all maps. Users need to search by name and filter by category and tag to find relevant maps quickly.

## Decision

### Text Search

- Search bar at top of gallery, searches Italian map names
- Case-insensitive substring match
- HTMX-powered: `hx-get` with debounced input triggers partial gallery update

### Category Filter

- Dropdown or chip bar listing all distinct categories
- Single-select: one category at a time (or "Tutte" for all)
- Chips preferred if categories are few (<10), dropdown otherwise

### Tag Filter

- Multi-select tag chips below the category filter
- Multiple tags can be active simultaneously (AND logic: map must have all selected tags)
- Horizontally scrollable on mobile

### URL State

- All filters reflected in query params: `?q=...&categoria=...&tag=...`
- Shareable, back-button friendly
- HTMX partial updates with `hx-push-url`

### Filter Bar Layout

```
[🔍 Cerca mappa...                    ]
[Tutte] [Dungeon] [Caverna] [Città] ...
[acqua] [trappola] [multilivello] ...
```

## Inputs

- `MappaRepository` query methods
- User input: search text, selected category, selected tags

## Outputs

- Search input component in gallery template
- Category chip/filter component
- Tag multi-select component
- Updated handler to accept and apply filter query params

## Edge Cases

- Search + filters combined → all conditions applied (AND)
- No results → "Nessuna mappa trovata" with clear filters button
- Category with 0 maps after tag filter → still shown but dimmed or hidden
- Empty search string → show all (filtered by category/tag if set)

## Error Conditions

- Invalid query param values → ignore, show unfiltered results

## Consequences

- Consistent with existing collection filtering UX patterns
- HTMX partial updates keep the page responsive
- URL state enables sharing filtered views
