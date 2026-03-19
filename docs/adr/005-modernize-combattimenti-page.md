# ADR-005: Modernize Combattimenti Page

## Status

Proposed

## Context

The `/combattimenti/` encounter calculator is functional but feels like a plain settings form rather than an interactive DM tool. Issues:

- **Linear form layout**: Top-to-bottom sections with a submit button — no interactivity until you click "Calcola". Feels static.
- **Plain radio buttons/selects**: Edition and party mode toggles are unstyled native radios. No visual weight.
- **Inline styles in template**: `style="display: grid; ..."` scattered through `home.templ` instead of classes.
- **Dead results area**: A dashed placeholder box saying "I risultati appariranno qui" — uninviting, wastes vertical space.
- **No visual identity**: The page has no D&D atmosphere. It's a generic calculator.
- **Monster browser not discoverable**: The landing page promises "browser dei mostri integrato" but the calculator page doesn't surface it.

## Decision

### Visual Identity — Dyson Logos Maps as Decoration

Use [Dyson Logos](https://dysonlogos.blog/) free-license maps as atmospheric background/decoration elements:

- **Hero background**: A faded, low-opacity dungeon map behind the page header. CSS `background-image` with `opacity: 0.06` (light) / `0.03` (dark), blended with the page background.
- **Results card**: Subtle map fragment as background texture on the result card when results appear.
- **Attribution**: Footer text "Cartography by Dyson Logos" with link to dysonlogos.blog, per license terms.

Ship 2-3 pre-selected map images as static assets (`.webp`, optimized). No runtime fetching.

### Layout — Side-by-Side on Desktop

Replace the single-column form with a two-panel layout on desktop (≥768px):

```
┌─────────────────────┬─────────────────────┐
│   PARTY & CONFIG    │    RESULTS PANEL    │
│                     │                     │
│  Edition toggle     │  (live-updating)    │
│  Party composition  │  XP Budget gauge    │
│  Difficulty         │  Difficulty tiers   │
│  Monster count      │  Monster browser    │
│                     │                     │
└─────────────────────┴─────────────────────┘
```

On mobile, stack vertically: config first, results below.

### Live-Updating Results (No Submit Button)

Remove the "Calcola Budget XP" button. Instead:

- Every input change triggers an HTMX request (`hx-trigger="change"`) to `/combattimenti/calculate`.
- Results panel updates in real-time as the user adjusts inputs.
- Add `hx-indicator` spinner on the results panel during calculation.
- On initial load, compute default results immediately (level 3, 4 players, moderate difficulty).

### Edition Toggle — Segmented Control

Replace plain radio buttons with a styled segmented control:

```html
<div class="segmented-control">
  <button class="segmented-control-option active">5.5e</button>
  <button class="segmented-control-option">5e</button>
</div>
```

CSS: pill-shaped container, active option gets `background: var(--primary)`, smooth sliding indicator. Reusable component in `components.css`.

### Party Composition — Stepper Controls

Replace bare `<input type="number">` with visual stepper components:

```
┌─────────────────────────────┐
│  Livello    [−] 3 [+]      │
│  Personaggi [−] 4 [+]      │
└─────────────────────────────┘
```

- Large tap targets (44px minimum) for +/- buttons.
- Number displayed prominently in the center.
- Constrained to valid ranges (level 1-20, count 1-8).
- For mixed-level mode: card-style character entries with stepper for each.

### Results Panel — Visual Budget Display

Replace the placeholder with an always-visible results panel:

- **XP Budget gauge**: Horizontal bar showing the budget with color-coded difficulty zones (green → yellow → red).
- **Difficulty tier table**: Show all tiers with the selected one highlighted.
- **Summary stats**: Party size, average level, XP per character.
- **Animated transitions**: Results slide/fade when values change.

### Monster Browser Integration

Add a collapsible "Sfoglia Mostri" section below/within the results panel:

- Quick-search input for monsters by name.
- Filter by CR range.
- Each monster row shows: name, CR, XP, and a "+" button to add to the encounter.
- Added monsters appear in an "Encounter Roster" with running XP total.
- Compare running total against the budget gauge.

This is the **most ambitious part** — could be a follow-up ADR if scope is too large.

### Cleanup — Remove Inline Styles

Move all inline styles from `home.templ` to proper CSS classes:
- `.form-grid-2col` for the two-column grid
- `.character-levels-list` for the flex column container
- `.form-actions` for the submit area
- `.result-area` for the results container

## Inputs

- Party level(s), count, edition, difficulty, monster count — same as current.
- Dyson Logos map images (2-3 `.webp` files, ~50-100KB each).

## Outputs

- Live-updating results panel (no page reload, no submit button).
- Side-by-side layout on desktop.
- Segmented control, stepper components.
- Visual XP budget gauge.
- Atmospheric map decoration.
- Attribution footer.

## Edge Cases

- **JS disabled**: Falls back to submit button (progressive enhancement). The `<button type="submit">` remains in HTML, hidden when JS is active.
- **Very slow connection**: Debounce HTMX requests (200ms) to avoid flooding the server with every keystroke on steppers.
- **Dark mode**: Map overlays must work on both light and dark backgrounds — use `mix-blend-mode: multiply` (light) / `screen` (dark).
- **Mobile**: Side-by-side collapses to stacked. Stepper buttons must be ≥44px touch targets.
- **Reduced motion**: Disable gauge animations and result transitions.

## Error Conditions

- HTMX calculate request fails → show error toast, keep last valid result visible.
- Invalid input values (level > 20, count < 1) → clamp at input level, don't send request.
- Map images fail to load → page works fine, just no decoration (CSS background gracefully degrades).

## Consequences

- More interactive, game-like feel — matches the tool's purpose.
- Removes the "click and wait" friction of the current submit-based flow.
- Adds ~150-300KB of static assets (map images).
- Segmented control and stepper are reusable components for other pages.
- Monster browser integration is the biggest scope risk — consider splitting into ADR-005b.
