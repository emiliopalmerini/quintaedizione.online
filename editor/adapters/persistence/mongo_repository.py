from __future__ import annotations

from typing import Any, Dict, List, Optional

from bson import ObjectId
from motor.motor_asyncio import AsyncIOMotorDatabase


class MongoRepository:
    def __init__(self, db: AsyncIOMotorDatabase):
        self._db = db

    def _col(self, name: str):
        return self._db[name]

    async def count(self, collection: str, filt: Dict[str, Any]) -> int:
        return await self._col(collection).count_documents(filt)

    async def aggregate_list(self, collection: str, pipeline: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        out: List[Dict[str, Any]] = []
        async for d in self._col(collection).aggregate(pipeline):
            out.append(d)
        return out

    async def find_one(self, collection: str, filt: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        d = await self._col(collection).find_one(filt)
        return d

    async def find_by_id(self, collection: str, id_str: str) -> Optional[Dict[str, Any]]:
        try:
            oid = ObjectId(id_str)
        except Exception:
            return None
        return await self.find_one(collection, {"_id": oid})

    async def find_one_sorted(
        self,
        collection: str,
        filt: Dict[str, Any],
        *,
        sort: List[tuple[str, int]],
        projection: Optional[Dict[str, int]] = None,
    ) -> Optional[Dict[str, Any]]:
        return await self._col(collection).find_one(filt, sort=sort, projection=projection)

