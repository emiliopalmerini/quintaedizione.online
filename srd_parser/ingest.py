# srd_parser/ingest.py
from __future__ import annotations

import argparse
import json
import logging
import os
from dataclasses import dataclass
from pathlib import Path
from typing import Callable, Dict, Iterable, List

from pymongo import ASCENDING, MongoClient
from pymongo.collection import Collection
from pymongo.errors import PyMongoError

from .utils import source_label
from .parsers.spells import parse_spells
from .parsers.magic_items import parse_magic_items
from .parsers.equipment import parse_equipment
from .parsers.rules import parse_rules_glossary
from .parsers.monsters import parse_monsters
from .parsers.classes import parse_classes

LOG_LEVEL = os.environ.get("LOG_LEVEL", "INFO").upper()
logging.basicConfig(
    level=getattr(logging, LOG_LEVEL, logging.INFO),
    format="%(levelname)s %(message)s",
)
log = logging.getLogger("srd-ingest")

@dataclass
class WorkItem:
    filename: str
    collection: str
    parser: Callable[[List[str]], List[Dict]]

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
    WorkItem("04_classes.md", "classes", parse_classes),
]

def unique_keys_for(collection: str) -> List[str]:
    mapping = {
        "spells": ["name", "level"],
        "rules_glossary": ["term"],
        "magic_items": ["name"],
        "armor": ["name"],
        "weapons": ["name"],
        "tools": ["name"],
        "mounts_vehicles": ["name"],
        "services": ["name"],
        "monsters": ["name"],
        "animals": ["name"],
        "classes": ["name"],
    }
    return mapping.get(collection, ["name"])

def _create_unique_index(col: Collection, unique_fields: List[str]) -> None:
    if not unique_fields:
        return
    try:
        col.create_index([(f, ASCENDING) for f in unique_fields],
                         name="uq_" + "_".join(unique_fields),
                         unique=True,
                         background=False)
    except PyMongoError as e:
        log.warning("Index create failed for %s: %s", col.name, e)

def upsert_many(col: Collection, unique_fields: List[str], docs: Iterable[Dict]) -> int:
    _create_unique_index(col, unique_fields)
    n = 0
    src = source_label()
    for d in docs:
        doc = {**d, "source": src}
        try:
            col.update_one(
                {k: doc[k] for k in unique_fields},
                {"$set": doc, "$setOnInsert": {"_source": "markdown"}},
                upsert=True,
            )
            n += 1
        except PyMongoError as e:
            ident = {k: doc.get(k) for k in unique_fields}
            log.error("Upsert failed for %s %s: %s", col.name, ident, e)
    return n

def _read_lines(path: Path) -> List[str]:
    try:
        return path.read_text(encoding="utf-8").splitlines()
    except Exception as e:
        log.error("Failed to read %s: %s", path, e)
        return []

def main() -> None:
    ap = argparse.ArgumentParser(description="Parse SRD markdown files and upsert into MongoDB.")
    ap.add_argument("--input-dir", default=os.environ.get("INPUT_DIR", "/data"))
    ap.add_argument("--mongo-uri", default=os.environ.get("MONGO_URI", "mongodb://localhost:27017"))
    ap.add_argument("--db-name", default=os.environ.get("DB_NAME", "dnd"))
    ap.add_argument("--dry-run", action="store_true", default=os.environ.get("DRY_RUN", "0") == "1")
    ap.add_argument("--only", nargs="*", help="Limit to collections (e.g. spells monsters classes)")
    args = ap.parse_args()

    base = Path(args.input_dir)
    if not base.exists():
        log.error("Input dir not found: %s", base)
        raise SystemExit(2)

    client = MongoClient(args.mongo_uri)
    db = client[args.db_name]

    target = [w for w in DEFAULT_WORK if not args.only or w.collection in set(args.only)]

    total_written = 0
    for w in target:
        path = base / w.filename
        if not path.exists():
            log.warning("Missing file: %s", path)
            continue
        log.info("Parsing %s → %s", path.name, w.collection)
        lines = _read_lines(path)
        try:
            docs = w.parser(lines)
        except Exception as e:
            log.error("Parser error in %s: %s", path.name, e)
            continue
        log.info("Parsed %d docs from %s", len(docs), path.name)
        if args.dry_run:
            preview_keys = ("name", "term", "level", "rarity", "type", "school")
            preview = [{k: d.get(k) for k in preview_keys if k in d} for d in docs[:5]]
            log.info("Preview: %s", json.dumps(preview, ensure_ascii=False))
            continue
        col = db[w.collection]
        written = upsert_many(col, unique_keys_for(w.collection), docs)
        total_written += written
        log.info("Upserted %d docs into %s.%s", written, args.db_name, w.collection)

    if not args.dry_run:
        log.info("Done. Total upserts: %d", total_written)

if __name__ == "__main__":
    main()

