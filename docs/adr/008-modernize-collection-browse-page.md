# ADR-008: Modernize Collection Browse Page

## Status

Accepted

## Context

The collection browse page (`/srd/:collection`) is the most-used page type — users spend time here filtering and scanning lists. Current issues:

- **Item list is flat text with dividers**: No visual differentiation between items. Scanning a list of 658 spells is tedious.
- **Filters are powerful but dense**: Multi-select dropdowns work for power users but feel overwhelming for casual browsing.
- **No quick-filter shortcuts**: To filter by spell level, you must open a dropdown, scroll, and check a box. For the most common filters, this is too many clicks.
- **No visual item type indicators**: All items look the same regardless of type/category.
- **Pagination is basic**: "1 di 33" with prev/next. No sense of progress or jump-to-page.

## Decision

### Quick Filter Chips for Top Filters

Add a row of tappable chips above the results for the most common filter dimension per collection:

**Incantesimi** → Spell level chips:
```
[Trucchetto] [1°] [2°] [3°] [4°] [5°] [6°] [7°] [8°] [9°]
```

**Mostri** → CR range chips:
```
[0-¼] [½-1] [2-4] [5-8] [9-12] [13-16] [17-20] [21+]
```

**Equipaggiamenti / Armi / Armature** → Type chips based on subcategory.

Implementation:
- Each collection defines a `quick_filter` in its configuration: which field, which values, display labels.
- Chips are rendered as toggle buttons that set the corresponding filter param.
- Chips sync with the dropdown filter (selecting a chip checks the corresponding dropdown option and vice versa).
- Chips use HTMX to trigger a filtered request, same as the existing filter system.
- Horizontally scrollable on mobile with scroll-snap.

### Enhanced Item Rows

Add subtle visual elements to item rows:

```
┌─────────────────────────────────────────┐
│  Palla di Fuoco                    5.5e │
│  Evocazione · 3° livello · 1 azione    │
└─────────────────────────────────────────┘
```

Changes:
- **Two-line layout**: Title on first line, metadata badges on second line (already partially done via `DisplayElements` — make it consistent).
- **Edition badge**: Small colored pill showing source edition, right-aligned.
- **Subtle left border**: Thin colored accent on hover (`border-left: 3px solid var(--primary)`).
- **Hover background**: `var(--color-hover)` on hover for clear affordance.

### Improved Pagination

Replace "1 di 33" prev/next with a more informative pagination bar:

```
Mostrando 1-20 di 658    [← Precedente]  1 2 3 ... 33  [Successiva →]
```

- Show item range ("1-20 di 658") for context.
- Page numbers with ellipsis for large page counts (show first, last, and 2 around current).
- Current page highlighted.
- Still uses HTMX for partial updates (no full page reload).
- On mobile: simplified to range + prev/next buttons (no page numbers — not enough space).

### Sticky Filter Bar Polish

The filter bar is already sticky — refine it:

- **Shadow on scroll**: Add subtle bottom shadow when the filter bar is stuck (already partially implemented in JS — ensure consistency).
- **Active filter count badge**: Show a small count badge on the filter toggle button when filters are active (already exists — ensure it's prominent).
- **Collapse/expand animation**: Smooth height transition when the filter panel opens/closes.

### Scroll-to-Top on Filter Change

When filters or pagination change and results update via HTMX:

- Scroll the results container to the top smoothly.
- Preserve scroll position only when using the browser back button (HTMX history restore).

## Inputs

- Collection configuration with `quick_filter` definitions.
- Existing filter parameters and pagination params.
- `DisplayElements` from the document mapper.

## Outputs

- Quick filter chip bar per collection.
- Enhanced two-line item rows with edition badges.
- Numbered pagination with item range.
- Polished sticky filter bar.
- Scroll-to-top on updates.

## Edge Cases

- **Collection with no quick filter defined**: Skip the chip bar, show only the dropdown filters.
- **Quick filter with many values** (>10): Horizontally scrollable with fade indicators on edges.
- **Page count is 1**: Hide pagination entirely.
- **Filters produce 0 results**: Show the existing empty state with a "Cancella filtri" button.
- **URL state**: Quick filter selections must be reflected in URL params so links are shareable and back button works.

## Error Conditions

- Quick filter field doesn't exist in collection data → skip chip bar, log warning.
- HTMX partial update fails → show error toast, keep current results visible.

## Consequences

- Most common filtering actions go from 3 clicks (open dropdown → scroll → check) to 1 tap.
- Item rows are easier to scan with two-line layout.
- Pagination gives better orientation in large collections.
- Quick filter config is a new per-collection data structure — small maintenance overhead.
- Chip bar adds a row of UI above results — acceptable tradeoff for the usability gain.
