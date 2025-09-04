# Contributing to 5e SRD Fast Viewer

Grazie per il tuo interesse nel contribuire al progetto! Questa guida ti aiuterà a iniziare velocemente.

## Setup Sviluppo (< 5 minuti)

### Prerequisiti
- Docker e Docker Compose installati
- Git per il clone del repository

### Setup rapido
```bash
# Clone del repository
git clone <repository-url>
cd due-draghi-5e-srd

# (Opzionale) Inizializza .env
make env-init

# Avvia tutti i servizi
make up

# (Opzionale) Ripristina dati di esempio
make seed-restore

# Verifica che tutto funzioni
open http://localhost:8000  # Editor
open http://localhost:8100  # Parser
```

## Architettura del Progetto

### Componenti Principali
- **Editor** (`/editor`): Applicazione web FastAPI + HTMX per visualizzare contenuti SRD
- **Parser** (`/srd_parser`): Sistema modulare per parsing e ingest dati SRD con architettura hexagonal/DDD
- **Shared Domain** (`/shared_domain`): Entità di dominio condivise
- **Data** (`/data`): Sorgenti SRD in italiano e inglese

### Stack Tecnologico
- **Backend**: FastAPI, Motor (MongoDB async), PyMongo (sync)
- **Frontend**: Jinja2, HTMX, Tailwind CSS
- **Database**: MongoDB con autenticazione
- **Testing**: pytest, test di integrazione
- **Tools**: Docker Compose, Makefile

## Workflow di Sviluppo

### 1. Branch Strategy
```bash
# Crea un branch per la tua feature/fix
git checkout -b feat/nuova-feature
# oppure
git checkout -b fix/bug-fix
```

### 2. Sviluppo
```bash
# Durante lo sviluppo, ricostruisci solo quando necessario
make build-editor    # se modifichi l'editor
make build-parser    # se modifichi il parser

# Controlla i log
make logs

# Esegui i test
python test_basic_integration.py  # test di integrazione
cd editor && pytest               # test editor
cd srd_parser && pytest           # test parser
```

### 3. Code Quality
```bash
# Formattazione e linting (opzionale, richiede ruff/black installati)
make format
make lint
```

### 4. Test
Prima di fare commit, assicurati che i test passino:
```bash
# Test di integrazione (richiede servizi attivi)
python test_basic_integration.py
./test_curl_integration.sh

# Test unitari
python test_domain_model.py
```

### 5. Commit
Usa [Conventional Commits](https://www.conventionalcommits.org/):
```bash
git add .
git commit -m "feat: aggiungi filtro per scuola di magia"
# oppure
git commit -m "fix: correggi parsing spell slot per warlock"
git commit -m "docs: aggiorna README con nuovi comandi"
```

## Linee Guida per Contributi

### Cosa Cerchiamo
- **Bug fixes**: Correzioni di errori, miglioramenti di performance
- **Nuove feature**: Filtri avanzati, miglioramenti UI/UX, nuovi parser
- **Documentazione**: Miglioramenti a README, ADR, commenti nel codice
- **Test**: Nuovi test, miglioramento coverage
- **Refactoring**: Miglioramenti architetturali mantenendo retrocompatibilità

### Code Style
- **Python**: Segui PEP 8, usa type hints dove possibile
- **HTML/CSS**: Segui pattern HTMX esistenti, usa Tailwind per styling
- **JavaScript**: Minimale, preferisci HTMX quando possibile
- **Database**: Usa Motor (async) per editor, PyMongo (sync) per parser

### Convenzioni File
- I template Jinja2 usano italiano per i testi UI
- Il codice Python è in inglese (nomi variabili, commenti, docstring)
- I file markdown usano italiano per documentazione utente
- ADR e documentation tecnica in italiano

### Sicurezza
- Non committare mai credenziali o secrets
- Non loggare contenuti sensibili (anche se attualmente gestiamo solo SRD)
- Valida sempre input utente
- Mantieni messaggi di errore generici nel UI

## Tipi di Contributi

### Parser per Nuovi Tipi di Contenuto
```bash
# Struttura per nuovo parser
srd_parser/parsers/mio_nuovo_parser.py

# Aggiungi la configurazione in
srd_parser/work.py

# Test del parser
srd_parser/tests/test_mio_parser.py
```

### Miglioramenti UI Editor
```bash
# Template
editor/templates/

# Styling
editor/static/css/

# Logica frontend
editor/routers/pages.py
```

### Nuove Funzionalità Database
Quando aggiungi nuove collection o modifichi lo schema:
1. Aggiorna `docs/adrs/0001-data-model.md`
2. Considera l'impatto su esistente
3. Aggiungi migration script se necessario

## Test e Debugging

### Test Locali
```bash
# Test quick (domain model)
python test_domain_model.py

# Test full (richiede servizi)
make up
python test_basic_integration.py

# Test specifici
cd editor && pytest tests/test_routes.py -v
cd srd_parser && pytest tests/test_parsers.py -v
```

### Debugging
```bash
# Log dei servizi
make logs

# Log singolo servizio
docker compose logs editor
docker compose logs srd-parser
docker compose logs mongo

# Accesso diretto ai container
docker compose exec editor bash
docker compose exec srd-parser bash
```

### Database Debug
```bash
# Backup/restore per test
make seed-dump
make seed-restore

# Accesso MongoDB
docker compose exec mongo mongosh -u admin -p password --authenticationDatabase admin
```

## Architectural Decision Records (ADR)

Per modifiche architetturali significative, crea un ADR:
```bash
# Copia template
cp docs/adrs/0000-template.md docs/adrs/000X-mia-decisione.md

# Documenta: Context, Decision, Consequences
# Commit ADR insieme al codice
```

## Getting Help

### Risorse
- **README**: Panoramica generale e quick start
- **CLAUDE.md**: Guida dettagliata per AI assistants
- **docs/adrs/**: Decisioni architetturali
- **Component READMEs**: `editor/README.md`, `srd_parser/README.md`

### Common Issues

**Servizi non partono**
```bash
make down && make up
# oppure
docker compose down -v && make up
```

**Database vuoto**
```bash
make seed-restore
```

**Parsing fallisce**
```bash
# Controlla formato dati sorgente in data/
# Verifica configurazione in srd_parser/work.py
# Test in modalità dry-run nel parser UI
```

**Test falliscono**
```bash
# Assicurati servizi attivi
make up

# Verifica connessioni
curl http://localhost:8000/health
curl http://localhost:8100/
```

## Rilascio

I rilasci seguono [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes all'API o architettura
- **MINOR**: Nuove feature backward-compatible
- **PATCH**: Bug fixes

## Licenza

Tutti i contributi sono sotto la stessa licenza del progetto. I contenuti SRD sono sotto Creative Commons Attribution 4.0 International License (CC-BY-4.0).

---

Grazie per aver contribuito al progetto! 🎲✨