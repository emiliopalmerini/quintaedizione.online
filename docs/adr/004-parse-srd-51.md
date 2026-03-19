# ADR-004: Parse SRD 5.1 Italian PDF

## Status

Proposed

## Context

We have the SRD CC v5.1 Italian PDF (`SRD_CC_v5.1_IT.pdf`, 453 pages) and need to parse it into JSON to populate the `srd-5e` source defined in ADR-003. The existing parser (`scripts/parse_srd/`) targets the SRD v5.2.1 (2024) PDF, which uses a completely different font family and color scheme.

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

The 5.1 SRD differs from 5.2.1 in content organization:

- **Races** (not "Specie") — includes subraces (Elfo Alto, Piedelesto, etc.)
- **No Glossary section** — rules are inline
- Additional sections: Trappole, Malattie, Follia, Veleni, Piani
- Stat blocks use `Calibri-Bold`/`Calibri` at 10pt body color — no separate Optima/stat-red color scheme. Stat blocks are distinguished only by bold labels, not by font family or color.

### Key Insight: Stat Blocks Are Structurally Simpler

In SRD 5.2.1, stat blocks use distinct fonts (Optima) and colors (`#540000`) making them easy to identify. In SRD 5.1, stat blocks use the **same fonts and colors as body text** (`Calibri` at `#000000`). Stat blocks must be identified by **position in the heading tree** (children of monster/creature H3 headings) rather than by font classification alone.

## Decision

### Make the Classifier Profile-Based

Extract font/color constants into a **profile** dataclass. The classification logic (`classify_span`) stays the same — it maps fonts to semantic roles (H1, H2, BODY, STAT_LABEL, etc.) — but the font-matching rules are parameterized.

```python
@dataclass(frozen=True)
class FontProfile:
    """Font/color calibration for a specific PDF."""
    name: str

    # Heading font and color
    heading_font: str          # "GillSans" or "Calibri"
    heading_color: int         # 0x8c2220 or 0x943634

    # Body color
    body_color: int            # 0x231f20 or 0x000000

    # Stat block — distinct font/color (5.2.1) or None (5.1)
    stat_font: str | None      # "Optima" or None
    stat_color: int | None     # 0x540000 or None

    # Link color
    link_color: int            # 0x1e5e9e or 0x0000ff

    # Footer
    footer_font: str           # "GillSans" or "Calibri"
    footer_color: int | None   # 0x808285 or None (5.1 uses body color)

PROFILE_521 = FontProfile(
    name="SRD 5.2.1",
    heading_font="GillSans", heading_color=0x8c2220,
    body_color=0x231f20,
    stat_font="Optima", stat_color=0x540000,
    link_color=0x1e5e9e,
    footer_font="GillSans", footer_color=0x808285,
)

PROFILE_51 = FontProfile(
    name="SRD 5.1",
    heading_font="Calibri", heading_color=0x943634,
    body_color=0x000000,
    stat_font=None, stat_color=None,
    link_color=0x0000ff,
    footer_font="Calibri", footer_color=None,
)
```

### Separate Section Definitions

New `SECTIONS_51` page ranges. These map to the same `parser_name` values where possible:

```python
SECTIONS_51: list[SectionDef] = [
    SectionDef("Razze",           (2, 7),     "races.json",          "races"),
    SectionDef("Barbaro",         (8, 10),    "classes.json",        "classes"),
    SectionDef("Bardo",           (11, 14),   "classes.json",        "classes"),
    SectionDef("Chierico",        (15, 18),   "classes.json",        "classes"),
    SectionDef("Druido",          (19, 23),   "classes.json",        "classes"),
    SectionDef("Guerriero",       (24, 26),   "classes.json",        "classes"),
    SectionDef("Ladro",           (27, 29),   "classes.json",        "classes"),
    SectionDef("Mago",            (30, 33),   "classes.json",        "classes"),
    SectionDef("Monaco",          (34, 37),   "classes.json",        "classes"),
    SectionDef("Paladino",        (38, 42),   "classes.json",        "classes"),
    SectionDef("Ranger",          (43, 47),   "classes.json",        "classes"),
    SectionDef("Stregone",        (48, 52),   "classes.json",        "classes"),
    SectionDef("Warlock",         (53, 59),   "classes.json",        "classes"),
    SectionDef("Multiclasse",     (60, 62),   "rules_multiclass.json", "rules"),
    SectionDef("Backgrounds",     (65, 67),   "backgrounds.json",    "backgrounds"),
    SectionDef("Equipaggiamento", (68, 84),   "equipment.json",      "equipment"),
    SectionDef("Regole",          (85, 113),  "rules_gameplay.json", "rules"),
    SectionDef("Incantesimi",     (114, 222), "spells.json",         "spells"),
    SectionDef("Trappole",        (223, 227), "rules_traps.json",    "rules"),
    SectionDef("Malattie",        (228, 229), "rules_diseases.json", "rules"),
    SectionDef("Follia",          (230, 231), "rules_madness.json",  "rules"),
    SectionDef("Oggetti Magici",  (232, 297), "magic_items.json",    "magic_items"),
    SectionDef("Mostri",          (298, 410), "monsters.json",       "monsters"),
]
```

Note: page ranges above are approximate and need calibration during implementation.

### Reuse Content Parsers

The content parsers (spells, monsters, equipment, etc.) operate on **semantic roles** (H1, H2, BODY, STAT_LABEL, etc.), not raw fonts. As long as classification maps correctly to the same roles, the existing parsers should work with minimal changes.

The main exception is the **monster parser**: in 5.1, stat block fields (CA, PF, Velocità, ability scores) are `Calibri-Bold` at `#000000` — the same as table headers. The classifier must rely on the heading tree context (under a monster heading) to assign STAT_LABEL/STAT_VALUE roles, or the monster parser must handle `BODY_BOLD`/`BODY` pairs as stat fields when inside a monster context.

**Recommended**: let the 5.1 classifier map `Calibri-Bold 10pt #000000` under monster headings to BODY_BOLD (not STAT_LABEL), and adapt the monster parser to accept BODY_BOLD as stat labels when no STAT_LABEL spans exist. This avoids needing page-context in the classifier.

### New Parser: Races

SRD 5.1 has "Razze" (races with subraces) instead of "Specie". This needs a new `races.py` parser. The output format should match `species.json` as closely as possible, with an additional `subraces` field.

### CLI Changes

The `__main__.py` entry point gains a `--profile` flag:

```
uv run scripts/parse_srd <pdf> --profile 5.1 --output-dir ./output
uv run scripts/parse_srd <pdf> --profile 5.2.1 --output-dir ./output  # default
```

The profile selects both the font profile and the section definitions.

## Inputs

- `SRD_CC_v5.1_IT.pdf` — source PDF (453 pages)
- Font profile constants (calibrated by `--debug-page` inspection)
- Section page ranges (from PDF table of contents)

## Outputs

- `data/ita/json/srd-5e/source.json` — source manifest
- `data/ita/json/srd-5e/classes.json`
- `data/ita/json/srd-5e/races.json` (maps to "specie" collection or new "razze" collection)
- `data/ita/json/srd-5e/backgrounds.json`
- `data/ita/json/srd-5e/equipment.json`
- `data/ita/json/srd-5e/spells.json`
- `data/ita/json/srd-5e/magic_items.json`
- `data/ita/json/srd-5e/monsters.json`
- `data/ita/json/srd-5e/rules_*.json`

## Edge Cases

- **Stat blocks without distinct styling**: monster parser must identify stat fields by label text pattern (`Classe Armatura`, `Punti Ferita`, `Velocità`, `FOR`, `DES`, etc.) rather than font role, since 5.1 uses body fonts for stat blocks
- **Footer detection**: 5.1 footers use `Calibri-Italic 8pt #000000` — same color as body. Must distinguish by size (8pt vs 10pt body) and italic + page position (bottom of page)
- **Classes split across separate TOC entries**: each class is its own L1 heading (unlike 5.2.1 which has a single "Classi" section). Section definitions list each class individually to capture correct page ranges
- **Subraces**: races contain nested subraces that need to be associated with their parent race
- **"Azioni" headers at 11pt**: action section headers in monster stat blocks are `Calibri-Bold 11pt #000000` — slightly larger than body, must be classified as H6 or equivalent

## Error Conditions

- Font profile mismatch (using 5.2.1 profile on 5.1 PDF) → most spans classified as UNKNOWN → quality check catches this via high UNKNOWN ratio
- Page range drift → validate section start by checking first heading matches expected section name

## Implementation Order

1. Add `FontProfile` dataclass and extract 5.2.1 constants into `PROFILE_521`
2. Parameterize `classify_span` to accept a profile
3. Calibrate `PROFILE_51` by running `--debug-page` on key pages (2, 8, 129, 300)
4. Add `SECTIONS_51` page ranges
5. Add `--profile` CLI flag
6. Run existing parsers on 5.1 sections, fix breakages iteratively (spells and equipment likely work first)
7. Adapt monster parser for body-font stat blocks
8. Write `races.py` parser
9. Generate all JSON, run quality checks
10. Place output in `data/ita/json/srd-5e/` with `source.json`
