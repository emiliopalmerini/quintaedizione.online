# ADR-010: Map Gallery Page

## Status

Accepted

## Context

With the map domain model in place (ADR-009), we need a page to browse the map collection as a visual gallery.

## Decision

### Route

`GET /strumenti/mappe` — gallery page showing all maps as a responsive grid of cards.

### Card Layout

```
┌──────────────────┐
│                  │
│   [thumbnail]    │
│                  │
├──────────────────┤
│ Nome della Mappa │
│ Categoria        │
│ tag1 · tag2      │
└──────────────────┘
```

- Thumbnail: lazy-loaded image, fixed aspect ratio container
- Name: Italian name, clickable (links to detail page)
- Category: displayed as subtle label
- Tags: displayed as small inline badges

### Responsive Grid

- Desktop: 4 columns
- Tablet: 3 columns
- Mobile: 2 columns
- CSS grid with gap, consistent with design system

### Navigation

- Add "Mappe" entry to site navigation/tools section

## Inputs

- `MappaRepository.FindAll()` for the full list
- Thumbnail images from static assets

## Outputs

- `web/templates/mappe/gallery.templ`: gallery page template
- `internal/adapters/web/handlers/mappe_handler.go`: HTTP handler
- Route registration in router

## Edge Cases

- No maps in data → show empty state message
- Broken thumbnail image → show placeholder with map name
- Large number of maps (>100) → pagination (reuse existing pagination pattern)

## Error Conditions

- Repository unavailable → 500 error page

## Consequences

- New route namespace `/strumenti/mappe` for map tooling
- New handler and template files following existing patterns
