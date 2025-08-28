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
  2. Esporta `MONGO_URI` e `DB_NAME` (default `dnd`). Se usi il Mongo del `docker-compose`:
     - `MONGO_URI="mongodb://admin:password@localhost:27017/?authSource=admin"`
     - altrimenti, senza auth locale: `mongodb://localhost:27017`
  3. `uvicorn main:app --reload --port 8000`

## Funzioni chiave
- Lista con filtri per collezione (spells, magic_items, monsters, ...)
- Vista show: corpo testuale in Markdown, metadati compatti, classi con layout dedicato
- Editor: filtri “solo traducibili/modificati”, textarea comode (select all, copy, expand), scorciatoie (Ctrl/Cmd+S, Ctrl+Alt+D), salvataggi parziali e globali
- Breadcrumb con quicksearch inline e link rapidi
- Navigazione prev/next alfabetica (name→term), coerente con la lista

## Struttura
- `main.py`: creazione app FastAPI
- `routers/pages.py`: route, viste, filtri
- `core/*`: DB, template env, utilità flatten/transform
- `templates/*`: template Jinja2/HTMX

## Estensioni/idee
- Anteprima Markdown live nell’editor per textarea lunghe
- Shortcut aggiuntive (text/higher_level)
- Dark mode CSS
