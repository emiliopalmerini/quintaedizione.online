from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, Iterable, List, Optional, Protocol

from .ingest_service import WorkItem, read_lines, unique_keys_for


class Repository(Protocol):
    def upsert_many(self, collection: str, unique_fields: List[str], docs: Iterable[Dict]) -> int: ...


@dataclass
class IngestResult:
    collection: str
    filename: str
    parsed: int
    written: int
    preview: Optional[str] = None
    error: Optional[str] = None


def filter_work(items: List[WorkItem], only: Optional[List[str]]) -> List[WorkItem]:
    if not only:
        return items
    wanted = set(only)
    return [w for w in items if w.collection in wanted]


def run_ingest(
    base_dir: Path,
    work_items: List[WorkItem],
    repo: Optional[Repository],
    *,
    dry_run: bool = True,
) -> List[IngestResult]:
    results: List[IngestResult] = []
    for w in work_items:
        path = base_dir / w.filename
        if not path.exists():
            results.append(
                IngestResult(
                    collection=w.collection,
                    filename=path.name,
                    parsed=0,
                    written=0,
                    error=f"Missing file: {path}",
                )
            )
            continue
        try:
            lines = read_lines(path)
            docs = w.parser(lines)
        except Exception as e:
            results.append(
                IngestResult(
                    collection=w.collection,
                    filename=path.name,
                    parsed=0,
                    written=0,
                    error=str(e),
                )
            )
            continue

        if dry_run or repo is None:
            preview_keys = ("name", "term", "level", "rarity", "type", "school", "nome", "titolo")
            preview = [
                {k: d.get(k) for k in preview_keys if k in d}
                for d in docs[:5]
            ]
            results.append(
                IngestResult(
                    collection=w.collection,
                    filename=path.name,
                    parsed=len(docs),
                    written=0,
                    preview=json.dumps(preview, ensure_ascii=False),
                )
            )
        else:
            written = repo.upsert_many(w.collection, unique_keys_for(w.collection), docs)
            results.append(
                IngestResult(
                    collection=w.collection,
                    filename=path.name,
                    parsed=len(docs),
                    written=written,
                )
            )
    return results

