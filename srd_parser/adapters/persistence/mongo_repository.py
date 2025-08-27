from __future__ import annotations

from typing import Dict, Iterable, List

from pymongo import ASCENDING
from pymongo.collection import Collection
from pymongo.database import Database
from pymongo.errors import PyMongoError

from ...utils import source_label


class MongoRepository:
    def __init__(self, db: Database):
        self._db = db

    def _collection(self, name: str) -> Collection:
        return self._db[name]

    def _ensure_unique_index(self, col: Collection, unique_fields: List[str]) -> None:
        if not unique_fields:
            return
        try:
            col.create_index(
                [(f, ASCENDING) for f in unique_fields],
                name="uq_" + "_".join(unique_fields),
                unique=True,
                background=False,
            )
        except PyMongoError:
            # Index creation failures are non-fatal for ingest
            pass

    def upsert_many(
        self, collection: str, unique_fields: List[str], docs: Iterable[Dict]
    ) -> int:
        col = self._collection(collection)
        self._ensure_unique_index(col, unique_fields)
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
            except PyMongoError:
                # Skip failed upserts; caller can log if needed
                continue
        return n
