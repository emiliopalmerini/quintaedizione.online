"""Edition-specific section configs mapping PDF page ranges to collections.

Each section config entry maps an output filename to:
  - header: the H1 header text to insert in the output file
  - pages: (start, end) 1-indexed inclusive page range in the PDF
"""

from dataclasses import dataclass


@dataclass(frozen=True)
class Section:
    header: str
    pages: tuple[int, int]  # 1-indexed, inclusive


# SRD v5.2.1 Italian — 405 pages total.
# Page numbers taken from the PDF table of contents.
SECTIONS_5_2_1: dict[str, Section] = {
    "classi.md": Section("Classi", (32, 92)),
    "backgrounds.md": Section("Backgrounds", (93, 97)),
    "talenti.md": Section("Talenti", (98, 100)),
    "armi.md": Section("Armi", (101, 103)),
    "armature.md": Section("Armature", (104, 105)),
    "strumenti.md": Section("Strumenti", (106, 107)),
    "equipaggiamenti.md": Section("Equipaggiamenti", (108, 112)),
    "cavalcature_veicoli_items.md": Section("Cavalcature e Veicoli", (113, 114)),
    "servizi.md": Section("Servizi", (115, 117)),
    "incantesimi.md": Section("Incantesimi", (118, 201)),
    "regole.md": Section("Regole", (202, 231)),
    "oggetti_magici.md": Section("Oggetti Magici", (232, 288)),
    "mostri.md": Section("Mostri", (289, 384)),
    "animali.md": Section("Animali", (385, 405)),
}
