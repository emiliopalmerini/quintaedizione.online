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
    }
    return mapping.get(collection, ["slug"])


@dataclass
class WorkItem:
    filename: str
    collection: str
    parser: Callable[[List[str]], List[Dict]]


def read_lines(path: Path) -> List[str]:
    return path.read_text(encoding="utf-8").splitlines()
