"""Split PDF into sections by hardcoded page ranges.

Page ranges match the Italian table of contents for each SRD version.
"""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True, slots=True)
class SectionDef:
    name: str
    pages: tuple[int, int]  # 1-indexed, inclusive
    output_file: str
    parser_name: str


# ── SRD v5.2.1 (2024) ──

SECTIONS: list[SectionDef] = [
    SectionDef("Come si gioca", (5, 20), "rules_gameplay.json", "rules"),
    SectionDef("Creazione del personaggio", (21, 31), "rules_creation.json", "rules"),
    SectionDef("Classi", (32, 92), "classes.json", "classes"),
    SectionDef("Backgrounds", (93, 97), "backgrounds.json", "backgrounds"),
    SectionDef("Specie", (93, 97), "species.json", "species"),
    SectionDef("Talenti", (98, 100), "feats.json", "feats"),
    SectionDef("Equipaggiamento", (101, 117), "equipment.json", "equipment"),
    SectionDef("Incantesimi", (118, 201), "spells.json", "spells"),
    SectionDef("Glossario delle regole", (202, 219), "glossary.json", "glossary"),
    SectionDef("Strumenti di gioco", (220, 231), "rules_tools.json", "rules"),
    SectionDef("Oggetti Magici", (232, 288), "magic_items.json", "magic_items"),
    SectionDef("Mostri", (289, 384), "monsters.json", "monsters"),
    SectionDef("Animali", (385, 405), "monsters.json", "monsters"),
]


# ── SRD v5.1 (2014) ──

SECTIONS_51: list[SectionDef] = [
    SectionDef("Razze", (2, 7), "species.json", "races"),
    SectionDef("Classi", (8, 59), "classes.json", "classes"),
    SectionDef("Multiclasse", (60, 62), "rules_multiclass.json", "rules"),
    SectionDef("Personalità", (63, 64), "rules_personality.json", "rules"),
    SectionDef("Backgrounds", (65, 67), "backgrounds.json", "backgrounds"),
    SectionDef("Equipaggiamento", (68, 83), "equipment.json", "equipment"),
    SectionDef("Talenti", (84, 84), "feats.json", "feats"),
    SectionDef("Regole", (85, 113), "rules_gameplay.json", "rules"),
    SectionDef("Incantesimi", (114, 222), "spells.json", "spells"),
    SectionDef("Trappole", (223, 227), "rules_traps.json", "rules"),
    SectionDef("Malattie", (228, 229), "rules_diseases.json", "rules"),
    SectionDef("Follia", (230, 231), "rules_madness.json", "rules"),
    SectionDef("Oggetti Magici", (232, 297), "magic_items.json", "magic_items"),
    SectionDef("Mostri", (298, 410), "monsters.json", "monsters"),
]


def get_sections_for_parser(
    parser_name: str,
    sections: list[SectionDef] | None = None,
) -> list[SectionDef]:
    """Get all section definitions for a given parser.

    Args:
        parser_name: Parser name to filter by.
        sections: Section list to search. Defaults to SECTIONS (5.2.1).
    """
    if sections is None:
        sections = SECTIONS
    return [s for s in sections if s.parser_name == parser_name]
