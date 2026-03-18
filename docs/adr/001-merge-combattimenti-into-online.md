# ADR-001: Merge Combattimenti Into Online as a Single App

## Status

Accepted

## Context

We have two separate Go web applications:

- **quintaedizione.online** — D&D 5e SRD content website (gin + templ, embedded JSON, in-memory store)
- **due-draghi/combattimenti** — D&D 5e encounter calculator (chi + templ + HTMX, in-memory store)

Both share the same tech stack (Go, templ, HTMX), target audience, and deployment model (single binary, no external DB). Keeping them as separate apps adds deployment and maintenance overhead.

## Decision

Merge combattimenti into online as a single Go binary with chi as the HTTP router.

### URL Structure

| Path | Description |
|------|-------------|
| `/` | Landing page with links to tools |
| `/srd` | SRD home (collection overview) |
| `/srd/:collection` | Collection list with filters/pagination |
| `/srd/:collection/rows` | HTMX partial for table rows |
| `/srd/:collection/:slug` | Item detail page |
| `/srd/search` | Full search page |
| `/srd/search/dropdown` | HTMX search dropdown partial |
| `/combattimenti` | Encounter calculator |
| `/combattimenti/calculate` | POST — calculate encounter XP |
| `/combattimenti/party-input` | GET — party input options |
| `/combattimenti/api/difficulties` | GET — difficulties for ruleset |
| `/combattimenti/api/monsters` | GET — monster search (HTMX) |
| `/health` | Health check |
| `/robots.txt` | SEO |
| `/sitemap.xml` | SEO |

### Router: chi over gin

- chi handlers use stdlib `(http.ResponseWriter, *http.Request)` — idiomatic Go
- `r.Mount("/srd", srdRouter)` and `r.Mount("/combattimenti", combattimentiRouter)` compose cleanly
- combattimenti already uses chi — zero adaptation needed
- gin's custom context makes sub-router mounting awkward

### Architecture After Merge

```
cmd/app/                        Single entry point
internal/
  srd/                          Current online domain + adapters (ported from gin to chi)
    domain/
    application/
    adapters/
      repositories/inmemory/
      web/                      HTTP handlers (chi)
    infrastructure/
      datastore/
  combattimenti/                Encounter calculator (copied from due-draghi)
    domain/
      encounter/
      monster/
    application/
      encounter/
      monster/
    infrastructure/
      persistence/memory/
      web/
        handlers/
        templates/
  shared/                       Shared middleware, config, landing page
    middleware/
    config/
web/
  static/                       Merged static assets (CSS, JS, fonts)
  templates/                    Shared templ templates (base layout, landing page)
data/                           Embedded JSON data (SRD + monsters)
```

### Single Binary, No External DB

Both apps use in-memory stores with embedded data. The merged binary:
- Embeds SRD JSON files and monster JSON at compile time
- Loads all data into memory on startup
- Serves both feature sets from a single process

## Inputs

- User navigates to `/`, `/srd/*`, or `/combattimenti/*`
- SRD: collection browsing, filtering, search, item detail
- Combattimenti: party config form, XP calculation, monster search/selection

## Outputs

- HTML pages rendered via templ
- HTMX partials for dynamic updates
- JSON endpoints for combattimenti API (`/combattimenti/api/*`)
- Health check JSON

## Edge Cases

- **Static asset conflicts**: Both apps have CSS design tokens and HTMX. Unify into a single `web/static/` with shared design system. Both already use `due-draghi-design-system` tokens.
- **Base template divergence**: Both have `base.templ` with similar structure (header, nav, footer, dark mode). Create a shared base layout; each feature provides its own content.
- **HTMX version mismatch**: Online bundles htmx.min.js locally; combattimenti loads from CDN. Standardize on local bundle.
- **Middleware overlap**: Both have logging, recovery, CORS, security headers. Unify at the root router level; feature-specific middleware (e.g., collection validation) stays on sub-routers.
- **SEO redirects**: If the SRD was previously served at `/`, existing links to `/:collection/:slug` need 301 redirects to `/srd/:collection/:slug`.

## Error Conditions

- Invalid collection name → 404 (SRD validation middleware)
- Invalid encounter form data → 400 with error message (combattimenti handler validation)
- Monster search with no results → empty table partial
- Panic in any handler → recovery middleware returns 500

## Implementation Steps

1. Switch router from gin to chi (port online handlers)
2. Restructure internal/ to accommodate both feature sets
3. Copy combattimenti domain/application/infrastructure into internal/combattimenti/
4. Merge static assets and base templates
5. Create landing page at `/`
6. Mount SRD routes under `/srd`, combattimenti under `/combattimenti`
7. Unify middleware stack at root level
8. Add SEO redirects for old `/` → `/srd` paths
9. Update Makefile, Dockerfile, CI
10. Verify all tests pass

## Consequences

- Single deployment unit — simpler ops
- Shared design system — consistent UX
- gin dependency removed — fewer dependencies, stdlib-compatible handlers
- Old URLs need redirects during transition
- Larger binary (monsters.json adds ~few MB embedded data)
