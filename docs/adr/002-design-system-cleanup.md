# ADR-002: Design System Cleanup — Unified Soft Aesthetic

## Status

Accepted

## Context

After merging combattimenti into online (ADR-001), the UI layer has accumulated significant inconsistencies:

- **Two competing aesthetics**: SRD uses brutalist (border-radius: 0, no shadows), combattimenti uses softer (rounded, shadows, hover transforms)
- **CSS duplication**: `encounters.css` (1549 lines) redefines forms/buttons/cards from `main.css`; `quintaedizione.css` (64KB) overrides everything
- **Rogue color variables**: `--celestial-blue`, `--bittersweet-shimmer` defined outside `tokens.css`
- **JS duplication**: Theme toggle and Patreon banner logic identical in `main.js` and `encounters.js`; two notification systems; three search implementations
- **Template inconsistency**: SRD base uses utility classes, combattimenti base uses inline styles and embeds 100 lines of form JS

## Decision

Unify the entire UI under a **soft, Notion-inspired aesthetic** with a single design system.

### 1. tokens.css — Single Source of Truth

All visual values defined here. Changes:
- Add `--radius-sm: 4px`, `--radius-md: 8px`, `--radius-lg: 12px` (replace brutalist `0` and random `1.5rem`)
- Add `--shadow-card: 0 1px 3px rgba(0,0,0,0.08)`, `--shadow-card-hover: 0 4px 12px rgba(0,0,0,0.1)`
- Add `--transition-base: 0.15s ease` (single transition token)
- Remove `--celestial-blue`, `--bittersweet-shimmer` and all rogue color vars — map to semantic tokens (`--color-info`, `--color-danger`, etc.)
- Consolidate `--notion-*` aliases into semantic names or remove redundancy

### 2. main.css — Component Library

Single definition for each component. All components adopt softer style:
- **Buttons**: `border-radius: var(--radius-md)`, subtle hover shadow
- **Cards**: `border-radius: var(--radius-md)`, `box-shadow: var(--shadow-card)`, hover lift
- **Forms**: `.field` with `border-radius: var(--radius-sm)`, focus ring
- **Tables**: Rounded container, subtle row hover
- Remove all `border-radius: 0` overrides
- Remove all `box-shadow: none` overrides
- Remove all `transform: none` overrides

### 3. Eliminate encounters.css

Move any encounters-specific styles (monster browser, encounter result, XP tracker) into main.css as components. Delete encounters.css entirely — no feature-specific CSS files.

### 4. Eliminate quintaedizione.css

Audit all 64KB. Anything that's a component → main.css. Anything that's a token → tokens.css. Anything that's a utility → utilities.css. Delete the file.

### 5. utilities.css — Deduplicate

Remove utilities already in main.css. Keep only genuine layout/spacing helpers (`.flex`, `.gap-*`, `.mx-auto`, `.py-*`, etc.) that aren't component-specific.

### 6. Shared JS module

- Extract `initThemeToggle()` and `initPatreonBanner()` into `main.js` (single source)
- Remove duplicates from `encounters.js`
- Consolidate to one notification system (CSS-based `.toast`)
- Move encounters form logic out of `base.templ` into `encounters.js`
- Deduplicate search logic in `main.js` (one search handler, parameterized)

### 7. Shared base template

- Replace combattimenti's `base.templ` with the SRD `BaseLayout` (or a new shared one)
- Remove all inline styles from templates — use utility classes
- Both features use the same header, footer, theme toggle, patreon banner

## Inputs

- `tokens.css` — design token definitions
- `main.css` — component styles
- `utilities.css` — layout helpers
- `main.js` — shared JS
- `encounters.js` — encounters-specific JS (no duplicates)
- `base.templ` — one shared layout

## Outputs

After cleanup:
```
web/static/
  css/
    tokens.css        Design tokens (colors, spacing, radius, shadows, transitions)
    main.css          All component styles (buttons, cards, forms, tables, nav, etc.)
    utilities.css     Layout utilities only
  main.js             Shared JS (theme, banner, notifications, search)
  js/
    encounters.js     Encounters-only JS (form logic, monster selection)
  htmx.min.js
  favicon.svg
```

## Edge Cases

- **Dark mode**: All new shadow/radius tokens must have dark mode equivalents in tokens.css
- **Reduced motion**: Hover transforms and transitions must respect `prefers-reduced-motion`
- **Mobile**: Touch targets must remain ≥44px; hover effects should degrade gracefully on touch
- **Print**: Shadows and background colors should be hidden in print media

## Error Conditions

- Missing CSS variable fallbacks → components render without style
- JS deduplication breaks event binding → test theme toggle, search, filters on both /srd and /combattimenti

## Implementation Steps

1. Refactor `tokens.css` — consolidate all variables, add radius/shadow/transition tokens
2. Rewrite `main.css` — soft aesthetic for all components, remove brutalist overrides
3. Absorb `encounters.css` into `main.css` — encounters-specific components as new sections
4. Absorb `quintaedizione.css` into `main.css` / `tokens.css` — delete the file
5. Deduplicate `utilities.css`
6. Refactor `main.js` — deduplicate search, extract shared functions
7. Refactor `encounters.js` — remove duplicates, move form logic from template
8. Unify `base.templ` — one shared layout, replace inline styles
9. Verify all pages render correctly (landing, SRD home, collection, item, search, combattimenti, calculate result)
10. Verify dark mode, mobile, accessibility

## Consequences

- Consistent visual identity across all pages
- ~60% reduction in CSS size (eliminate 64KB quintaedizione.css + 1549-line encounters.css duplication)
- Single source of truth for design decisions
- Easier to add new features with consistent component library
- Breaking change for any external CSS customization (unlikely given this is a self-contained app)
