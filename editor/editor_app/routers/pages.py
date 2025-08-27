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
from editor_app.application.query_service import build_filter, alpha_sort_expr
from editor_app.application.list_service import list_page as svc_list_page
from editor_app.application.show_service import show_doc as svc_show_doc
from editor_app.core.templates import env
from editor_app.application.home_service import load_home_document as svc_home_doc
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

    # Carica un documento da 'documenti' per la homepage via service
    doc_data = await svc_home_doc(MongoRepository(db), page)
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


@router.get("/home/doc", response_class=HTMLResponse)
# legacy path kept empty; homepage renders doc via service


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
    res = await svc_list_page(repo, collection, request.query_params, q, page, page_size)
    items = res["items"]
    pages = res["pages"]
    page = res["page"]
    total = res["total"]
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
    doc, prev_id, next_id, doc_title = await svc_show_doc(repo, collection, doc_id, request.query_params, q)
    if not doc:
        raise HTTPException(404, "Documento non trovato")
    fields = flatten_for_form(doc)

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
