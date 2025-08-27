from __future__ import annotations

import math
from typing import Any, Dict, Mapping

from .query_service import build_filter, alpha_sort_expr


async def list_page(repo, collection: str, params: Mapping[str, str], q: str, page: int, page_size: int) -> Dict[str, Any]:
    filt = build_filter(q, collection, params)
    total = await repo.count(collection, filt)
    pages = max(1, math.ceil(total / page_size))
    page = max(1, min(page, pages))
    pipe = [
        {"$match": filt},
        {"$addFields": {"_sortkey": alpha_sort_expr()}},
        {"$sort": {"_sortkey": 1, "_id": 1}},
        {"$skip": (page - 1) * page_size},
        {"$limit": page_size},
    ]
    items = await repo.aggregate_list(collection, pipe)
    return {"items": items, "page": page, "pages": pages, "total": total}

