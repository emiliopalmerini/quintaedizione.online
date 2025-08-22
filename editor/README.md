# Editor (FastAPI + HTMX)

Applicazione web per cercare, visualizzare e modificare i documenti SRD.

## Stack
- FastAPI, Jinja2, Motor (MongoDB)
- HTMX per interazioni progressive (ricarichi parziali)
- Tailwind via CDN + CSS custom leggero

## Avvio in locale
- Via Docker Compose: `make up` e apri http://localhost:8000
- Senza Docker:
  1. `cd editor && pip install -r requirements.txt`
  2. Esporta `MONGO_URI` (es. `mongodb://localhost:27017`) e `DB_NAME` (default `dnd`)
  3. `uvicorn main:app --reload --port 8000`

## Funzioni chiave
- Lista con filtri per collezione (spells, magic_items, monsters, ...)
- Vista show: corpo testuale in Markdown, metadati compatti, classi con layout dedicato
- Editor: filtri “solo traducibili/modificati”, textarea comode (select all, copy, expand), scorciatoie (Ctrl/Cmd+S, Ctrl+Alt+D), salvataggi parziali e globali
- Breadcrumb con quicksearch inline e link rapidi
- Navigazione prev/next alfabetica (name→term), coerente con la lista

## Struttura
- `editor_app/main.py`: creazione app FastAPI
- `editor_app/routers/pages.py`: route, viste, filtri
- `editor_app/core/*`: DB, template env, utilità flatten/transform
- `editor_app/templates/*`: template Jinja2/HTMX

## Estensioni/idee
- Anteprima Markdown live nell’editor per textarea lunghe
- Shortcut aggiuntive (text/higher_level)
- Dark mode CSS

