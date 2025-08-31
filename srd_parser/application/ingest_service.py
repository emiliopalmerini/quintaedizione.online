from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Callable, Dict, Iterable, List, Optional, Protocol


class Repository(Protocol):
    def upsert_many(self, collection: str, unique_fields: List[str], docs: Iterable[Dict]) -> int: ...


def unique_keys_for(collection: str) -> List[str]:
    mapping = {
        "documenti": ["slug"],
        "documenti_en": ["slug"],
        "classi": ["slug"],
        "backgrounds": ["slug"],
        "incantesimi": ["slug"],
        "spells_en": ["slug"],
        "armi": ["slug"],
        "weapons_en": ["slug"],
        "armature": ["slug"],
        "armor_en": ["slug"],
        "strumenti": ["slug"],
        "tools_en": ["slug"],
        "servizi": ["slug"],
        "services_en": ["slug"],
        "equipaggiamento": ["slug"],
        "adventuring_gear_en": ["slug"],
        "oggetti_magici": ["slug"],
        "magic_items_en": ["slug"],
        "mostri": ["slug"],
        "monsters_en": ["slug"],
        "animali": ["slug"],
        "animals_en": ["slug"],
        "talenti": ["slug"],
        "feats_en": ["slug"],
    }
    return mapping.get(collection, ["slug"])


@dataclass
class WorkItem:
    filename: str
    collection: str
    parser: Callable[[List[str]], List[Dict]]


def read_lines(path: Path) -> List[str]:
    return path.read_text(encoding="utf-8").splitlines()
