# ADR-007: Modernize Search Experience

## Status

Proposed

## Context

Search is the primary way users find specific content. The current implementation has two modes — dropdown autocomplete from `/srd` and a full results page at `/srd/search`. Issues:

- **Empty state is unhelpful**: "Nessun risultato trovato — Prova con altri termini di ricerca" gives no guidance. No suggestions, no alternatives.
- **No fuzzy matching feedback**: Search uses `FuzzySearchService` internally but doesn't communicate to users why something didn't match or suggest corrections.
- **Back link is wrong**: Search results page links "Torna alla home" to `/` (landing) instead of `/srd` (where the search originated).
- **No context snippets**: Results show title + badges but no preview of the matching content.
- **Dropdown-to-page transition is jarring**: Clicking "Vedi tutti" in the dropdown navigates to a different page layout with no visual continuity.
- **Collection filter in dropdown is mobile-only**: The overlay search has collection filter chips, but desktop dropdown doesn't.

## Decision

### Fix the Back Link

Immediate fix: change `search.templ` back link from `/` to `/srd`.

### Empty State with Suggestions

Replace the minimal empty state with a helpful one:

```
┌─────────────────────────────────────┐
│         🔍 Nessun risultato         │
│                                     │
│  Suggerimenti:                      │
│  · Controlla l'ortografia           │
│  · Prova con meno parole chiave     │
│  · Cerca in italiano (es. "Palla    │
│    di Fuoco" non "Fireball")        │
│                                     │
│  Collezioni popolari:               │
│  [Incantesimi] [Mostri] [Classi]   │
└─────────────────────────────────────┘
```

- Contextual tips (the site is in Italian — remind users to search in Italian).
- Quick links to popular collections as fallback navigation.

### Context Snippets in Results

Add a one-line text snippet below each result title showing where the match occurred:

```
Palla di Fuoco
Evocazione · 3° livello · 5.5e
"...Ogni creatura nell'area deve effettuare un tiro salvezza..."
```

Implementation:
- `FuzzySearchService` already has access to the document body.
- Extract ~100 characters around the first match position.
- Highlight the matched term with `<mark>`.
- Snippet is plain text (strip HTML/markdown tags before extracting).

### Collection Filter on Desktop Dropdown

Add filter chips to the desktop dropdown search, matching the mobile overlay behavior:

```
┌──────────────────────────────────┐
│ 🔍 [search input................]│
│ [Tutte] [Incantesimi] [Mostri]  │ ← filter chips
│──────────────────────────────────│
│ Incantesimi (3)                  │
│   Palla di Fuoco · 3° livello   │
│   ...                            │
└──────────────────────────────────┘
```

- Chips are scrollable horizontally if they overflow.
- Selected chip filters the HTMX dropdown request via `collection` query param (already supported by the handler).
- Active chip gets `var(--primary)` background.

### Smoother Dropdown → Full Page Transition

When the user clicks "Vedi tutti →" or presses Enter with a query:

- Navigate to `/srd/search?q=...` but preserve the search query in the URL.
- On the search results page, auto-focus the search input with the query pre-filled so the user can refine.
- Add a search input at the top of the results page (currently absent — the page only shows the query as text).

### Search Input on Results Page

Add a search bar to the results page header so users can refine without going back:

```
┌──────────────────────────────────┐
│ ← SRD                           │
│                                  │
│ 🔍 [palla di fuoco............] │
│                                  │
│ 12 risultati                     │
│ ─── Incantesimi (3) ──────────  │
│ ...                              │
└──────────────────────────────────┘
```

- Same search input component as the home page.
- Submitting navigates to the same page with updated `q` param.
- No dropdown on the results page — direct submission only.

## Inputs

- Search query string (`q` param).
- Optional collection filter (`collection` param).
- Document body text for snippet extraction.

## Outputs

- Improved empty state with suggestions and collection links.
- Context snippets on search results (both dropdown and full page).
- Collection filter chips on desktop dropdown.
- Search input on the results page.
- Fixed back link.

## Edge Cases

- **Query matches title but not body**: Show title match without body snippet.
- **Very long matched text**: Truncate snippet to 120 characters with ellipsis.
- **HTML in document body**: Strip all tags before snippet extraction to avoid broken markup.
- **Special characters in query**: Escape for regex highlighting, pass through for display.
- **No collections loaded**: Filter chips section hidden.

## Error Conditions

- Snippet extraction on malformed HTML → fall back to no snippet (title-only display).
- Search service timeout → show error message with retry option.

## Consequences

- Users get meaningful feedback when search fails instead of a dead end.
- Context snippets help users evaluate results without clicking through.
- Desktop search gets the same filtering power as mobile.
- Search feels like a continuous experience rather than two disconnected modes.
- Snippet extraction adds minor processing overhead per search request — acceptable for the result set sizes involved (max ~20 results displayed).
