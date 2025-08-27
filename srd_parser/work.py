from __future__ import annotations

from dataclasses import dataclass
from typing import Callable, List

from .parsers.classes import parse_classes
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
]
