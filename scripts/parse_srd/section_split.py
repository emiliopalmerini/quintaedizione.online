"""Split PDF into sections by hardcoded page ranges.

Page ranges match the SRD v5.2.1 Italian table of contents.
"""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True, slots=True)
class SectionDef:
    name: str
    pages: tuple[int, int]  # 1-indexed, inclusive
    output_file: str
    parser_name: str


SECTIONS: list[SectionDef] = [
    SectionDef("Come si gioca", (5, 20), "rules_gameplay.json", "rules"),
    SectionDef("Creazione del personaggio", (21, 31), "rules_creation.json", "rules"),
    SectionDef("Classi", (32, 92), "classes.json", "classes"),
    SectionDef("Backgrounds", (93, 97), "backgrounds.json", "backgrounds"),
    SectionDef("Specie", (93, 97), "species.json", "species"),
    SectionDef("Talenti", (98, 100), "feats.json", "feats"),
    SectionDef("Armi", (101, 103), "equipment.json", "equipment"),
    SectionDef("Armature", (104, 105), "equipment.json", "equipment"),
    SectionDef("Strumenti", (106, 107), "equipment.json", "equipment"),
    SectionDef("Equipaggiamenti", (108, 112), "equipment.json", "equipment"),
    SectionDef("Cavalcature e Veicoli", (113, 114), "equipment.json", "equipment"),
    SectionDef("Servizi", (115, 117), "equipment.json", "equipment"),
    SectionDef("Incantesimi", (118, 201), "spells.json", "spells"),
    SectionDef("Glossario delle regole", (202, 219), "glossary.json", "glossary"),
    SectionDef("Strumenti di gioco", (220, 231), "rules_tools.json", "rules"),
    SectionDef("Oggetti Magici", (232, 288), "magic_items.json", "magic_items"),
    SectionDef("Mostri", (289, 384), "monsters.json", "monsters"),
    SectionDef("Animali", (385, 405), "monsters.json", "monsters"),
]


def get_sections_for_parser(parser_name: str) -> list[SectionDef]:
    """Get all section definitions for a given parser."""
    return [s for s in SECTIONS if s.parser_name == parser_name]
