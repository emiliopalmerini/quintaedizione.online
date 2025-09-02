from __future__ import annotations

import os
from pathlib import Path
from typing import List
from urllib.parse import urlparse, urlunparse

from fastapi import FastAPI, Form, Request
from fastapi.responses import HTMLResponse, PlainTextResponse
from fastapi.templating import Jinja2Templates
from pymongo import MongoClient

from .adapters.persistence.mongo_repository import MongoRepository
from .application.ingest_runner import run_ingest
from .work import DEFAULT_WORK

templates = Jinja2Templates(directory=str(Path(__file__).parent / "templates"))

app = FastAPI(title="SRD Parser Web")


def _read_lines(path: Path) -> List[str]:
    return path.read_text(encoding="utf-8").splitlines()


def _mask_instance(mongo_uri: str) -> str:
    try:
        p = urlparse(mongo_uri)
        # netloc may be user:pass@host1,host2:port
        netloc = p.netloc.split("@")[-1]
        return netloc or "localhost:27017"
    except Exception:
        return "localhost:27017"


def _build_uri(base_uri: str, instance: str) -> str:
    # Preserve credentials and query/options from base_uri; replace hosts with instance
    bp = urlparse(base_uri)
    # normalize instance: remove scheme and userinfo if present
    ip = urlparse(instance) if "://" in instance else None
    hosts = (ip.netloc or ip.path) if ip else instance
    hosts = hosts.split("@")[-1]  # strip userinfo if any
    userinfo = ""
    if bp.username:
        userinfo = bp.username
        if bp.password:
            userinfo += f":{bp.password}"
    netloc = f"{userinfo + '@' if userinfo else ''}{hosts}"
    scheme = (ip.scheme if ip and ip.scheme else bp.scheme) or "mongodb"
    return urlunparse((scheme, netloc, bp.path, bp.params, bp.query, bp.fragment))


def _default_env() -> dict:
    base_uri = os.environ.get("MONGO_URI", "mongodb://localhost:27017")
    return {
        "input_dir": os.environ.get("INPUT_DIR", "data"),
        "mongo_instance": _mask_instance(base_uri),
        "db_name": os.environ.get("DB_NAME", "dnd"),
        "dry_run": True,
    }


@app.get("/", response_class=HTMLResponse)
async def index(request: Request):
    env = _default_env()
    work_items = [
        {"idx": i, "collection": w.collection, "filename": w.filename}
        for i, w in enumerate(DEFAULT_WORK)
    ]
    return templates.TemplateResponse(
        "parser_form.html",
        {
            "request": request,
            "env": env,
            "work_items": work_items,
            "messages": [],
            "selected": [],
            "flt_it": True,
            "flt_en": True,
            "flt_docs": True,
            "group": False,
            "q": "",
        },
    )


@app.get("/healthz", response_class=PlainTextResponse)
async def healthz():
    return PlainTextResponse("ok")


@app.post("/run", response_class=HTMLResponse)
async def run(
    request: Request,
    input_dir: str = Form(...),
    mongo_instance: str = Form(...),
    db_name: str = Form(...),
    dry_run: str | None = Form(None),
    selected: List[int] | None = Form(None),
):
    is_dry = dry_run is not None
    sel = selected or []
    work_items = [
        {"idx": i, "collection": w.collection, "filename": w.filename}
        for i, w in enumerate(DEFAULT_WORK)
    ]

    messages: List[str] = []
    base = Path(input_dir)
    if not sel:
        messages.append("Nessuna collezione selezionata.")
    if not base.exists():
        messages.append(f"Cartella input non trovata: {base}")

    total = 0
    repo = None
    if not is_dry and sel and base.exists():
        try:
            base_uri = os.environ.get("MONGO_URI", "mongodb://localhost:27017")
            uri = _build_uri(base_uri, mongo_instance)
            client = MongoClient(uri)
            repo = MongoRepository(client[db_name])
        except Exception as e:
            messages.append(f"Connessione Mongo fallita: {e}")

    if sel and base.exists():
        try:
            chosen = [DEFAULT_WORK[int(idx)] for idx in sel]
        except Exception:
            messages.append("Indice selezionato non valido")
            chosen = []
        if chosen:
            results = run_ingest(base, chosen, None if is_dry else repo, dry_run=is_dry)
            for r in results:
                if r.error:
                    messages.append(f"Errore parser in {r.filename}: {r.error}")
                    continue
                messages.append(f"Parsing {r.filename} → {r.collection}")
                messages.append(f"Estratti {r.parsed} documenti da {r.filename}")
                if is_dry:
                    total += r.parsed
                else:
                    total += r.written
                    messages.append(
                        f"Upsert {r.written} documenti in {db_name}.{r.collection}"
                    )

    if is_dry:
        messages.append(f"Dry-run completato. Totale analizzati: {total}")
    else:
        messages.append(f"Fatto. Totale upsert: {total}")

    env = {
        "input_dir": input_dir,
        "mongo_instance": mongo_instance,
        "db_name": db_name,
        "dry_run": is_dry,
    }

    return templates.TemplateResponse(
        "parser_form.html",
        {
            "request": request,
            "env": env,
            "work_items": work_items,
            "messages": messages,
            "selected": sel,
            "flt_it": True,
            "flt_en": True,
            "flt_docs": True,
            "group": False,
            "q": "",
        },
    )


@app.post("/test-conn", response_class=HTMLResponse)
async def test_conn(
    request: Request,
    mongo_instance: str = Form(...),
    db_name: str = Form(...),
):
    base_uri = os.environ.get("MONGO_URI", "mongodb://localhost:27017")
    uri = _build_uri(base_uri, mongo_instance)
    ok = False
    err: str | None = None
    try:
        # Short timeouts for snappy UX
        client = MongoClient(uri, serverSelectionTimeoutMS=1500, connectTimeoutMS=1500)
        client.admin.command("ping")
        _ = client[db_name].name  # touch db ref
        ok = True
    except Exception as e:
        err = str(e)
    show_err = os.environ.get("DEBUG_UI", "0").strip().lower() in (
        "1",
        "true",
        "yes",
        "on",
    )
    ctx = {"request": request, "ok": ok, "err": err, "show_err": show_err}
    return templates.TemplateResponse("_conn_test_result.html", ctx)


def _work_items_list():
    return [
        {"idx": i, "collection": w.collection, "filename": w.filename}
        for i, w in enumerate(DEFAULT_WORK)
    ]


def _filter_params(
    input_q: str | None,
    flt_it: str | None,
    flt_en: str | None,
    flt_docs: str | None,
    group: str | None,
):
    return {
        "q": (input_q or "").strip().lower(),
        "flt_it": flt_it is not None,
        "flt_en": flt_en is not None,
        "flt_docs": flt_docs is not None,
        "group": group is not None,
    }


@app.post("/collections", response_class=HTMLResponse)
async def collections_partial(
    request: Request,
    q: str | None = Form(default=""),
    flt_it: str | None = Form(default=None),
    flt_en: str | None = Form(default=None),
    flt_docs: str | None = Form(default=None),
    group: str | None = Form(default=None),
    selected: List[int] | None = Form(default=None),
):
    wi = _work_items_list()
    params = _filter_params(q, flt_it, flt_en, flt_docs, group)
    return templates.TemplateResponse(
        "_collections.html",
        {"request": request, "work_items": wi, "selected": selected or [], **params},
    )


def _lang_of(coll: str) -> str:
    lc = coll.lower()
    if lc.endswith("_en"):
        return "en"
    if "documenti" in lc:
        return "doc"
    return "it"


def _group_key(coll: str) -> str:
    c = coll
    if c.startswith("documenti"):
        return "docs"
    if c in ("classi", "classi_en"):
        return "classes"
    if c.startswith("backgrounds"):
        return "backgrounds"
    if c in ("incantesimi", "spells_en"):
        return "spells"
    if c in ("armi", "weapons_en"):
        return "weapons"
    if c in ("armature", "armor_en"):
        return "armor"
    if c in ("strumenti", "tools_en"):
        return "tools"
    if c in ("equipaggiamento", "adventuring_gear_en"):
        return "gear"
    if c in ("servizi", "services_en"):
        return "services"
    if c in ("oggetti_magici", "magic_items_en"):
        return "magic_items"
    if c in ("mostri", "monsters_en"):
        return "monsters"
    if c in ("animali", "animals_en"):
        return "animals"
    return "other"


def _filtered_items(q: str, flt_it: bool, flt_en: bool, flt_docs: bool) -> List[dict]:
    items = _work_items_list()
    out: List[dict] = []
    for w in items:
        lang = _lang_of(w["collection"])
        if lang == "it" and not flt_it:
            continue
        if lang == "en" and not flt_en:
            continue
        if lang == "doc" and not flt_docs:
            continue
        if q and (q not in w["collection"].lower() and q not in w["filename"].lower()):
            continue
        out.append(w)
    return out


@app.post("/select-group", response_class=HTMLResponse)
async def select_group(
    request: Request,
    group: str = Form(...),
    mode: str = Form(...),
    q: str | None = Form(default=""),
    flt_it: str | None = Form(default=None),
    flt_en: str | None = Form(default=None),
    flt_docs: str | None = Form(default=None),
    group_view: str | None = Form(default=None),
    selected: List[int] | None = Form(default=None),
):
    params = _filter_params(q, flt_it, flt_en, flt_docs, group_view)
    visible = _filtered_items(
        params["q"], params["flt_it"], params["flt_en"], params["flt_docs"]
    )
    # Build index set for the target group
    grp_idxs = {w["idx"] for w in visible if _group_key(w["collection"]) == group}
    cur = set(selected or [])
    if mode == "all":
        cur.update(grp_idxs)
    elif mode == "none":
        cur.difference_update(grp_idxs)
    elif mode == "invert":
        for i in list(grp_idxs):
            if i in cur:
                cur.remove(i)
            else:
                cur.add(i)
    elif mode == "only":
        cur = set(grp_idxs)
    wi = _work_items_list()
    return templates.TemplateResponse(
        "_collections.html",
        {"request": request, "work_items": wi, "selected": sorted(cur), **params},
    )


@app.post("/select-bulk", response_class=HTMLResponse)
async def select_bulk(
    request: Request,
    mode: str = Form(...),
    q: str | None = Form(default=""),
    flt_it: str | None = Form(default=None),
    flt_en: str | None = Form(default=None),
    flt_docs: str | None = Form(default=None),
    group: str | None = Form(default=None),
    selected: List[int] | None = Form(default=None),
):
    params = _filter_params(q, flt_it, flt_en, flt_docs, group)
    visible = _filtered_items(
        params["q"], params["flt_it"], params["flt_en"], params["flt_docs"]
    )
    vis_idxs = {w["idx"] for w in visible}
    cur = set(selected or [])
    if mode == "it-structured":
        # structured IT: exclude EN and docs
        cur = {
            w["idx"]
            for w in visible
            if (
                not w["collection"].endswith("_en")
                and "documenti" not in w["collection"]
            )
        }
    elif mode == "en-structured":
        cur = {w["idx"] for w in visible if w["collection"].endswith("_en")}
    elif mode == "docs-only":
        cur = {w["idx"] for w in visible if "documenti" in w["collection"]}
    elif mode == "all_visible":
        cur.update(vis_idxs)
    elif mode == "none_visible":
        cur.difference_update(vis_idxs)
    elif mode == "invert_visible":
        for i in list(vis_idxs):
            if i in cur:
                cur.remove(i)
            else:
                cur.add(i)
    wi = _work_items_list()
    return templates.TemplateResponse(
        "_collections.html",
        {"request": request, "work_items": wi, "selected": sorted(cur), **params},
    )


@app.post("/select-only", response_class=HTMLResponse)
async def select_only(
    request: Request,
    idx: int = Form(...),
    q: str | None = Form(default=""),
    flt_it: str | None = Form(default=None),
    flt_en: str | None = Form(default=None),
    flt_docs: str | None = Form(default=None),
    group: str | None = Form(default=None),
    selected: List[int] | None = Form(default=None),
):
    params = _filter_params(q, flt_it, flt_en, flt_docs, group)
    cur = {idx}
    wi = _work_items_list()
    return templates.TemplateResponse(
        "_collections.html",
        {"request": request, "work_items": wi, "selected": sorted(cur), **params},
    )
