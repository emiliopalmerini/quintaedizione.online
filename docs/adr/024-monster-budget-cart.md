# ADR-024: Monster Budget Cart

## Status

Proposed

## Context

`ADR-005` introduced a monster browser embedded inside the encounter result panel on `/combattimenti`. That browser drifted out of sync with the server-computed XP budget (commits `af9da79`, `53ae4ba`, `bacbbbc`, `ebbb26b`) and was ultimately removed in `c30b680`, replaced by a static link to `/srd/mostri`.

The user now wants the monster picker back, but framed differently: the XP budget is a **wallet**; each monster has a **price** (its listed XP); picking monsters spends from the wallet. This makes the feature single-purpose (encounter construction) rather than generic browsing, which is what `/srd/mostri` already covers well.

Key prior decisions relevant here:

- `0a5d1c9` unified monster data with the SRD embed.FS; the combattimenti-owned `monsters.json` was duplicated data.
- `ADR-022` introduced multi-version documents and a default source for deduplication.
- 2014 rules apply an encounter multiplier based on monster count (`EncounterRepository.GetMultiplierFor2014`); 2024 rules do not.

## Decision

Add a server-rendered monster cart under the result panel on `/combattimenti`. The cart re-prices itself on each change by driving a full form recalculation, using the existing HTMX submission pipeline. Cart state is ephemeral (client-side in the DOM between recalculations); it does not survive a party/difficulty recalculation, per the user's choice.

### Architecture

- **Port** in `internal/combattimenti/domain/monster` defining a narrow read model (ID, name, XP, CR, source) and a search signature filtered by edition and max XP.
- **Adapter** in `internal/combattimenti/infrastructure/persistence/srd` that wraps the SRD `DocumentRepository` and parses XP from the `cr` field (format `"10 (PE 5.900; BC +4)"`, already present on every monster document via the SRD loader at `internal/srd/infrastructure/datastore/loader.go:274`).
- No new JSON data, no duplicate entities, no dependency from SRD on combattimenti.

### Inputs

Cart operates as a pair of handlers:

1. `GET /combattimenti/monsters?source={short}&max_xp={int}&q={string}` — returns HTML fragment listing affordable monsters.
2. `POST /combattimenti/calculate` — accepts repeated `monsters[]={id}@{source}` entries alongside party/difficulty form fields, re-renders result panel including the cart subtotal, effective cost after multiplier (2014 only), and remaining budget.

Where:

- `source` is the edition short name (e.g. `5.5e`, `5e`) matching the currently selected ruleset source.
- `max_xp` defaults to the current budget for the selected difficulty; clamped at `[0, 1_000_000]`.
- `q` is a case-insensitive substring match against monster name; empty → no name filter.
- `monsters[]` is the cart; 0..N entries; unknown IDs silently dropped (log at debug).

### Output

1. `GET /combattimenti/monsters` renders a `templ` partial `MonsterPicker` with:
   - Search input (debounced, triggers `GET` with `hx-include`).
   - Monster list rows: name, CR, XP, "Aggiungi" button that posts the monster into the cart via HTMX.
   - "Affordable only" toggle that flips `max_xp` between budget and `1_000_000`.
2. `POST /combattimenti/calculate` returns the existing `Result` template, extended with:
   - `CartEntries []CartEntry{ID, Name, Source, CR, UnitXP}` rendered as chips with remove buttons.
   - `CartSubtotal int` — sum of raw unit XP.
   - `CartEffectiveCost int` — `CartSubtotal * multiplier(len(cart))` for 2014, equal to `CartSubtotal` for 2024.
   - `CartRemaining int` — `TotalXP - CartEffectiveCost`, signed (can go negative).

### Behavior

1. On page load, the result panel renders with an empty cart and the "Aggiungi mostri" picker closed.
2. Each form change fires `calculate` (existing `encounters.js` debounce). The handler re-renders cart rows from the posted `monsters[]`, re-prices the cart against the freshly-computed budget.
3. Clicking "Aggiungi" on a monster row issues a hidden form submit that appends a `monsters[]` input, then fires `calculate`. The picker stays open.
4. Clicking a cart chip's remove button removes the hidden `monsters[]` input and fires `calculate`.
5. Switching edition (`ruleset`) clears the cart (monsters from one edition should not cost against another edition's budget).
6. Switching ruleset 2014↔2024 rebuilds the effective-cost row (shown only for 2014).

### 2014 Effective Price

For 2014 rules the effective cost is:

```
effective = raw_sum * multiplier(cart_size)
```

where `multiplier` is the existing `EncounterRepository.GetMultiplierFor2014` table. The `num_monsters_2014` form field becomes derived from cart size (the stepper is hidden when the cart has ≥1 monster; shown as a manual override when the cart is empty, preserving the current behavior).

The budget itself stays computed from `num_monsters_2014` stepper OR cart size, whichever is non-zero (cart wins).

### Edge Cases

- Empty cart: subtotal = 0, effective = 0, remaining = budget; no warning.
- Over budget: remaining is negative; UI shows it in red, no hard error.
- Monster missing XP (parse fails): repository returns XP = 0; monster is still listed but with "XP: —"; adding it costs 0.
- Unknown source: returns empty list, no error.
- `q` yielding zero results: empty list with a "nessun mostro" hint.
- Duplicate adds of the same monster: each instance is a separate cart entry (parties can fight 3 goblins).
- Ruleset change mid-cart: cart is cleared server-side on the next calculate when the submitted `source` diverges from every cart entry's source.

### Error Conditions

- Malformed `monsters[]` entry (missing `@source` separator): entry silently dropped; warn log.
- SRD repository read error: cart re-renders without that monster; warn log with request ID.

## Consequences

- `/combattimenti` gets back an embedded picker without reintroducing duplicated monster data.
- The combattimenti slice depends on the SRD slice's public repository interface (`srd/domain/repositories.DocumentReader`), which is acceptable per hexagonal architecture as long as the dependency points inward (it does; SRD has no reverse dependency).
- The old `internal/combattimenti/domain/monster` package from commit `c30b680^` is NOT restored verbatim; the new port is narrower (no traits/actions/skills).
- `CalculateXPRequest` grows a `MonsterIDs []string` field; the XP calculation itself is unchanged.
- `/srd/mostri` link from `result.templ` is replaced by the inline picker.
