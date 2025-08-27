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
from .parsers.classes import parse_classes
from .parsers.backgrounds import parse_backgrounds
from .parsers.documents import parse_document

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
    # Document pages (Italian)
    WorkItem("ita/01_informazioni_legali.md", "documenti", lambda lines: parse_document(lines, "01_informazioni_legali.md")),
    WorkItem("ita/02_giocare_il_gioco.md", "documenti", lambda lines: parse_document(lines, "02_giocare_il_gioco.md")),
    WorkItem("ita/03_creazione_personaggio.md", "documenti", lambda lines: parse_document(lines, "03_creazione_personaggio.md")),
    WorkItem("ita/04_classi.md", "documenti", lambda lines: parse_document(lines, "04_classi.md")),
    WorkItem("ita/05_origini_personaggio.md", "documenti", lambda lines: parse_document(lines, "05_origini_personaggio.md")),
    WorkItem("ita/06_talenti.md", "documenti", lambda lines: parse_document(lines, "06_talenti.md")),
    WorkItem("ita/07_equipaggiamento.md", "documenti", lambda lines: parse_document(lines, "07_equipaggiamento.md")),
    WorkItem("ita/08_equipaggiamento_items.md", "documenti", lambda lines: parse_document(lines, "08_equipaggiamento_items.md")),
    WorkItem("ita/09_armi_items.md", "documenti", lambda lines: parse_document(lines, "09_armi_items.md")),
    WorkItem("ita/10_oggetti_magici_items.md", "documenti", lambda lines: parse_document(lines, "10_oggetti_magici_items.md")),
    WorkItem("ita/11_armatura_items.md", "documenti", lambda lines: parse_document(lines, "11_armatura_items.md")),
    WorkItem("ita/12_strumenti_items.md", "documenti", lambda lines: parse_document(lines, "12_strumenti_items.md")),
    WorkItem("ita/13_servizi_items.md", "documenti", lambda lines: parse_document(lines, "13_servizi_items.md")),
    WorkItem("ita/14_cavalcature_veicoli_items.md", "documenti", lambda lines: parse_document(lines, "14_cavalcature_veicoli_items.md")),
    WorkItem("ita/15_incantesimi.md", "documenti", lambda lines: parse_document(lines, "15_incantesimi.md")),
    WorkItem("ita/16_incantesimi_items.md", "documenti", lambda lines: parse_document(lines, "16_incantesimi_items.md")),
    WorkItem("ita/17_glossario_regole.md", "documenti", lambda lines: parse_document(lines, "17_glossario_regole.md")),
    WorkItem("ita/18_strumenti_gioco.md", "documenti", lambda lines: parse_document(lines, "18_strumenti_gioco.md")),
    WorkItem("ita/19_mostri.md", "documenti", lambda lines: parse_document(lines, "19_mostri.md")),
    WorkItem("ita/20_mostri_items.md", "documenti", lambda lines: parse_document(lines, "20_mostri_items.md")),
    WorkItem("ita/21_animali.md", "documenti", lambda lines: parse_document(lines, "21_animali.md")),
    # Structured classi
    WorkItem("ita/04_classi.md", "classi", parse_classes),
    # Structured backgrounds
    WorkItem("ita/05_origini_personaggio.md", "backgrounds", parse_backgrounds),
]

def unique_keys_for(collection: str) -> List[str]:
    mapping = {
        "documenti": ["slug"],
        # For classi we key on slug (stable)
        "classi": ["slug"],
        # Backgrounds keyed on slug
        "backgrounds": ["slug"],
    }
    return mapping.get(collection, ["slug"]) 

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
