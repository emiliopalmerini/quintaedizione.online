from __future__ import annotations

from typing import Any, Dict, List, Optional


async def load_home_document(repo, page: Optional[int], *, collection: str = "documenti") -> Dict[str, Any]:
    out: Dict[str, Any] = {
        "doc": None,
        "prev_page": None,
        "next_page": None,
        "prev_title": None,
        "next_title": None,
        "pages_list": [],
        "pages_items": [],
        "cur_page": None,
    }
    try:
        # Current page
        if page is None:
            cur = await repo.find_one_sorted(
                collection, {}, sort=[("numero_di_pagina", 1), ("_id", 1)]
            )
        else:
            cur = await repo.find_one(collection, {"numero_di_pagina": page})
            if not cur:
                cur = await repo.find_one_sorted(
                    collection, {}, sort=[("numero_di_pagina", 1), ("_id", 1)]
                )
        if cur:
            out["doc"] = {**cur, "_id": str(cur["_id"]) }
            cur_page = int(cur.get("numero_di_pagina") or 0)
            out["cur_page"] = cur_page

            # Prev and next
            prev = await repo.find_one_sorted(
                collection,
                {"numero_di_pagina": {"$lt": cur_page}},
                sort=[("numero_di_pagina", -1), ("_id", -1)],
                projection={"numero_di_pagina": 1, "titolo": 1, "title": 1, "slug": 1},
            )
            if prev:
                out["prev_page"] = int(prev.get("numero_di_pagina") or 0)
                out["prev_title"] = str(prev.get("titolo") or prev.get("title") or prev.get("slug") or out["prev_page"])  # type: ignore[index]

            nxt = await repo.find_one_sorted(
                collection,
                {"numero_di_pagina": {"$gt": cur_page}},
                sort=[("numero_di_pagina", 1), ("_id", 1)],
                projection={"numero_di_pagina": 1, "titolo": 1, "title": 1, "slug": 1},
            )
            if nxt:
                out["next_page"] = int(nxt.get("numero_di_pagina") or 0)
                out["next_title"] = str(nxt.get("titolo") or nxt.get("title") or nxt.get("slug") or out["next_page"])  # type: ignore[index]

            # Pages list via aggregation
            pipeline = [
                {"$project": {"numero_di_pagina": 1, "titolo": 1, "title": 1, "slug": 1}},
                {"$sort": {"numero_di_pagina": 1, "_id": 1}},
            ]
            items = await repo.aggregate_list(collection, pipeline)
            pages_items: List[Dict[str, Any]] = []
            pages_list: List[int] = []
            for d in items:
                p = d.get("numero_di_pagina")
                if p is None:
                    continue
                try:
                    p_int = int(p)
                except Exception:
                    continue
                if p_int in pages_list:
                    continue
                title = str(d.get("titolo") or d.get("title") or d.get("slug") or p_int)
                pages_items.append({"page": p_int, "title": title})
                pages_list.append(p_int)
            out["pages_items"] = pages_items
            out["pages_list"] = pages_list
    except Exception:
        pass
    return out

