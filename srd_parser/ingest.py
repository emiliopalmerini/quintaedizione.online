# srd_parser/ingest.py
from __future__ import annotations

import argparse
import logging
import os
from pathlib import Path
from typing import List

from pymongo import MongoClient

from .adapters.persistence.mongo_repository import MongoRepository
from .application.ingest_runner import filter_work, run_ingest
from .work import DEFAULT_WORK

LOG_LEVEL = os.environ.get("LOG_LEVEL", "INFO").upper()
logging.basicConfig(
    level=getattr(logging, LOG_LEVEL, logging.INFO),
    format="%(levelname)s %(message)s",
)
log = logging.getLogger("srd-ingest")

## work items are defined in srd_parser/work.py

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
    repo = MongoRepository(db)

    target = filter_work(DEFAULT_WORK, args.only)
    results = run_ingest(base, target, None if args.dry_run else repo, dry_run=args.dry_run)
    total_written = sum(r.written for r in results)
    for r in results:
        if r.error:
            log.warning("%s %s: %s", r.collection, r.filename, r.error)
            continue
        log.info("Parsed %d from %s → %s", r.parsed, r.filename, r.collection)
        if args.dry_run and r.preview:
            log.info("Preview: %s", r.preview)
        elif not args.dry_run:
            log.info("Upserted %d docs into %s.%s", r.written, args.db_name, r.collection)
    if not args.dry_run:
        log.info("Done. Total upserts: %d", total_written)

if __name__ == "__main__":
    main()
