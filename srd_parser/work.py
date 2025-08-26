from __future__ import annotations

from dataclasses import dataclass
from typing import List, Callable

from .parsers.spells import parse_spells
from .parsers.magic_items import parse_magic_items
from .parsers.equipment import parse_equipment
from .parsers.rules import parse_rules_glossary
from .parsers.monsters import parse_monsters
from .parsers.classes import parse_classes


@dataclass
class WorkItem:
    filename: str
    collection: str
    parser: Callable


DEFAULT_WORK: List[WorkItem] = [
    WorkItem("08_b_spellsaz.md", "spells", parse_spells),
    WorkItem("07_magic_items.md", "magic_items", parse_magic_items),
    WorkItem("07_armor_items.md", "armor", parse_equipment),
    WorkItem("07_weapons_items.md", "weapons", parse_equipment),
    WorkItem("07_tools_items.md", "tools", parse_equipment),
    WorkItem("07_mounts_vehicles_items.md", "mounts_vehicles", parse_equipment),
    WorkItem("07_services_items.md", "services", parse_equipment),
    WorkItem("09_rules_glossary.md", "rules_glossary", parse_rules_glossary),
    WorkItem("13_monsters_az.md", "monsters", parse_monsters),
    WorkItem("14_animals.md", "animals", parse_monsters),
    # Italian classes file
    WorkItem("ita/04_classi.md", "classes", parse_classes),
]
