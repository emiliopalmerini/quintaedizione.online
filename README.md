# 5e SRD Fast Editor

Editor e visualizzatore veloce per i contenuti SRD di D&D 5e, pensato per aiutare traduttori ed editor a cercare, visualizzare e tradurre rapidamente i dati.

• Backend FastAPI + Motor (MongoDB)
• Templating Jinja2 + HTMX (interazioni semplici e progressive)
• UI Tailwind (CDN) con componenti custom leggeri
• DB MongoDB con seed opzionale versionabile
• Docker Compose per ambiente locale

## Funzionalità principali
- Ricerca con filtri specifici per collezione (incantesimi, oggetti magici, mostri, ...)
- Vista “show” leggibile: corpo in Markdown, metadati compatti, classi con layout dedicato
- Editor ottimizzato per traduzione: textarea comode, scorciatoie, filtri “solo traducibili/modificati”, salvataggi rapidi
- Breadcrumb interattivo con quicksearch in‑place
- Navigazione prev/next alfabetica coerente con la lista

## Avvio rapido
Requisiti: Docker, Docker Compose.

- Avvia Mongo + Editor
```
make up
```
- (Opzionale) Ripristina seed se presente
```
make seed-restore
```
- Apri l’editor: http://localhost:8000/

TUI per il parser SRD (on‑demand):
```
# Avvia la TUI in un container
make tui
# Oppure avvia Mongo e poi la TUI
make tui-up
```

Vedi anche `Makefile` per altri comandi utili.

## Struttura del repo
- `editor/`: applicazione FastAPI + template HTMX/Jinja2
- `srd_parser/`: parser/ingest dei dati SRD in MongoDB
- `seed/`: dump e script per ripristino del DB in dev
- `docs/adr/`: Architectural Decision Records

## Documentazione
- Editor: `editor/README.md`
- Parser: `srd_parser/README.md`
- Seed: `seed/README.md`
- Agenti/LLM: `AGENTS.md`, `LLMS.md`
- ADR: cartella `docs/adr/`

## Contribuire
- Usa Conventional Commits (feat, fix, docs, chore, build, refactor, perf, test)
- Mantieni patch piccole e focalizzate
- Aggiorna i README e gli ADR quando prendi decisioni architetturali
