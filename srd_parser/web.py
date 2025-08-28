from __future__ import annotations

import os
from pathlib import Path
from typing import List
from urllib.parse import urlparse, urlunparse

from fastapi import FastAPI, Form, Request
from fastapi.responses import HTMLResponse, PlainTextResponse
from fastapi.templating import Jinja2Templates
from pymongo import MongoClient

from .work import DEFAULT_WORK
from .adapters.persistence.mongo_repository import MongoRepository
from .application.ingest_runner import run_ingest


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
    work_items = [{"idx": i, "collection": w.collection, "filename": w.filename} for i, w in enumerate(DEFAULT_WORK)]
    return templates.TemplateResponse(
        "parser_form.html",
        {
            "request": request,
            "env": env,
            "work_items": work_items,
            "messages": [],
            "selected": [],
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
    work_items = [{"idx": i, "collection": w.collection, "filename": w.filename} for i, w in enumerate(DEFAULT_WORK)]

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
                    messages.append(f"Upsert {r.written} documenti in {db_name}.{r.collection}")

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
    show_err = os.environ.get("DEBUG_UI", "0").strip().lower() in ("1", "true", "yes", "on")
    ctx = {"request": request, "ok": ok, "err": err, "show_err": show_err}
    return templates.TemplateResponse("_conn_test_result.html", ctx)
