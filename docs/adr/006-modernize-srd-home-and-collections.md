# ADR-006: Modernize SRD Home & Collection Grid

## Status

Accepted

## Context

The `/srd` page serves as the main hub for browsing all SRD content. Current issues:

- **Flat collection grid**: All 11+ collections displayed as equal-weight cards in a uniform grid. No grouping, no hierarchy — "Classi" and "Servizi" have the same visual weight despite vastly different importance.
- **Emoji icons**: Collection icons are emoji characters (⚔️, 📜, ✨). They render inconsistently across platforms, are small, and don't create a strong visual identity.
- **Hero is text-only**: "SRD 5e" heading with a subtitle. No atmosphere, no visual hook.
- **No content stats**: Users don't see the breadth of available content at a glance beyond individual card counts.
- **Search works but feels disconnected**: The dropdown search is functional but the transition to the full search results page is abrupt.

## Decision

### Dyson Logos Map as Hero Background

Add an atmospheric dungeon map to the hero section:

- Full-width background image behind the hero text, heavily faded (`opacity: 0.05`).
- On dark mode: same image, lighter blend (`opacity: 0.03`, `mix-blend-mode: screen`).
- Use the same pre-selected map images from ADR-005 (shared static assets).
- The map adds texture without distracting from the search bar, which remains the primary action.

### Collection Grouping with Section Headers

Group collections into semantic categories with thin section dividers:

```
── Personaggi ──────────────────────
  Classi  ·  Specie  ·  Backgrounds  ·  Talenti

── Magia & Mostri ──────────────────
  Incantesimi  ·  Mostri  ·  Oggetti Magici

── Equipaggiamento ─────────────────
  Equipaggiamenti  ·  Armi  ·  Armature  ·  Strumenti

── Riferimento ─────────────────────
  Regole  ·  Servizi  ·  Glossario
```

Implementation:
- Define groups in Go (ordered slice of `CollectionGroup{Label, Collections}`).
- Template iterates groups, rendering a section header + card grid per group.
- Cards within a group maintain the current card design but are visually contained.
- On mobile: groups stack naturally, cards go 2-column within each group.

### SVG Collection Icons (Replace Emoji)

Replace emoji with custom SVG icons — simple, monochrome line icons that match the design tokens:

- Color: `var(--color-text-secondary)`, transitions to `var(--primary)` on card hover.
- Size: 32px (up from ~16px emoji).
- Style: Thin line-art, consistent stroke width, D&D-themed.
- Source: Hand-drawn or from an open icon set (Lucide, Tabler Icons, or custom).

Icon mapping:
| Collection | Icon concept |
|---|---|
| Classi | Shield with sword |
| Specie | Silhouette figures |
| Backgrounds | Scroll |
| Talenti | Star/diamond |
| Incantesimi | Sparkles/wand |
| Mostri | Dragon head |
| Oggetti Magici | Glowing ring |
| Equipaggiamenti | Backpack |
| Armi | Crossed swords |
| Armature | Chestplate |
| Strumenti | Wrench/tools |
| Regole | Open book |
| Servizi | Coins |
| Glossario | Dictionary/index |

Icons are inline SVG in `icons.templ` (already has the `CollectionIcon` component — replace emoji with SVG).

### Stats Banner

Add a compact stats row between hero and collection grid:

```
📊 2,400+ voci  ·  14 collezioni  ·  2 edizioni
```

- Small, muted text. Not a hero element — just context.
- Numbers are computed at startup from the datastore.
- Updates automatically when new sources are added.

### Collection Card Enhancements

Keep the current card structure but refine:

- **Larger icon area**: Icon sits in a colored circle/pill background that uses the group's accent.
- **Count badge position**: Move from header to a subtle corner badge.
- **Hover effect**: Card lifts slightly (`translateY(-2px)`) with shadow increase — already partially implemented, make it consistent.
- **Active state**: Brief press feedback on mobile (`:active` scale).

## Inputs

- Collection definitions with group assignments (Go struct).
- SVG icon definitions (inline in `icons.templ`).
- Dyson Logos map images (shared with ADR-005).
- Collection counts from datastore.

## Outputs

- Grouped collection grid with section headers.
- SVG icons replacing emoji.
- Stats banner.
- Atmospheric hero background.
- Refined card hover/active states.

## Edge Cases

- **Single collection in a group**: Render the group normally (even with one card).
- **New collection added**: Must be assigned to a group in Go code, otherwise falls into a default "Altro" group.
- **SVG accessibility**: Each SVG gets `aria-hidden="true"` since the card title already provides the label.
- **Dark mode**: SVG icons use `currentColor` so they adapt automatically. Map background blend mode switches.

## Error Conditions

- Missing SVG for a collection → fall back to a generic document icon.
- Map image fails to load → hero renders normally without background texture.

## Consequences

- Stronger visual hierarchy — users find important collections faster.
- More polished, app-like feel vs. the current flat grid.
- SVG icons render consistently across all platforms (unlike emoji).
- Group definitions are a new data structure to maintain — but changes are rare.
- Stats banner adds a sense of scale ("2,400+ items") that builds trust.
