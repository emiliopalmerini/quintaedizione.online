# 5e SRD Fast Viewer

Visualizzatore veloce per i contenuti SRD di D&D 5e, pensato per cercare e visualizzare rapidamente i dati.

• Backend FastAPI + Motor (MongoDB)
• Templating Jinja2 + HTMX (interazioni semplici e progressive)
• UI Tailwind (CDN) con componenti custom leggeri
• DB MongoDB con seed opzionale versionabile
• Docker Compose per ambiente locale

## Funzionalità principali
- Ricerca con filtri specifici per collezione (incantesimi, oggetti magici, mostri, ...)
- Vista "show" leggibile: corpo in Markdown, metadati compatti, classi con layout dedicato
- Breadcrumb interattivo con quicksearch in‑place
- Navigazione prev/next alfabetica coerente con la lista

## Avvio rapido
Requisiti: Docker, Docker Compose.

- (Opzionale) Inizializza variabili da `.env.example`
```
make env-init
```
- Avvia Mongo + Editor
```
make up
```
- (Opzionale) Ripristina seed se presente
```
make seed-restore
```
- Apri il visualizzatore: http://localhost:8000/

Parser SRD via Web UI:
```
# Avvia il servizio del parser web (se non usi make up)
docker compose up -d srd-parser
# Apri la web app del parser
open http://localhost:8100
```

Vedi anche `Makefile` per altri comandi utili.

Comandi utili:
```
make build           # build viewer + srd-parser
make build-editor    # build solo viewer
make build-parser    # build solo srd-parser
make lint            # esegue ruff/pyflakes se presenti
make format          # esegue black se presente
```

## Struttura del repo
- `editor/`: applicazione FastAPI + template HTMX/Jinja2 (visualizzatore)
- `srd_parser/`: parser/ingest dei dati SRD in MongoDB
- `seed/`: dump e script per ripristino del DB in dev
- `docs/adr/`: Architectural Decision Records

## Documentazione
- Visualizzatore: `editor/README.md`
- Parser: `srd_parser/README.md`
- Seed: `seed/README.md`
- Agenti/LLM: `AGENTS.md`, `LLMS.md`
 - ADR: cartella `docs/adrs/`

## Note su MongoDB e autenticazione
- In `docker-compose.yml` Mongo viene avviato con utente root (`admin/password`).
- Le app usano `MONGO_URI=mongodb://admin:password@mongo:27017/?authSource=admin` di default.
- In sviluppo locale (senza Docker) imposta `MONGO_URI` coerente, ad esempio `mongodb://admin:password@localhost:27017/?authSource=admin`.

## Sicurezza
- Nessun dato sensibile: l'app gestisce solo contenuti SRD.
- Evita di loggare contenuti dei documenti o body delle richieste.
- La web‑app del parser (pagina di test connessione) mostra messaggi generici; per vedere dettagli di errore abilita esplicitamente `DEBUG_UI=1` nell'ambiente (off di default).

## Contribuire
- Usa Conventional Commits (feat, fix, docs, chore, build, refactor, perf, test)
- Mantieni patch piccole e focalizzate
- Aggiorna i README e gli ADR quando prendi decisioni architetturali
