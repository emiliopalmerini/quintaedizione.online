# ADR-012: Map Detail View

## Status

Accepted

## Context

Users need to view a map at full size with attribution and download options. The gallery (ADR-010) links to individual map pages.

## Decision

### Route

`GET /strumenti/mappe/:slug` — detail page for a single map.

### Layout

```
┌─────────────────────────────────────┐
│ ← Torna alla galleria               │
│                                     │
│ Nome della Mappa                    │
│ Categoria · tag1 · tag2 · tag3      │
│                                     │
│ ┌─────────────────────────────────┐ │
│ │                                 │ │
│ │        [full-size map]          │ │
│ │                                 │ │
│ └─────────────────────────────────┘ │
│                                     │
│ Autore: Dyson Logos                 │
│ Licenza: Commercial Use Allowed     │
│ [Scarica originale ↗]              │
└─────────────────────────────────────┘
```

- Full-width map image, zoomable on click/pinch
- Back link to gallery (preserves previous filter state via referrer or query param)
- Attribution block: author name, license info, link to original source
- Download link to original source URL

### SEO

- Page title: `{Nome} - Mappe | Quinta Edizione Online`
- Meta description from category + tags

## Inputs

- `MappaRepository.FindBySlug(slug)`

## Outputs

- `web/templates/mappe/detail.templ`: detail page template
- Handler method for the detail route

## Edge Cases

- Unknown slug → 404 page
- Missing `URLOriginale` → hide download link
- Very large map image → constrain to viewport width, allow zoom

## Error Conditions

- `FindBySlug` returns error → 404 response

## Consequences

- Proper attribution ensures license compliance
- Back-to-gallery link maintains browsing flow
- Slug-based URLs are clean and shareable
