from __future__ import annotations

from typing import Any, Dict, Mapping, Optional, Tuple

from .query_service import build_filter, neighbors_alpha_repo


async def show_doc(repo, collection: str, doc_id: str, params: Mapping[str, str], q: Optional[str], *, logical_collection: str | None = None) -> Tuple[Optional[Dict[str, Any]], Optional[str], Optional[str], str]:
    doc = await repo.find_by_id(collection, doc_id)
    if not doc:
        return None, None, None, ""
    # Build title and navigation key
    doc_title = str(
        doc.get("name")
        or doc.get("term")
        or doc.get("title")
        or doc.get("titolo")
        or doc.get("nome")
        or doc_id
    )
    cur_key = str(
        doc.get("slug")
        or doc.get("name")
        or doc.get("term")
        or doc.get("title")
        or doc.get("titolo")
        or doc.get("nome")
        or ""
    )
    coll_for_filter = logical_collection or collection
    filt_nav = build_filter(q or "", coll_for_filter, params)
    prev_id, next_id = await neighbors_alpha_repo(repo, collection, cur_key, filt_nav)
    return doc, prev_id, next_id, doc_title
