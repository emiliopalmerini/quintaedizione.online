# app/routers/pages.py
from __future__ import annotations

import json
import math
from typing import Any, Dict, Mapping, Optional, Tuple
from urllib.parse import urlencode

from bson import ObjectId
from editor_app.core.config import COLLECTION_LABELS, COLLECTIONS
from editor_app.core.db import get_db
from editor_app.adapters.persistence.mongo_repository import MongoRepository
from editor_app.core.flatten import flatten_for_form
from editor_app.utils.markdown import render_md
from editor_app.application.query_service import build_filter, neighbors_alpha, alpha_sort_expr
from editor_app.core.templates import env
from editor_app.core.transform import to_jsonable
from fastapi import APIRouter, HTTPException, Query, Request
from fastapi.responses import HTMLResponse, PlainTextResponse

router = APIRouter()

# ---- helpers ---------------------------------------------------------------


## moved helpers to application.query_service


# ---- pages -----------------------------------------------------------------


@router.get("/", response_class=HTMLResponse)
async def index(page: int | None = Query(default=None)) -> HTMLResponse:
    tpl = env.get_template("index.html")
    # Mostra in home solo le collezioni non marcate come EN
    visible_cols = [c for c in COLLECTIONS if "(EN)" not in COLLECTION_LABELS.get(c, "")]
    cols_sorted = sorted(visible_cols, key=lambda c: COLLECTION_LABELS.get(c, c).lower())
    counts: Dict[str, int] = {}
    db = await get_db()
    for c in cols_sorted:
        try:
            counts[c] = await db[c].count_documents({})
        except Exception:
            counts[c] = 0
    total = sum(counts.values()) if counts else 0

    # Carica un documento da 'documenti' per la homepage
    doc_data = await _load_home_document(db, page)
    # Renderizza HTML per la prima visualizzazione (coerente con la partial)
    doc_html = ""
    if doc_data.get("doc") and doc_data["doc"].get("content"):
        doc_html = render_md(str(doc_data["doc"].get("content") or ""))

    return HTMLResponse(
        tpl.render(
            collections=cols_sorted,
            labels=COLLECTION_LABELS,
            counts=counts,
            total=total,
            doc=to_jsonable(doc_data.get("doc")) if doc_data.get("doc") else None,
            doc_html=doc_html,
            prev_page=doc_data.get("prev_page"),
            next_page=doc_data.get("next_page"),
            prev_title=doc_data.get("prev_title"),
            next_title=doc_data.get("next_title"),
            pages_list=doc_data.get("pages_list"),
            pages_items=doc_data.get("pages_items"),
            cur_page=doc_data.get("cur_page"),
        )
    )


async def _load_home_document(db, page: int | None) -> Dict[str, Any]:
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
        doc_col = db["documenti"]
        # Trova pagina corrente: se non indicata, prendi la più piccola
        if page is None:
            cur = await doc_col.find_one({}, sort=[("numero_di_pagina", 1), ("_id", 1)])
        else:
            cur = await doc_col.find_one({"numero_di_pagina": page})
            if not cur:
                cur = await doc_col.find_one({}, sort=[("numero_di_pagina", 1), ("_id", 1)])
        if cur:
            out["doc"] = {**cur, "_id": str(cur["_id"]) }
            cur_page = int(cur.get("numero_di_pagina") or 0)
            out["cur_page"] = cur_page
            # Prev
            prev = await doc_col.find_one(
                {"numero_di_pagina": {"$lt": cur_page}}, sort=[("numero_di_pagina", -1), ("_id", -1)]
            )
            if prev:
                out["prev_page"] = int(prev.get("numero_di_pagina") or 0)
                out["prev_title"] = str(prev.get("titolo") or prev.get("title") or prev.get("slug") or out["prev_page"])  # type: ignore[index]
            # Next
            nxt = await doc_col.find_one(
                {"numero_di_pagina": {"$gt": cur_page}}, sort=[("numero_di_pagina", 1), ("_id", 1)]
            )
            if nxt:
                out["next_page"] = int(nxt.get("numero_di_pagina") or 0)
                out["next_title"] = str(nxt.get("titolo") or nxt.get("title") or nxt.get("slug") or out["next_page"])  # type: ignore[index]
            # Pages list
            # Build pages list and titles for tooltips
            try:
                # Fetch all pages with titles
                cursor = doc_col.find(
                    {},
                    projection={"numero_di_pagina": 1, "titolo": 1, "title": 1, "slug": 1},
                    sort=[("numero_di_pagina", 1), ("_id", 1)],
                )
                pages_items = []
                pages_list: list[int] = []
                async for d in cursor:
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
                out["pages_items"] = []
                try:
                    pages = await doc_col.distinct("numero_di_pagina")
                    pages = [int(p) for p in pages if p is not None]
                    pages.sort()
                    out["pages_list"] = pages
                except Exception:
                    out["pages_list"] = []
    except Exception:
        pass
    return out


@router.get("/home/doc", response_class=HTMLResponse)
# removed unused partial route; index renders homepage doc directly


@router.get("/list/{collection}", response_class=HTMLResponse)
async def list_page(
    request: Request, collection: str, q: str = "", page: int = 1, page_size: int = 20
) -> HTMLResponse:
    if collection not in COLLECTIONS:
        raise HTTPException(404)
    tpl = env.get_template("list.html")
    return HTMLResponse(
        tpl.render(
            collection=collection, q=q, page=page, page_size=page_size, request=request
        )
    )


@router.get("/view/{collection}", response_class=HTMLResponse)
async def view_rows(
    request: Request, collection: str, q: str = "", page: int = 1, page_size: int = 20
) -> HTMLResponse:
    if collection not in COLLECTIONS:
        raise HTTPException(404)
    db = await get_db()
    repo = MongoRepository(db)
    filt = build_filter(q, collection, request.query_params)
    total = await repo.count(collection, filt)
    pages = max(1, math.ceil(total / page_size))
    page = max(1, min(page, pages))
    # sort by alpha of name/term via aggregation
    pipe = [
        {"$match": filt},
        {"$addFields": {"_sortkey": alpha_sort_expr()}},
        {"$sort": {"_sortkey": 1, "_id": 1}},
        {"$skip": (page - 1) * page_size},
        {"$limit": page_size},
    ]
    items = await repo.aggregate_list(collection, pipe)
    for doc in items:
        if doc.get("_id") is not None:
            doc["_id"] = str(doc["_id"])
    tpl = env.get_template("_rows.html")
    qs = (
        urlencode(dict(request.query_params))
        if request and request.query_params
        else ""
    )
    return HTMLResponse(
        tpl.render(
            collection=collection,
            items=items,
            page=page,
            pages=pages,
            total=total,
            page_size=page_size,
            q=q,
            qs=qs,
        )
    )


@router.get("/quicksearch/{collection}", response_class=HTMLResponse)
async def quicksearch(request: Request, collection: str, q: str = "") -> HTMLResponse:
    if collection not in COLLECTIONS:
        raise HTTPException(404)
    tpl = env.get_template("_quicksearch.html")
    if not q.strip():
        return HTMLResponse(tpl.render(collection=collection, q=q, items=[]))
    db = await get_db()
    repo = MongoRepository(db)
    # Quick mode: prefisso su name/term/title
    filt = build_filter(q, collection, request.query_params, quick=True)
    pipe = [
        {"$match": filt},
        {"$addFields": {"_sortkey": alpha_sort_expr()}},
        {"$sort": {"_sortkey": 1, "_id": 1}},
        {"$limit": 10},
        {"$project": {"_id": 1, "name": 1, "term": 1, "title": 1, "titolo": 1, "nome": 1}},
    ]
    items = await repo.aggregate_list(collection, pipe)
    for d in items:
        if d.get("_id") is not None:
            d["_id"] = str(d["_id"])
    return HTMLResponse(tpl.render(collection=collection, q=q, items=items))


@router.get("/show/{collection}/{doc_id}", response_class=HTMLResponse)
async def show_doc(
    request: Request, collection: str, doc_id: str, q: str | None = Query(default=None)
) -> HTMLResponse:
    if collection not in COLLECTIONS:
        raise HTTPException(404)
    db = await get_db()
    repo = MongoRepository(db)
    doc = await repo.find_by_id(collection, doc_id)
    if not doc:
        raise HTTPException(404, "Documento non trovato")

    fields = flatten_for_form(doc)
    filt_nav = build_filter(q or "", collection, request.query_params)
    cur_key = str(
        doc.get("slug")
        or doc.get("name")
        or doc.get("term")
        or doc.get("title")
        or doc.get("titolo")
        or doc.get("nome")
        or ""
    )
    # neighbors still use collection aggregate; pass motor collection directly for now
    col = db[collection]
    prev_id, next_id = await neighbors_alpha(col, cur_key, filt_nav)
    doc_title = str(
        doc.get("name")
        or doc.get("term")
        or doc.get("title")
        or doc.get("titolo")
        or doc.get("nome")
        or doc_id
    )

    if collection in ("classi", "classes"):
        tpl_name = "show_class.html"
    elif collection in ("backgrounds",):
        tpl_name = "show_background.html"
    else:
        tpl_name = "show.html"
    tpl = env.get_template(tpl_name)
    qs = (
        urlencode(dict(request.query_params))
        if request and request.query_params
        else ""
    )
    # Server-side markdown render for document body
    body_raw: str | None = None
    body_html: str = ""
    cand = (
        doc.get("description")
        or doc.get("description_md")
        or doc.get("content")
    )
    if not cand:
        for k, v in doc.items():
            if isinstance(k, str) and k.endswith("_md") and v:
                cand = v
                break
    body_raw = str(cand) if cand is not None else None
    if body_raw:
        body_html = render_md(body_raw)

    return HTMLResponse(
        tpl.render(
            collection=collection,
            doc_id=str(doc["_id"]),
            doc_title=doc_title,
            doc_slug=str(doc.get("slug") or ""),
            fields=fields,
            doc_obj=to_jsonable(doc),
            body_raw=body_raw or "",
            body_html=body_html,
            q=q or "",
            prev_id=prev_id,
            next_id=next_id,
            request=request,
            qs=qs,
        )
    )


# Rimosso editor per-campi: mantenuta solo la modalità JSON


## editing removed
