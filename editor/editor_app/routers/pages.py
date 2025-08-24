# app/routers/pages.py
from __future__ import annotations

import math
from typing import Optional, Any, Dict, Tuple, Mapping

from bson import ObjectId
from fastapi import APIRouter, HTTPException, Request, Query
from urllib.parse import urlencode
from fastapi.responses import HTMLResponse, PlainTextResponse
import json

from editor_app.core.config import COLLECTIONS, COLLECTION_LABELS
from editor_app.core.db import get_db
from editor_app.core.flatten import flatten_for_form
from editor_app.core.transform import to_jsonable
from editor_app.core.templates import env

router = APIRouter()

# ---- helpers ---------------------------------------------------------------

def _rx(val: str) -> Dict[str, Any]:
    return {"$regex": val, "$options": "i"}

def _parse_bool(val: Optional[str]) -> Optional[bool]:
    if val is None: return None
    v = val.strip().lower()
    if v in ("1", "true", "yes", "y", "si", "s"): return True
    if v in ("0", "false", "no", "n"): return False
    return None

def build_filter(q: str | None, collection: str, params: Mapping[str, str]) -> Dict[str, Any]:
    filt: Dict[str, Any] = {}
    if q:
        r = _rx(q)
        filt["$or"] = [{"name": r}, {"term": r}, {"title": r}, {"description": r}]

    if collection == "spells":
        lvl = params.get("level")
        if lvl and lvl.isdigit():
            filt["level"] = int(lvl)
        school = params.get("school")
        if school:
            filt["school"] = _rx(school)
        ritual = _parse_bool(params.get("ritual"))
        if ritual is not None:
            filt["ritual"] = ritual
        classes = params.get("classes")
        if classes:
            filt["classes"] = {"$elemMatch": _rx(classes)}

    elif collection == "magic_items":
        rarity = params.get("rarity")
        if rarity:
            filt["rarity"] = _rx(rarity)
        itype = params.get("type")
        if itype:
            filt["type"] = _rx(itype)
        att = _parse_bool(params.get("attunement"))
        if att is not None:
            filt["attunement"] = att

    elif collection == "monsters":
        size = params.get("size")
        if size:
            filt["size"] = _rx(size)
        mtype = params.get("type")
        if mtype:
            filt["type"] = _rx(mtype)
        align = params.get("alignment")
        if align:
            filt["alignment"] = _rx(align)
        cr = params.get("cr")
        if cr:
            try:
                cr_num = float(cr) if "." in cr or "/" not in cr else None
            except Exception:
                cr_num = None
            if cr_num is not None:
                filt.setdefault("$and", []).append({"$or": [{"challenge_rating": cr_num}, {"cr": cr_num}]})
            else:
                r = _rx(cr)
                filt.setdefault("$and", []).append({"$or": [{"challenge_rating": r}, {"cr": r}]})

    return filt

def _alpha_sort_expr() -> Dict[str, Any]:
    return {"$toLower": {"$ifNull": ["$name", {"$ifNull": ["$term", ""]}]}}

async def _neighbors_alpha(col, cur_key: str, filt: Dict[str, Any]) -> Tuple[Optional[str], Optional[str]]:
    key = (cur_key or "").lower()
    prev_pipe = [
        {"$match": filt},
        {"$addFields": {"_sortkey": _alpha_sort_expr()}},
        {"$match": {"_sortkey": {"$lt": key}}},
        {"$sort": {"_sortkey": -1, "_id": -1}},
        {"$limit": 1},
        {"$project": {"_id": 1}},
    ]
    next_pipe = [
        {"$match": filt},
        {"$addFields": {"_sortkey": _alpha_sort_expr()}},
        {"$match": {"_sortkey": {"$gt": key}}},
        {"$sort": {"_sortkey": 1, "_id": 1}},
        {"$limit": 1},
        {"$project": {"_id": 1}},
    ]
    prev_id: Optional[str] = None
    next_id: Optional[str] = None
    async for d in col.aggregate(prev_pipe):
        prev_id = str(d.get("_id"))
    async for d in col.aggregate(next_pipe):
        next_id = str(d.get("_id"))
    return prev_id, next_id

# ---- pages -----------------------------------------------------------------

@router.get("/", response_class=HTMLResponse)
async def index() -> HTMLResponse:
    tpl = env.get_template("index.html")
    cols_sorted = sorted(COLLECTIONS, key=lambda c: COLLECTION_LABELS.get(c, c).lower())
    first = cols_sorted[0] if cols_sorted else ""
    counts: Dict[str, int] = {}
    db = await get_db()
    for c in cols_sorted:
        try:
            counts[c] = await db[c].count_documents({})
        except Exception:
            counts[c] = 0
    total = sum(counts.values()) if counts else 0
    return HTMLResponse(tpl.render(collections=cols_sorted, labels=COLLECTION_LABELS, counts=counts, total=total, first=first))

@router.get("/list/{collection}", response_class=HTMLResponse)
async def list_page(request: Request, collection: str, q: str = "", page: int = 1, page_size: int = 20) -> HTMLResponse:
    if collection not in COLLECTIONS:
        raise HTTPException(404)
    tpl = env.get_template("list.html")
    return HTMLResponse(tpl.render(collection=collection, q=q, page=page, page_size=page_size, request=request))

@router.get("/view/{collection}", response_class=HTMLResponse)
async def view_rows(request: Request, collection: str, q: str = "", page: int = 1, page_size: int = 20) -> HTMLResponse:
    if collection not in COLLECTIONS:
        raise HTTPException(404)
    db = await get_db()
    col = db[collection]
    filt = build_filter(q, collection, request.query_params)
    total = await col.count_documents(filt)
    pages = max(1, math.ceil(total / page_size))
    page = max(1, min(page, pages))
    # sort by alpha of name/term via aggregation
    pipe = [
        {"$match": filt},
        {"$addFields": {"_sortkey": _alpha_sort_expr()}},
        {"$sort": {"_sortkey": 1, "_id": 1}},
        {"$skip": (page - 1) * page_size},
        {"$limit": page_size},
    ]
    items = []
    async for doc in col.aggregate(pipe):
        doc["_id"] = str(doc["_id"])
        items.append(doc)
    tpl = env.get_template("_rows.html")
    qs = urlencode(dict(request.query_params)) if request and request.query_params else ""
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
    col = db[collection]
    filt = build_filter(q, collection, request.query_params)
    pipe = [
        {"$match": filt},
        {"$addFields": {"_sortkey": _alpha_sort_expr()}},
        {"$sort": {"_sortkey": 1, "_id": 1}},
        {"$limit": 10},
        {"$project": {"_id": 1, "name": 1, "term": 1, "title": 1}},
    ]
    items = []
    async for d in col.aggregate(pipe):
        d["_id"] = str(d["_id"])
        items.append(d)
    return HTMLResponse(tpl.render(collection=collection, q=q, items=items))

@router.get("/edit/{collection}/{doc_id}", response_class=HTMLResponse)
async def doc_form(request: Request, collection: str, doc_id: str, q: str | None = Query(default=None)) -> HTMLResponse:
    if collection not in COLLECTIONS:
        raise HTTPException(404)
    db = await get_db()
    col = db[collection]
    try:
        oid = ObjectId(doc_id)
    except Exception:
        raise HTTPException(400, "invalid _id")

    doc = await col.find_one({"_id": oid})
    if not doc:
        raise HTTPException(404, "Documento non trovato")

    fields = flatten_for_form(doc)
    # Determina un titolo leggibile per breadcrumb
    doc_title = str(doc.get("name") or doc.get("term") or doc.get("title") or doc_id)
    filt_nav = build_filter(q or "", collection, request.query_params)
    cur_key = str(doc.get("name") or doc.get("term") or "")
    prev_id, next_id = await _neighbors_alpha(col, cur_key, filt_nav)

    tpl = env.get_template("_doc_form.html")
    from urllib.parse import urlencode as _urlencode
    qs = _urlencode(dict(request.query_params)) if request and request.query_params else ""
    return HTMLResponse(
        tpl.render(
            collection=collection,
            doc_id=str(doc["_id"]),
            doc_title=doc_title,
            fields=fields,
            q=q or "",
            prev_id=prev_id,
            next_id=next_id,
            request=request,
            qs=qs,
        )
    )

@router.get("/show/{collection}/{doc_id}", response_class=HTMLResponse)
async def show_doc(request: Request, collection: str, doc_id: str, q: str | None = Query(default=None)) -> HTMLResponse:
    if collection not in COLLECTIONS:
        raise HTTPException(404)
    db = await get_db()
    col = db[collection]
    try:
        oid = ObjectId(doc_id)
    except Exception:
        raise HTTPException(400, "invalid _id")

    doc = await col.find_one({"_id": oid})
    if not doc:
        raise HTTPException(404, "Documento non trovato")

    fields = flatten_for_form(doc)
    filt_nav = build_filter(q or "", collection, request.query_params)
    cur_key = str(doc.get("name") or doc.get("term") or "")
    prev_id, next_id = await _neighbors_alpha(col, cur_key, filt_nav)
    doc_title = str(doc.get("name") or doc.get("term") or doc.get("title") or doc_id)

    tpl_name = "show_class.html" if collection == "classes" else "show.html"
    tpl = env.get_template(tpl_name)
    qs = urlencode(dict(request.query_params)) if request and request.query_params else ""
    return HTMLResponse(
        tpl.render(
            collection=collection,
            doc_id=str(doc["_id"]),
            doc_title=doc_title,
            fields=fields,
            doc_obj=to_jsonable(doc),
            q=q or "",
            prev_id=prev_id,
            next_id=next_id,
            request=request,
            qs=qs,
        )
    )

@router.put("/edit/{collection}/{doc_id}", response_class=PlainTextResponse)
async def edit_doc(collection: str, doc_id: str, request: Request) -> PlainTextResponse:
    if collection not in COLLECTIONS:
        raise HTTPException(404)
    form = await request.form()
    update: Dict[str, Any] = {}
    for k, v in form.items():
        if not k.startswith("f."):
            continue
        path = k[2:]
        update[path] = v
    if not update:
        return PlainTextResponse("No changes")

    db = await get_db()
    try:
        oid = ObjectId(doc_id)
    except Exception:
        raise HTTPException(400, "invalid _id")

    await db[collection].update_one({"_id": oid}, {"$set": update})
    return PlainTextResponse("Saved")

@router.get("/edit_raw/{collection}/{doc_id}", response_class=HTMLResponse)
async def edit_raw_get(request: Request, collection: str, doc_id: str, q: str | None = Query(default=None)) -> HTMLResponse:
    if collection not in COLLECTIONS:
        raise HTTPException(404)
    db = await get_db()
    try:
        oid = ObjectId(doc_id)
    except Exception:
        raise HTTPException(400, "invalid _id")
    doc = await db[collection].find_one({"_id": oid})
    if not doc:
        raise HTTPException(404, "Documento non trovato")

    json_str = json.dumps(to_jsonable(doc), ensure_ascii=False, indent=2, sort_keys=True)
    doc_title = str(doc.get("name") or doc.get("term") or doc.get("title") or doc_id)
    filt_nav = build_filter(q or "", collection, request.query_params)
    cur_key = str(doc.get("name") or doc.get("term") or "")
    prev_id, next_id = await _neighbors_alpha(db[collection], cur_key, filt_nav)
    tpl = env.get_template("edit_raw.html")
    qs = urlencode(dict(request.query_params)) if request and request.query_params else ""
    return HTMLResponse(
        tpl.render(
            collection=collection,
            doc_id=str(doc["_id"]),
            doc_title=doc_title,
            raw_json=json_str,
            request=request,
            q=q or "",
            prev_id=prev_id,
            next_id=next_id,
            qs=qs,
        )
    )

@router.put("/edit_raw/{collection}/{doc_id}", response_class=PlainTextResponse)
async def edit_raw_put(collection: str, doc_id: str, request: Request) -> PlainTextResponse:
    if collection not in COLLECTIONS:
        raise HTTPException(404)
    form = await request.form()
    raw = form.get("json", "")
    try:
        data = json.loads(raw)
    except Exception as e:
        raise HTTPException(400, f"JSON non valido: {e}")
    if not isinstance(data, dict):
        raise HTTPException(400, "Il documento deve essere un oggetto JSON")
    try:
        oid = ObjectId(doc_id)
    except Exception:
        raise HTTPException(400, "invalid _id")
    # Imposta/forza l'_id corretto
    data["_id"] = oid
    db = await get_db()
    await db[collection].replace_one({"_id": oid}, data, upsert=False)
    return PlainTextResponse("Saved")
