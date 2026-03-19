# ADR-003: Multi-Edition Support

## Status

Accepted

## Context

The app currently serves only the SRD 5.2.1 (2024, "5.5e"). We need to support:

- SRD 5e (2014)
- SRD 5.5e (2024) — current
- Future 5e-compatible third-party editions

This affects both the SRD viewer (all collections) and the encounter calculator (monsters + XP rules).

## Decision

### Source-Based Architecture

Each edition is a **source** — a self-contained dataset with metadata. Sources are the unit of data, not editions or rulesets.

#### Data Structure

```
data/ita/json/
  srd-5.5e/                ← current data (renamed)
    source.json            ← manifest
    spells.json
    monsters.json
    classes.json
    ...
  srd-5e/                  ← new (to be produced)
    source.json
    spells.json
    monsters.json
    ...
  <future-source>/
    source.json
    ...
```

#### Source Manifest (`source.json`)

```json
{
  "id": "srd-5.5e",
  "name": "SRD 5.2.1 (2024)",
  "short_name": "5.5e",
  "year": 2024,
  "ruleset": "2024",
  "xp_system": "2024",
  "default": true
}
```

Fields:
- `id` — unique identifier, used in URLs and data tagging
- `name` — display name (e.g. "SRD 5.2.1 (2024)")
- `short_name` — badge/filter label (e.g. "5.5e")
- `year` — for sorting (newest first)
- `ruleset` — which D&D ruleset this source follows ("2014" or "2024")
- `xp_system` — which XP calculation algorithm to use ("2014" or "2024")
- `default` — whether this is the default source when no filter is applied

### Document Tagging

At load time, every document gets two injected fields:

- `_source` — source ID (e.g. `"srd-5.5e"`)
- `_source_short` — short name for display (e.g. `"5.5e"`)

Documents from all sources are loaded into the same collections. A spell named "Palla di Fuoco" may exist in both `srd-5e` and `srd-5.5e` — they are separate documents with different IDs.

### Document IDs & Slugs

To avoid collisions, document IDs are prefixed with the source ID when multiple sources define the same item:

- `palla-di-fuoco` → if only one source has it, slug stays clean
- `srd-5e--palla-di-fuoco` / `srd-5.5e--palla-di-fuoco` → if multiple sources have it

The loader detects collisions at startup and prefixes only the conflicting IDs.

**Alternative (simpler, recommended):** Always prefix with source for all documents:

```
/srd/incantesimi/srd-5.5e/palla-di-fuoco
/srd/incantesimi/srd-5e/palla-di-fuoco
```

This uses a 3-segment URL: `/:collection/:source/:slug`. No collision detection needed. Source becomes a path segment, not part of the slug.

### SRD Viewer

#### Collection List (`/srd/:collection`)

- **Edition filter**: new filter in the filter panel, populated from loaded sources
- **Edition badge**: each row shows the source short name (e.g. "5.5e", "5e")
- **Default**: show all editions (no filter pre-applied), newest first
- **Sort**: within same name, newest edition first

#### Item Detail (`/srd/:collection/:source/:slug`)

- Shows the item from the specified source
- **Cross-edition links**: if the same item exists in other sources, show a bar: "Disponibile anche in: 5e | 5.5e" with links to the other versions
- Breadcrumb includes the source: `SRD / incantesimi / 5.5e / Palla di Fuoco`

### Encounter Calculator

#### Ruleset Selection

Current UI:
```
[x] D&D 2024 (One D&D)
[ ] D&D 2014 (5ª Edizione)
```

New UI — driven by loaded sources:
```
Edizione:
[x] SRD 5.2.1 (2024)    ← shows source name
[ ] SRD 5e (2014)        ← shows source name
```

Each option maps to a source. When selected:
- XP calculation uses the source's `xp_system`
- Difficulty options match the source's `ruleset`
- Monster browser filters to monsters from that source

If multiple sources share the same `xp_system`, their monsters are combined.

#### Monster Browser

- Monsters filtered by the selected source's `xp_system`
- Edition badge on each monster row
- "Vedi pagina completa →" links to the correct source-specific SRD page

### Domain Model Changes

```go
// Source represents a loaded edition/dataset
type Source struct {
    ID        string // "srd-5.5e"
    Name      string // "SRD 5.2.1 (2024)"
    ShortName string // "5.5e"
    Year      int
    Ruleset   string // "2014" or "2024"
    XPSystem  string // "2014" or "2024"
    Default   bool
}
```

The datastore loader iterates over subdirectories in `data/ita/json/`, reads each `source.json`, and loads all collection files with source tagging.

### Migration

1. Create `data/ita/json/srd-5.5e/` directory
2. Move all current JSON files into it
3. Add `source.json` manifest
4. Update `embed.go` to embed subdirectories
5. Update loader to iterate sources
6. Add `_source` field injection
7. Add edition filter to SRD collection UI
8. Update collection URL routing for source segment
9. Update encounter calculator to use source-driven rulesets
10. Update monster repository to accept source filtering

## Inputs

- `data/ita/json/<source>/source.json` — source manifest
- `data/ita/json/<source>/*.json` — collection data files

## Outputs

- Documents tagged with `_source` and `_source_short` fields
- Edition filter available on all collection pages
- Source-specific item URLs: `/srd/:collection/:source/:slug`
- Source-driven ruleset selection in encounter calculator

## Edge Cases

- **Source with no monsters**: calculator hides that source from ruleset selection
- **Source with unknown xp_system**: falls back to "2024" system with warning log
- **Empty source directory**: skip with warning
- **Missing source.json**: skip directory with warning
- **Single source loaded**: edition filter hidden (no need), URLs work without source segment for backward compatibility

## Error Conditions

- Invalid `source.json` schema → log error, skip source
- Duplicate source IDs → fatal at startup
- No default source → use the one with highest year

## Consequences

- Adding a new edition = adding a directory with `source.json` + JSON files. No code changes.
- URL structure changes from `/srd/:collection/:slug` to `/srd/:collection/:source/:slug`
- Old URLs need redirects (single-source backward compatibility)
- Larger binary (more embedded JSON)
- Filter UI always shows edition filter when multiple sources loaded
