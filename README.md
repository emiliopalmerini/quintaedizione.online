# 5e SRD Fast Viewer

Visualizzatore veloce per i contenuti SRD di D&D 5e, pensato per cercare e visualizzare rapidamente i dati in italiano e inglese.

## Tecnologie utilizzate

• **Backend**: FastAPI + Motor (MongoDB async)
• **Frontend**: Jinja2 + HTMX (progressive enhancement)
• **UI**: Tailwind CSS (CDN) con componenti custom
• **Database**: MongoDB con autenticazione
• **Deployment**: Docker Compose per sviluppo locale
• **Parser**: Sistema modulare per ingest dati SRD
• **Architettura**: Hexagonal Architecture con DDD

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

## Struttura del progetto

- `editor/`: Applicazione web FastAPI con template HTMX/Jinja2 (visualizzatore principale)
- `srd_parser/`: Parser modulare con web UI per ingest dati SRD in MongoDB
- `shared_domain/`: Entità di dominio condivise tra editor e parser
- `data/`: Dati sorgente SRD in italiano (`ita/`) e inglese (`eng/`)
- `seed/`: Dump database e script per backup/ripristino in sviluppo
- `docs/adrs/`: Architectural Decision Records (decisioni architetturali)
- File di configurazione root: `.env`, `docker-compose.yml`, `Makefile`

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

## Testing

### Test di integrazione
```bash
# Test integrazione base (richiede servizi attivi)
python test_basic_integration.py

# Test integrazione con curl
./test_curl_integration.sh

# Test domain model (unità)
python test_domain_model.py
```

### Test specifici per componenti
```bash
# Editor tests
cd editor && pytest

# Parser tests  
cd srd_parser && pytest
```

## Sicurezza
- Nessun dato sensibile: l'app gestisce solo contenuti SRD.
- Evita di loggare contenuti dei documenti o body delle richieste.
- La web‑app del parser (pagina di test connessione) mostra messaggi generici; per vedere dettagli di errore abilita esplicitamente `DEBUG_UI=1` nell'ambiente (off di default).

## Contribuire
- Usa Conventional Commits (feat, fix, docs, chore, build, refactor, perf, test)
- Mantieni patch piccole e focalizzate
- Aggiorna i README e gli ADR quando prendi decisioni architetturali
