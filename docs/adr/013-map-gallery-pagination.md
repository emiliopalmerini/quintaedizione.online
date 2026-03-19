# ADR-013: Map Gallery Pagination

## Status

Accepted

## Context

The gallery page (ADR-010) renders all 437 map cards at once, loading ~670 MB of images. Mobile browsers crash with out-of-memory errors. We need to paginate the gallery to limit initial load.

## Decision

### Load More Pattern

- Default page size: 40 maps
- HTMX "Load More" button appends the next page of cards
- Filter/search changes reset to page 1

### Domain Changes

- `SearchFilters` gains `Offset int` and `Limit int` fields (`Limit=0` means no limit)
- `Repository.Search` returns `([]Mappa, int)` — page slice + total matching count
- `GalleryData` gains `Offset`, `Limit`, `HasMore` fields

### HTMX Flow

1. **Filter change** (offset=0): form `hx-get` replaces `#mappe-grid` with first page + Load More button
2. **Load More** (offset>0): button `hx-get` replaces itself (`hx-swap="outerHTML"`) with next batch of cards + new button
3. **Full page load**: renders complete page with first 40 cards

### URL State

- `hx-push-url` only on filter changes (offset=0), not on Load More
- Bookmarked URLs always load page 1 of the filtered view

## Inputs

- `SearchFilters` with `Offset` and `Limit`
- Query param `offset` from Load More button URL

## Outputs

- `GalleryGrid` template: full grid container (filter changes + initial load)
- `GalleryCards` template: cards fragment (load-more responses)
- `LoadMoreButton` template: self-replacing HTMX button

## Edge Cases

- Offset beyond total results: return empty slice, total unchanged
- Last page has fewer than `Limit` items: no Load More button rendered
- Filter change while scrolled down: grid resets to top with page 1

## Error Conditions

- Invalid offset param: default to 0

## Consequences

- Mobile browsers no longer crash
- Initial page load is fast (40 cards instead of 437)
- No JavaScript changes needed — pure HTMX
- Stats banner moves inside `#mappe-grid` so it updates on filter swaps
