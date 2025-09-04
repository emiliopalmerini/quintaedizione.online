# 5e SRD Fast Viewer

Visualizzatore veloce per i contenuti SRD di D&D 5e, pensato per cercare e visualizzare rapidamente i dati in italiano.

**🚀 Migrato a Go per migliori performance e mantenibilità!**

## Tecnologie utilizzate

• **Backend**: Go + Gin + mongo-go-driver
• **Frontend**: Go templates + HTMX (progressive enhancement)  
• **UI**: Tailwind CSS (CDN) con componenti custom
• **Database**: MongoDB con autenticazione
• **Deployment**: Docker Compose per sviluppo locale
• **Parser**: Sistema modulare Go per ingest dati SRD italiani
• **Architettura**: Hexagonal Architecture con DDD
• **Performance**: Monitoring integrato, caching, metriche
• **Legacy**: Versioni Python disponibili per compatibilità

## Funzionalità principali
- Ricerca con filtri specifici per collezione (incantesimi, oggetti magici, mostri, ...)
- Vista "show" leggibile: corpo in Markdown, metadati compatti, classi con layout dedicato
- Breadcrumb interattivo con quicksearch in‑place
- Navigazione prev/next alfabetica coerente con la lista

## Avvio rapido
Requisiti: Docker, Docker Compose.

- (Opzionale) Inizializza variabili da `.env.example`
```bash
make env-init
```

- Avvia i servizi Go (raccomandato)
```bash
make up                    # MongoDB + Editor + Parser (Go)
```

- Oppure usa le versioni Python legacy
```bash
make up-python             # MongoDB + Editor + Parser (Python)
```

- (Opzionale) Ripristina seed se presente
```bash
make seed-restore FILE=backup_file.archive.gz
```

- Apri il visualizzatore: http://localhost:8000/
- Parser SRD: http://localhost:8100/

## Vantaggi della migrazione Go

### Performance
- **~3-5x** più veloce del equivalente Python
- **Memoria ridotta** grazie alla gestione nativa
- **Concorrenza** migliore con goroutine native
- **Caching integrato** con TTL per contenuti frequenti

### Monitoraggio
- Metriche di performance in tempo reale (`/health`)
- Tracking delle richieste e tempi di risposta
- Monitoraggio dell'uso memoria e goroutine
- Cache hit rate e statistiche

### Affidabilità  
- **Graceful shutdown** con timeout configurabile
- **Type safety** a compile time
- **Gestione errori** esplicita e robusta
- **Deploy semplificato** con binari statici

## Testing e qualità

```bash
# Test Go
make test-go               # Unit test
make test-integration      # Test di integrazione completi
make benchmark             # Performance benchmarks
make lint-go               # Code quality

# Test Python (legacy)
make test                  # Test integrazione Python
make lint                  # Lint Python
```
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
