from __future__ import annotations

from dataclasses import dataclass
from typing import Callable, List

from .parsers.classes import parse_classes
from .parsers.backgrounds import parse_backgrounds
from .parsers.documents import parse_document


@dataclass
class WorkItem:
    filename: str
    collection: str
    parser: Callable


DEFAULT_WORK: List[WorkItem] = [
    # Document pages (Italian)
    WorkItem(
        "ita/01_informazioni_legali.md",
        "documenti",
        lambda lines: parse_document(lines, "01_informazioni_legali.md"),
    ),
    WorkItem(
        "ita/02_giocare_il_gioco.md",
        "documenti",
        lambda lines: parse_document(lines, "02_giocare_il_gioco.md"),
    ),
    WorkItem(
        "ita/03_creazione_personaggio.md",
        "documenti",
        lambda lines: parse_document(lines, "03_creazione_personaggio.md"),
    ),
    WorkItem(
        "ita/04_classi.md",
        "documenti",
        lambda lines: parse_document(lines, "04_classi.md"),
    ),
    WorkItem(
        "ita/05_origini_personaggio.md",
        "documenti",
        lambda lines: parse_document(lines, "05_origini_personaggio.md"),
    ),
    WorkItem(
        "ita/06_talenti.md",
        "documenti",
        lambda lines: parse_document(lines, "06_talenti.md"),
    ),
    WorkItem(
        "ita/07_equipaggiamento.md",
        "documenti",
        lambda lines: parse_document(lines, "07_equipaggiamento.md"),
    ),
    WorkItem(
        "ita/08_equipaggiamento_items.md",
        "documenti",
        lambda lines: parse_document(lines, "08_equipaggiamento_items.md"),
    ),
    WorkItem(
        "ita/09_armi_items.md",
        "documenti",
        lambda lines: parse_document(lines, "09_armi_items.md"),
    ),
    WorkItem(
        "ita/10_oggetti_magici_items.md",
        "documenti",
        lambda lines: parse_document(lines, "10_oggetti_magici_items.md"),
    ),
    WorkItem(
        "ita/11_armatura_items.md",
        "documenti",
        lambda lines: parse_document(lines, "11_armatura_items.md"),
    ),
    WorkItem(
        "ita/12_strumenti_items.md",
        "documenti",
        lambda lines: parse_document(lines, "12_strumenti_items.md"),
    ),
    WorkItem(
        "ita/13_servizi_items.md",
        "documenti",
        lambda lines: parse_document(lines, "13_servizi_items.md"),
    ),
    WorkItem(
        "ita/14_cavalcature_veicoli_items.md",
        "documenti",
        lambda lines: parse_document(lines, "14_cavalcature_veicoli_items.md"),
    ),
    WorkItem(
        "ita/15_incantesimi.md",
        "documenti",
        lambda lines: parse_document(lines, "15_incantesimi.md"),
    ),
    WorkItem(
        "ita/16_incantesimi_items.md",
        "documenti",
        lambda lines: parse_document(lines, "16_incantesimi_items.md"),
    ),
    WorkItem(
        "ita/17_glossario_regole.md",
        "documenti",
        lambda lines: parse_document(lines, "17_glossario_regole.md"),
    ),
    WorkItem(
        "ita/18_strumenti_gioco.md",
        "documenti",
        lambda lines: parse_document(lines, "18_strumenti_gioco.md"),
    ),
    WorkItem(
        "ita/19_mostri.md",
        "documenti",
        lambda lines: parse_document(lines, "19_mostri.md"),
    ),
    WorkItem(
        "ita/20_mostri_items.md",
        "documenti",
        lambda lines: parse_document(lines, "20_mostri_items.md"),
    ),
    WorkItem(
        "ita/21_animali.md",
        "documenti",
        lambda lines: parse_document(lines, "21_animali.md"),
    ),
    # Structured classi
    WorkItem("ita/04_classi.md", "classi", parse_classes),
    # Structured backgrounds
    WorkItem("ita/05_origini_personaggio.md", "backgrounds", parse_backgrounds),
    # English document pages (full-page ingestion)
    WorkItem(
        "eng/01_legal_information.md",
        "documenti_en",
        lambda lines: parse_document(lines, "01_legal_information.md"),
    ),
    WorkItem(
        "eng/02_playing_the_game.md",
        "documenti_en",
        lambda lines: parse_document(lines, "02_playing_the_game.md"),
    ),
    WorkItem(
        "eng/03_character_creation.md",
        "documenti_en",
        lambda lines: parse_document(lines, "03_character_creation.md"),
    ),
    WorkItem(
        "eng/04_classes.md",
        "documenti_en",
        lambda lines: parse_document(lines, "04_classes.md"),
    ),
    WorkItem(
        "eng/05_character_origins.md",
        "documenti_en",
        lambda lines: parse_document(lines, "05_character_origins.md"),
    ),
    WorkItem(
        "eng/06_feats.md",
        "documenti_en",
        lambda lines: parse_document(lines, "06_feats.md"),
    ),
    WorkItem(
        "eng/07_equipment_rules.md",
        "documenti_en",
        lambda lines: parse_document(lines, "07_equipment_rules.md"),
    ),
    # Items split (align numero_di_pagina to Italian list)
    WorkItem(
        "eng/07_adventuring_gear.md",
        "documenti_en",
        lambda lines: parse_document(lines, "08_adventuring_gear.md"),
    ),
    WorkItem(
        "eng/07_weapons_items.md",
        "documenti_en",
        lambda lines: parse_document(lines, "09_weapons_items.md"),
    ),
    WorkItem(
        "eng/07_magic_items.md",
        "documenti_en",
        lambda lines: parse_document(lines, "10_magic_items.md"),
    ),
    WorkItem(
        "eng/07_armor_items.md",
        "documenti_en",
        lambda lines: parse_document(lines, "11_armor_items.md"),
    ),
    WorkItem(
        "eng/07_tools_items.md",
        "documenti_en",
        lambda lines: parse_document(lines, "12_tools_items.md"),
    ),
    WorkItem(
        "eng/07_services_items.md",
        "documenti_en",
        lambda lines: parse_document(lines, "13_services_items.md"),
    ),
    WorkItem(
        "eng/07_mounts_vehicles_items.md",
        "documenti_en",
        lambda lines: parse_document(lines, "14_mounts_vehicles_items.md"),
    ),
    WorkItem(
        "eng/08_spells.md",
        "documenti_en",
        lambda lines: parse_document(lines, "15_spells.md"),
    ),
    WorkItem(
        "eng/08_spells_items.md",
        "documenti_en",
        lambda lines: parse_document(lines, "16_spells_items.md"),
    ),
    WorkItem(
        "eng/09_rules_glossary.md",
        "documenti_en",
        lambda lines: parse_document(lines, "17_rules_glossary.md"),
    ),
    WorkItem(
        "eng/10_gameplay_toolbox.md",
        "documenti_en",
        lambda lines: parse_document(lines, "18_gameplay_toolbox.md"),
    ),
    WorkItem(
        "eng/12_monsters_rules.md",
        "documenti_en",
        lambda lines: parse_document(lines, "19_monsters_rules.md"),
    ),
    WorkItem(
        "eng/13_monsters_items.md",
        "documenti_en",
        lambda lines: parse_document(lines, "20_monsters_items.md"),
    ),
    WorkItem(
        "eng/14_animals_items.md",
        "documenti_en",
        lambda lines: parse_document(lines, "21_animals_items.md"),
    ),
]
