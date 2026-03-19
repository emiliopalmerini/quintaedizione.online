# ADR-004: Parse SRD 5.1 Italian PDF

## Status

Accepted

## Context

We have the SRD CC v5.1 Italian PDF (`SRD_CC_v5.1_IT.pdf`, 453 pages) and need to parse it into JSON to populate the `srd-5e` source defined in ADR-003. The existing parser (`scripts/parse_srd/`) targets the SRD v5.2.1 (2024) PDF, which uses a completely different font family, color scheme, and document structure.

### Font/Color Differences

| Role | SRD 5.2.1 | SRD 5.1 |
|---|---|---|
| Headings | `GillSans-SemiBold`, `#8c2220` | `Calibri`, `#943634` |
| Body | `Cambria`, `#231f20` | `Cambria`, `#000000` |
| Stat block | `Optima` family, `#540000` | `Calibri`/`Calibri-Bold`, `#000000` |
| Sidebar/table | `GillSans` family | `Calibri` family |
| Links | `Cambria`, `#1e5e9e` | `Calibri`, `#0000ff` |
| Footer | `GillSans`, `#808285` | `Calibri-Italic`, `#000000` |
| Action headers | `GillSans` 12pt, `#8c2220` (H6) | `Calibri-Bold` 11pt, `#000000` |

### Structure Differences

The two PDFs differ fundamentally in how content is organized — not just at the font level but at the heading hierarchy and content layout level:

| Aspect | SRD 5.2.1 | SRD 5.1 |
|---|---|---|
| Species/Races | H5 under "Specie dei personaggi" | H2 under H1("Razze") with subraces |
| Classes | H2(ClassName) > H4(features) | H1(ClassName) > H2("Privilegi") > H3(features) |
| Subclasses | "Sottoclasse del X:" pattern | Known names list or separate H2 sibling |
| Backgrounds | H5 under "Descrizioni dei background" | H3 under H2("Background") |
| Spells subtitle | `{Scuola} di {N}º livello ({classi})` | `{Scuola} di {N}° livello` (no classes) |
| Spell metadata | TABLE_HEADER_SMALL paragraphs | Merged into BODY_ITALIC paragraph |
| Equipment tables | Markdown pipe tables from `tables.py` | Sequential TABLE_HEADER_SMALL/TABLE_BODY paragraphs |
| Monster stat blocks | STAT_LABEL/STAT_VALUE (Optima, #540000) | BODY_BOLD/SIDEBAR (Calibri, #000000) |
| Monster names | H3 | H5 |
| Glossary | Dedicated section (pp. 202-219) | Does not exist |
| Feats | H5 with category subtitles | H3, single feat, no categories |
| Extra sections | — | Trappole, Malattie, Follia, Multiclasse, Personalità |

## Decision

### Separate Parser Packages

The structural differences between editions are too deep for shared parsers with branching logic. Instead, each edition gets its own parser package as a first-class citizen:

```
scripts/parse_srd/
  _cli.py              # Profile resolution → (FontProfile, sections, parser_package)
  profiles.py          # FontProfile dataclass + PROFILE_521, PROFILE_51
  classify.py          # Profile-parameterized classifier (shared)
  extract.py           # Raw span extraction (shared)
  merge.py             # Paragraph grouping (shared, profile-parameterized)
  heading_tree.py      # Heading tree builder (shared)
  markdown_gen.py      # Markdown generation (shared)
  tables.py            # Table detection (shared, profile-parameterized)
  schemas.py           # Output TypedDicts (shared)
  slugify.py           # ID generation (shared)
  quality.py           # Validation (shared)
  section_split.py     # SECTIONS + SECTIONS_51 (shared)
  __main__.py          # CLI entry point (delegates to parser package)
  parsers/             # 5.2.1 parsers
    __init__.py         # Registry for 5.2.1
    spells.py
    monsters.py
    classes.py
    backgrounds.py
    equipment.py
    magic_items.py
    feats.py
    species.py
    rules.py
  parsers_51/           # 5.1 parsers
    __init__.py          # Registry for 5.1
    spells.py
    monsters.py
    classes.py
    backgrounds.py
    equipment.py
    magic_items.py
    feats.py
    races.py             # → outputs species.json
    rules.py
```

Both packages use the same `register` decorator pattern. `__main__.py` imports from the correct package based on `--profile`:

```python
profile, sections = resolve_profile(args.profile)

if profile is PROFILE_51:
    from .parsers_51 import get_parser
else:
    from .parsers import get_parser
```

### Shared Infrastructure

The pipeline stages that operate on semantic roles (not edition-specific structure) remain shared:

- **`profiles.py`** — `FontProfile` dataclass with `PROFILE_521` and `PROFILE_51`
- **`classify.py`** — `classify_span(span, profile)` with `_classify_521` / `_classify_51` branches (font→role mapping is inherently profile-specific, but the SpanRole enum is shared)
- **`extract.py`** — Raw pymupdf span extraction (PDF-agnostic)
- **`merge.py`** — `blocks_to_paragraphs(blocks, profile)` paragraph grouping
- **`tables.py`** — `process_tables(blocks, profile)` table detection
- **`heading_tree.py`** — Stack-based heading tree builder
- **`markdown_gen.py`** — Role→markdown rendering
- **`schemas.py`** — Output TypedDicts (same JSON schema for both editions)
- **`quality.py`** — Validation rules
- **`section_split.py`** — `SECTIONS` + `SECTIONS_51` page ranges

### Section Definitions

```python
SECTIONS_51: list[SectionDef] = [
    SectionDef("Razze",           (2, 7),     "species.json",          "races"),
    SectionDef("Classi",          (8, 59),    "classes.json",          "classes"),
    SectionDef("Multiclasse",     (60, 62),   "rules_multiclass.json", "rules"),
    SectionDef("Personalità",     (63, 64),   "rules_personality.json","rules"),
    SectionDef("Backgrounds",     (65, 67),   "backgrounds.json",      "backgrounds"),
    SectionDef("Equipaggiamento", (68, 83),   "equipment.json",        "equipment"),
    SectionDef("Talenti",         (84, 84),   "feats.json",            "feats"),
    SectionDef("Regole",          (85, 113),  "rules_gameplay.json",   "rules"),
    SectionDef("Incantesimi",     (114, 222), "spells.json",           "spells"),
    SectionDef("Trappole",        (223, 227), "rules_traps.json",      "rules"),
    SectionDef("Malattie",        (228, 229), "rules_diseases.json",   "rules"),
    SectionDef("Follia",          (230, 231), "rules_madness.json",    "rules"),
    SectionDef("Oggetti Magici",  (232, 297), "magic_items.json",      "magic_items"),
    SectionDef("Mostri",          (298, 410), "monsters.json",         "monsters"),
]
```

### CLI

```
uv run scripts/parse_srd <pdf> --profile 5.1 --output-dir ./data/ita/json/srd-5e
uv run scripts/parse_srd <pdf> --profile 5.2.1 --output-dir ./data/ita/json/srd-5.5e  # default
```

### Go Loader Changes

The Go loader (`internal/infrastructure/datastore/loader.go`) needs two changes:

1. **Skip missing collections**: not all sources have all JSON files (e.g., 5.1 has no `glossary.json`). `loadSourceData` skips loaders that fail with `fs.ErrNotExist`.
2. **Discover rules files dynamically**: `loadRules` scans for `rules_*.json` instead of hardcoding three filenames.

## Inputs

- `SRD_CC_v5.1_IT.pdf` — source PDF (453 pages)
- Font profile constants (calibrated by `--debug-page` inspection)
- Section page ranges (from PDF table of contents)

## Outputs

- `data/ita/json/srd-5e/source.json` — source manifest
- `data/ita/json/srd-5e/species.json` (9 races → Species schema)
- `data/ita/json/srd-5e/classes.json` (12 classes)
- `data/ita/json/srd-5e/backgrounds.json` (1 — Accolito)
- `data/ita/json/srd-5e/equipment.json` (52 — weapons + armor)
- `data/ita/json/srd-5e/feats.json` (1 — Lottatore)
- `data/ita/json/srd-5e/spells.json` (319 spells)
- `data/ita/json/srd-5e/magic_items.json` (241 items)
- `data/ita/json/srd-5e/monsters.json` (200 monsters)
- `data/ita/json/srd-5e/rules_*.json` (gameplay, multiclass, personality, traps, diseases, madness)

## Edge Cases

- **Spell subtitle format**: 5.1 uses `°` (U+00B0) not `º` (U+00BA), no class lists, `(rituale)` suffix instead of class parenthetical
- **Spell metadata merged**: all fields in one BODY_ITALIC paragraph; duration extraction needs sentence-boundary heuristic
- **Column-break artifacts**: "Dominare persone" school merged with "Eroismo"; "Creazione" duration lost across column — handled via `_SPELL_OVERRIDES`
- **Monster stat blocks in body fonts**: BODY_BOLD labels + SIDEBAR values, ability scores as sequential paragraphs
- **Monster subtitle merged with stat block**: BODY_ITALIC paragraph contains subtitle + all stat fields; must truncate at "Classe Armatura"
- **Equipment as sequential paragraphs**: TABLE_HEADER_SMALL/TABLE_BODY sequences instead of pipe tables
- **Inline subclasses**: 4 classes (Barbaro, Bardo, Chierico, Druido) have subclass as H3 inside "Privilegi di classe" — identified by `_INLINE_SUBCLASS_NAMES` set
- **Variable rarity magic items**: "rarità variabile" pattern for multi-variant items
- **Avatar della morte**: stat block inside Deck of Many Things, overridden as type "Creatura"

## Error Conditions

- Font profile mismatch (using 5.2.1 profile on 5.1 PDF) → most spans classified as UNKNOWN → quality check catches this
- Missing JSON files for a source → Go loader skips with `fs.ErrNotExist`
- Duplicate source IDs → fatal at startup

## Implementation Order

1. ~~Make classifier profile-based~~ ✓
2. ~~Calibrate PROFILE_51~~ ✓
3. ~~Add SECTIONS_51~~ ✓
4. ~~Add --profile CLI flag~~ ✓
5. ~~Write 5.1 parsers (spells, magic_items, classes, backgrounds, equipment, monsters, races, feats)~~ ✓
6. Split `parsers/` into `parsers/` (5.2.1) + `parsers_51/` (5.1) — revert dual-edition branching from `parsers/`
7. ~~Generate JSON, run quality checks~~ ✓
8. ~~Place output in `data/ita/json/srd-5e/`~~ ✓
9. ~~Update Go loader for missing files and dynamic rules discovery~~ ✓
10. ~~Update ADR status~~ ✓
