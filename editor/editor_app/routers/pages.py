# app/routers/pages.py
from __future__ import annotations

import math
from typing import Optional, Any, Dict

from bson import ObjectId
from fastapi import APIRouter, HTTPException, Request
from fastapi.responses import HTMLResponse, PlainTextResponse

from editor_app.core.config import COLLECTIONS
from editor_app.core.db import get_db
from editor_app.core.flatten import flatten_for_form
from editor_app.core.templates import env

router = APIRouter()


def build_filter(q: str) -> Dict[str, Any]:
    if not q:
        return {}
    r = {"$regex": q, "$options": "i"}
    return {"$or": [{"name": r}, {"term": r}, {"title": r}, {"description": r}]}


@router.get("/", response_class=HTMLResponse)
async def index() -> HTMLResponse:
    tpl = env.get_template("index.html")
    first = COLLECTIONS[0] if COLLECTIONS else ""
    return HTMLResponse(tpl.render(collections=COLLECTIONS, first=first))


@router.get("/list/{collection}", response_class=HTMLResponse)
async def list_page(collection: str, q: str = "", page: int = 1, page_size: int = 20) -> HTMLResponse:
    if collection not in COLLECTIONS:
        raise HTTPException(404)
    tpl = env.get_template("list.html")
    return HTMLResponse(tpl.render(collection=collection, q=q, page=page, page_size=page_size))


@router.get("/view/{collection}", response_class=HTMLResponse)
async def view_rows(collection: str, q: str = "", page: int = 1, page_size: int = 20) -> HTMLResponse:
    if collection not in COLLECTIONS:
        raise HTTPException(404)
    db = await get_db()
    col = db[collection]
    filt = build_filter(q)
    total = await col.count_documents(filt)
    pages = max(1, math.ceil(total / page_size))
    page = max(1, min(page, pages))
    cursor = col.find(filt).sort("name", 1).skip((page - 1) * page_size).limit(page_size)
    items = []
    async for doc in cursor:
        doc["_id"] = str(doc["_id"])
        items.append(doc)
    tpl = env.get_template("_rows.html")
    return HTMLResponse(
        tpl.render(
            collection=collection,
            items=items,
            page=page,
            pages=pages,
            total=total,
            page_size=page_size,
            q=q,
        )
    )


@router.get("/doc/{collection}/{doc_id}", response_class=HTMLResponse)
async def doc_form(collection: str, doc_id: str, return_to: Optional[str] = None) -> HTMLResponse:
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
        raise HTTPException(404)
    fields = flatten_for_form(doc)
    tpl = env.get_template("_doc_form.html")
    return HTMLResponse(
        tpl.render(
            collection=collection,
            doc_id=str(doc["_id"]),
            fields=fields,
            return_to=return_to or f"/list/{collection}",
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

