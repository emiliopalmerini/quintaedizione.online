# Quinta Edizione.online - Sistema di Gestione SRD D&D 5e

Un sistema web completo per la gestione e visualizzazione dei contenuti del System Reference Document (SRD) di Dungeons & Dragons 5a Edizione in italiano. Sviluppato con Go, Amore e Lacrime con un'architettura pulita e tecnologie web moderne.

## 🚀 Funzionalità

- **Supporto Solo Lingua Italiana**: Contenuti SRD completi in italiano
- **Viewer Web**: Interfaccia user-friendly per visualizzare i contenuti D&D
- **Parser CLI**: Strumento da linea di comando per elaborare file markdown in documenti di database strutturati
- **Ricerca e Navigazione**: Scoperta e navigazione rapida dei contenuti
- **Rendering Basato su Template**: Interfaccia web pulita e responsive con HTMX + Templ
- **Integrazione Docker**: Deploy containerizzato con Docker Compose

## 📋 Requisiti

- Docker e Docker Compose
- Go 1.24+ (per sviluppo locale)
- Make (per i comandi di build)

## 🔧 Avvio Rapido

1. **Clonare il repository**:
   ```bash
   git clone https://github.com/emiliopalmerini/due-draghi-5e-srd.git
   cd due-draghi-5e-srd
   ```

2. **Inizializzare l'ambiente**:
   ```bash
   make env-init
   ```

3. **Avviare i servizi**:
   ```bash
   make up
   ```

4. **Accedere all'applicazione**:
   - Viewer: http://localhost:8000

## 🏗️ Architettura

### Pattern Clean Architecture
```
├── cmd/                    # Punti di ingresso dell'applicazione
│   ├── viewer/            # Servizio viewer web
│   └── cli-parser/        # Tool CLI per parsing contenuti
├── internal/
│   ├── domain/            # Logica di business ed entità core
│   ├── application/       # Casi d'uso, servizi, handler
│   │   ├── handlers/      # Handler richieste HTTP
│   │   ├── parsers/       # Strategie di parsing contenuti
│   │   │   ├── strategy.go     # Interfaccia ParsingStrategy
│   │   │   ├── registry.go     # Gestione registry parser
│   │   │   ├── content_types.go # Definizioni tipi contenuto
│   │   │   └── *_strategy.go   # Implementazioni parser concrete
│   │   └── services/      # Servizi di business
│   ├── adapters/          # Interfacce esterne
│   │   ├── repositories/  # Interfacce e implementazioni repository
│   │   │   ├── factory.go # Factory repository per dependency injection
│   │   │   └── mongodb/   # Implementazioni repository specifiche MongoDB
│   │   └── web/           # Handler web e routing
│   ├── infrastructure/    # Configurazione e setup
│   └── shared/            # Modelli e utility comuni
├── pkg/                   # Pacchetti riutilizzabili
│   ├── mongodb/           # Client e configurazione MongoDB
│   └── templates/         # Utility template
├── data/                  # File contenuti SRD
│   └── ita/              # Contenuti SRD italiani
│       ├── lists/        # **Sorgente primaria parsing**: Liste entità pulite
│       │   ├── animali.md      # Definizioni animali
│       │   ├── armi.md         # Definizioni armi
│       │   ├── armature.md     # Definizioni armature
│       │   ├── backgrounds.md  # Definizioni background
│       │   ├── classi.md       # Definizioni classi
│       │   ├── equipaggiamenti.md # Definizioni equipaggiamenti
│       │   ├── incantesimi.md  # Definizioni incantesimi
│       │   ├── mostri.md       # Definizioni mostri
│       │   ├── oggetti_magici.md # Definizioni oggetti magici
│       │   ├── regole.md       # Definizioni regole
│       │   └── talenti.md      # Definizioni talenti
│       ├── docs/         # Documentazione SRD originale (backup)
│       └── DIZIONARIO_CAMPI_ITALIANI.md # Terminologia campi italiani
└── web/                   # Asset web
    ├── static/           # CSS, JS, immagini
    └── templates/        # Template HTML
```

### Architettura Parser

Il servizio parser usa **Pattern Strategy + Registry** per elaborazione flessibile contenuti:

- **Pattern Strategy**: Ogni tipo di contenuto (incantesimi, mostri, classi) ha la propria strategia di parsing
- **Pattern Registry**: Gestione centralizzata e thread-safe di tutti i parser disponibili  
- **Oggetti Domain**: I parser restituiscono entità domain fortemente tipizzate, non mappe generiche
- **Clean Architecture**: Separazione chiara tra entità domain e logica di parsing

**Componenti Chiave:**
- Interfaccia `ParsingStrategy` nel layer applicativo (non domain)
- `ParserRegistry` per registrazione e recupero dinamico parser
- Strategie concrete che restituiscono oggetti domain (es. `SpellsStrategy` → `domain.Incantesimo`)

### Architettura Repository

Il sistema implementa repository specifici per entità usando il pattern Repository:

- **Repository Factory**: `internal/adapters/repositories/factory.go` fornisce dependency injection
- **Repository MongoDB**: Ogni entità domain ha la propria implementazione repository MongoDB
- **Base Repository**: Operazioni CRUD comuni in `base_mongo_repository.go`
- **Type Safety**: Operazioni specifiche per dominio per ogni tipo entità

Questo pattern assicura separazione pulita tra logica domain e accesso dati, rendendo il sistema facilmente testabile e manutenibile.

### Servizi

#### 1. Servizio Viewer (Porta 8000)
- Interfaccia web per visualizzare contenuti D&D
- Rendering basato su template con tecnologie web moderne
- Navigazione e capacità di ricerca user-friendly

#### 2. Parser CLI
- Tool da linea di comando per elaborare file markdown dalla directory `data/`
- Converte contenuti in documenti MongoDB strutturati
- Supporta tipi contenuto multipli: classi, background, incantesimi, oggetti, ecc.

### Struttura Database

**Collezioni MongoDB:**

```
[
  'animali',
  'armi',
  'armature', 
  'backgrounds',
  'cavalcature_e_veicoli',
  'classi',
  'documenti',
  'equipaggiamento',
  'incantesimi',
  'mostri',
  'oggetti_magici',
  'regole',
  'servizi',
  'specie',
  'strumenti',
  'talenti'
]
```

**Schema Documento:**
- **BaseEntity**: Campi comuni (ID, timestamp, versione, sorgente)
- **MarkdownContent**: Contenuto multi-formato (markdown grezzo, HTML, testo semplice)
- **SearchableContent**: Metadati di ricerca ottimizzati

## 🛠️ Sviluppo

### Comandi Disponibili

**Servizi Docker:**
```bash
make up          # Avvia MongoDB + Quinta Edizione.online Viewer
make down        # Ferma tutti i servizi
make logs        # Visualizza log servizi
make build       # Costruisce immagini Go
```

**Sviluppo Go:**
```bash
make lint        # Esegue linting (go vet + golangci-lint)
make test        # Esegue unit test
make test-integration  # Esegue integration test
make benchmark   # Esegue performance benchmark
```

**Gestione Database:**
```bash
make seed-dump                    # Crea backup con timestamp
make seed-restore FILE=backup.gz  # Ripristina da backup
make mongo-sh                     # Accede a shell MongoDB
```

**Accesso Container:**
```bash
make viewer-sh   # Accede al container viewer
make mongo-sh    # Accede al container MongoDB
```

### Setup Sviluppo Locale

1. **Installare dipendenze**:
   ```bash
   go mod download
   ```

2. **Configurare ambiente**:
   ```bash
   cp .env.example .env
   # Modifica .env con la tua configurazione
   ```

3. **Eseguire servizi localmente**:
   ```bash
   # Avvia MongoDB
   docker compose up -d mongo
   
   # Esegui viewer localmente
   cd cmd/viewer && go run main.go
   
   # Esegui parser CLI localmente
   cd cmd/cli-parser && go run main.go
   ```

## 📄 Standard Formato Dati

Tutti i file in `data/ita/lists/` seguono formattazione standardizzata per parsing consistente:

### Gerarchia Header
- **H1** (`#`) - Titolo file
- **H2** (`##`) - Singole voci entità  
- **H3** (`###`) - Sottosezioni entità (Tratti, Azioni, ecc.)

### Formattazione Campi
- **Campi regolari**: `**Campo:** valore`
- **Statistiche mostri/animali**: `- **Campo:** valore` (punti elenco)

### Formato Tabella (Mostri/Animali)
```markdown
| Caratteristica | Valore | Modificatore | Tiro Salvezza |
|----------------|--------|--------------|---------------|
| FOR | 21 | +5 | +5 |
| DES | 9 | -1 | +3 |
```

### Formati Metadati
- **Incantesimi**: `*Livello 2 Invocazione (Mago)*` o `*Trucchetto di Invocazione (Stregone, Mago)*`
- **Oggetti Magici**: `*Oggetto meraviglioso, molto raro (richiede sintonia)*`
- **Mostri/Animali**: `*Aberrazione Grande, Legale Malvagio*`
- **Talenti**: `*Talento di Origine*` o `*Talento Generale (Prerequisito: Livello 4+)*`

### Esempio Struttura Entità
```markdown
## Nome Entità

*Metadati in corsivo*

**Campo Standard:** Valore del campo

- **Campo Mostro:** Valore con punto elenco

| Tabella | Se | Necessaria |
|---------|----| ---------- |
| Riga 1  | 10 | +5         |

### Sottosezione (se necessaria)

Contenuto della sottosezione.
```

### Variabili Ambiente

| Variabile | Default | Descrizione |
|----------|---------|-------------|
| `MONGO_URI` | `mongodb://localhost:27017` | Stringa connessione MongoDB |
| `DB_NAME` | `dnd` | Nome database |
| `PORT` | `8000` | Porta servizio |
| `ENVIRONMENT` | `development` | Ambiente runtime |

## 📝 Gestione Contenuti

### Struttura Dati

Il sistema elabora contenuti D&D da file markdown organizzati per lingua:

**Contenuti Italiani (`data/ita/`):**
- Traduzione italiana completa dei contenuti SRD
- Terminologia localizzata e regole
- Adattamenti culturali dove appropriato

### Elaborazione Contenuti

Il parser CLI automaticamente:
1. Legge file markdown dalla directory data
2. Estrae informazioni strutturate (titoli, sezioni, metadati)
3. Converte contenuti in formati multipli (markdown, HTML, testo semplice)
4. Memorizza in MongoDB con indicizzazione appropriata per ricerca
5. Mantiene controllo versione e tracciamento sorgente

## 🧪 Testing

### Unit Test
```bash
make test
```

### Integration Test
```bash
make test-integration
```

### Performance Benchmark
```bash
make benchmark
```

### Test Manuali
1. Avvia servizi: `make up`
2. Verifica viewer su http://localhost:8000
3. Testa elaborazione e visualizzazione contenuti

## 🚀 Deployment

### Deployment Produzione Docker

1. **Costruisce immagini produzione**:
   ```bash
   make build
   ```

2. **Avvia in modalità produzione**:
   ```bash
   ENVIRONMENT=production make up
   ```

3. **Inizializza database** (prima volta):
   ```bash
   # Esegui parser CLI per elaborare contenuti iniziali
   ./bin/cli-parser
   ```

### Backup/Ripristino Database

**Crea Backup:**
```bash
make seed-dump
# Crea backup con timestamp: dnd_backup_YYYYMMDD_HHMMSS.archive.gz
```

**Ripristina Backup:**
```bash
make seed-restore FILE=dnd_backup_20240904_120000.archive.gz
```

## 🔍 Monitoraggio e Debug

### Visualizza Log
```bash
make logs                    # Tutti i servizi
docker compose logs viewer  # Servizio specifico
docker compose logs -f viewer  # Segui log
```

### Accesso Database
```bash
make mongo-sh
# In shell MongoDB:
use dnd
db.documenti.find().limit(5)
db.classi.countDocuments()
```

### Health Check
- Health viewer: http://localhost:8000/health

## 📄 Licenza

Questo progetto contiene contenuti D&D 5e SRD licenziati sotto Creative Commons Attribution 4.0 International License (CC-BY-4.0). Vedere le sezioni informazioni legali nei file contenuto per dettagli licenza completi.

Il codice applicazione è disponibile sotto i termini specificati nella licenza progetto.

## 🤝 Contribuire

1. Fai fork del repository
2. Crea un branch feature: `git checkout -b feature/tua-feature`
3. Fai le tue modifiche seguendo lo stile codice esistente
4. Esegui test: `make test`
5. Esegui linting: `make lint`
6. Committa le tue modifiche: `git commit -m "feat: descrizione tua feature"`
7. Pusha al tuo fork: `git push origin feature/tua-feature`
8. Crea una pull request

### Stile Codice
- Segui best practice e convenzioni Go
- Usa principi clean architecture
- Scrivi test completi per nuove feature
- Aggiorna documentazione per modifiche significative

## 📚 Risorse Aggiuntive

- [D&D 5e SRD Ufficiale](https://dnd.wizards.com/resources/systems-reference-document)
- [Documentazione Go](https://golang.org/doc/)
- [Documentazione MongoDB](https://docs.mongodb.com/)
- [Documentazione Docker Compose](https://docs.docker.com/compose/)

## 🐛 Risoluzione Problemi

### Problemi Comuni

**I servizi non si avviano:**
```bash
# Controlla se le porte sono in uso
netstat -tulpn | grep :8000

# Resetta ambiente Docker
make down
docker system prune -f
make up
```

**Problemi connessione MongoDB:**
```bash
# Verifica che MongoDB sia in esecuzione
docker compose ps mongo

# Controlla log MongoDB
docker compose logs mongo

# Testa connessione
make mongo-sh
```

**Contenuto non appare:**
```bash
# Esegui parser CLI manualmente
./bin/cli-parser

# Controlla contenuto database
make mongo-sh
# In shell MongoDB: db.documenti.countDocuments()
```

Per supporto aggiuntivo, apri un issue nel repository GitHub.