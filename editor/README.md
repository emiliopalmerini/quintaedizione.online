# Viewer (FastAPI + HTMX)

Applicazione web per cercare e visualizzare i documenti SRD.

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
- Breadcrumb con quicksearch inline e link rapidi
- Navigazione prev/next alfabetica (name→term), coerente con la lista

## Struttura
- `main.py`: creazione app FastAPI
- `routers/pages.py`: route, viste, filtri
- `core/*`: DB, template env, utilità flatten/transform
- `templates/*`: template Jinja2/HTMX

## Estensioni/idee
- Dark mode CSS
- Advanced search filters
- Enhanced mobile UI
