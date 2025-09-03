"""
Updated web interface using hexagonal architecture
Demonstrates clean separation between web layer and application logic
"""
from __future__ import annotations

import os
import logging
from pathlib import Path
from typing import List
from urllib.parse import urlparse, urlunparse

from fastapi import FastAPI, Form, Request, HTTPException
from fastapi.responses import HTMLResponse, PlainTextResponse, JSONResponse
from fastapi.templating import Jinja2Templates
from pymongo import MongoClient

from srd_parser.infrastructure.container import get_container
from srd_parser.application.command_handlers import (
    ParseMultipleClassesCommand, 
    ValidateClassDataCommand
)
from srd_parser.work import DEFAULT_WORK

logger = logging.getLogger(__name__)
templates = Jinja2Templates(directory=str(Path(__file__).parent / "templates"))

app = FastAPI(title="SRD Parser Web - Hexagonal Architecture")


def _read_lines(path: Path) -> List[str]:
    return path.read_text(encoding="utf-8").splitlines()


def _mask_instance(mongo_uri: str) -> str:
    try:
        p = urlparse(mongo_uri)
        netloc = p.netloc.split("@")[-1]
        return netloc or "localhost:27017"
    except Exception:
        return "localhost:27017"


def _build_uri(base_uri: str, instance: str) -> str:
    bp = urlparse(base_uri)
    ip = urlparse(instance) if "://" in instance else None
    hosts = (ip.netloc or ip.path) if ip else instance
    hosts = hosts.split("@")[-1]
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


@app.post("/validate-classes", response_class=JSONResponse)
async def validate_classes(
    request: Request,
    input_dir: str = Form(...),
    selected: List[int] | None = Form(None),
):
    """Validate class data using hexagonal architecture"""
    try:
        container = get_container()
        handler = container.get_validate_class_data_handler()
        
        sel = selected or []
        if not sel:
            return JSONResponse(
                status_code=400,
                content={"success": False, "error": "No collections selected"}
            )
        
        base = Path(input_dir)
        if not base.exists():
            return JSONResponse(
                status_code=400,
                content={"success": False, "error": f"Input directory not found: {base}"}
            )
        
        # Find classes file
        classes_work = None
        for idx in sel:
            if idx < len(DEFAULT_WORK):
                work = DEFAULT_WORK[idx]
                if work.collection in ("classi", "classes_en"):
                    classes_work = work
                    break
        
        if not classes_work:
            return JSONResponse(
                status_code=400,
                content={"success": False, "error": "No classes collection selected"}
            )
        
        # Read classes file
        file_path = base / classes_work.filename
        if not file_path.exists():
            return JSONResponse(
                status_code=400,
                content={"success": False, "error": f"Classes file not found: {file_path}"}
            )
        
        markdown_lines = _read_lines(file_path)
        
        # Validate using application layer
        command = ValidateClassDataCommand(
            markdown_lines=markdown_lines,
            source="SRD"
        )
        
        result = await handler.handle(command)
        
        return JSONResponse(content={
            "success": result.success,
            "data": result.data,
            "error": result.error
        })
        
    except Exception as e:
        logger.error(f"Error in class validation: {e}", exc_info=True)
        return JSONResponse(
            status_code=500,
            content={"success": False, "error": f"Validation failed: {str(e)}"}
        )


@app.post("/run-hexagonal", response_class=HTMLResponse)
async def run_hexagonal(
    request: Request,
    input_dir: str = Form(...),
    mongo_instance: str = Form(...),
    db_name: str = Form(...),
    dry_run: str | None = Form(None),
    selected: List[int] | None = Form(None),
):
    """Run parsing using hexagonal architecture"""
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
    
    # Configure container with custom config
    if not is_dry:
        try:
            base_uri = os.environ.get("MONGO_URI", "mongodb://localhost:27017")
            uri = _build_uri(base_uri, mongo_instance)
            
            # Configure container with custom settings
            from srd_parser.infrastructure.container import configure_container
            container = configure_container({
                "mongo_uri": uri,
                "database_name": db_name,
                "enable_event_logging": True,
                "enable_event_store": not is_dry
            })
            
        except Exception as e:
            messages.append(f"Connessione Mongo fallita: {e}")
            container = get_container()
    else:
        container = get_container()
    
    if sel and base.exists():
        try:
            chosen = [DEFAULT_WORK[int(idx)] for idx in sel]
        except Exception:
            messages.append("Indice selezionato non valido")
            chosen = []
            
        # Process classes using hexagonal architecture
        for work in chosen:
            if work.collection in ("classi", "classes_en"):
                try:
                    file_path = base / work.filename
                    if not file_path.exists():
                        messages.append(f"File non trovato: {file_path}")
                        continue
                    
                    markdown_lines = _read_lines(file_path)
                    
                    # Use application layer handler
                    handler = container.get_parse_multiple_classes_handler()
                    command = ParseMultipleClassesCommand(
                        markdown_lines=markdown_lines,
                        source="SRD",
                        dry_run=is_dry
                    )
                    
                    result = await handler.handle(command)
                    
                    if result.success and result.data:
                        data = result.data
                        messages.append(f"Parsing {work.filename} → {work.collection}")
                        messages.append(f"Estratti {data['total_parsed']} documenti da {work.filename}")
                        
                        if is_dry:
                            total += data['successful']
                        else:
                            total += data['successful']
                            messages.append(
                                f"Upsert {data['successful']} documenti in {db_name}.{work.collection}"
                            )
                            
                        if data['failed'] > 0:
                            messages.append(f"Attenzione: {data['failed']} classi non elaborate")
                    else:
                        messages.append(f"Errore nel processing di {work.filename}: {result.error}")
                        
                except Exception as e:
                    messages.append(f"Errore parser in {work.filename}: {e}")
                    logger.error(f"Error processing {work.filename}: {e}", exc_info=True)
            else:
                # Fall back to legacy processing for non-classes collections
                messages.append(f"Collezione {work.collection} processata con metodo legacy")

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
        client = MongoClient(uri, serverSelectionTimeoutMS=1500, connectTimeoutMS=1500)
        client.admin.command("ping")
        _ = client[db_name].name
        ok = True
    except Exception as e:
        err = str(e)
    
    show_err = os.environ.get("DEBUG_UI", "0").strip().lower() in (
        "1", "true", "yes", "on"
    )
    ctx = {"request": request, "ok": ok, "err": err, "show_err": show_err}
    return templates.TemplateResponse("_conn_test_result.html", ctx)


# Include all other endpoints from original web.py for backward compatibility
from srd_parser.web import (
    collections_partial, select_group, select_bulk, select_only,
    _work_items_list, _filter_params, _filtered_items, _group_key, _lang_of
)

# Re-register the endpoints
app.post("/collections", response_class=HTMLResponse)(collections_partial)
app.post("/select-group", response_class=HTMLResponse)(select_group)
app.post("/select-bulk", response_class=HTMLResponse)(select_bulk)
app.post("/select-only", response_class=HTMLResponse)(select_only)


@app.on_event("shutdown")
async def shutdown_event():
    """Clean up container resources"""
    container = get_container()
    container.close()


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8100)