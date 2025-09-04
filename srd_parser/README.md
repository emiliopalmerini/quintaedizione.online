# SRD Parser (Web)

Parser e ingest dei contenuti SRD in MongoDB, con interfaccia web minimale.

## Obiettivo
- Convertire le fonti SRD in documenti normalizzati per le collezioni del DB.
- Estrarre strutture utili (es. `features_by_level` per le classi) per una resa migliore nello show.

## Architettura

Implementa **Hexagonal Architecture** con **Domain-Driven Design** per separazione pulita delle responsabilità.

### Struttura
- `domain/`: Entità di dominio, value objects, e servizi (DDD core)
- `parsers/*.py`: Parser specifici per dominio (incantesimi, mostri, classi, ...)
- `work.py`: Configurazione delle collezioni e file sorgente
- `application/`: Service layer con ingest runner e service
- `adapters/`: Adapter per persistenza MongoDB e interfacce esterne
- `web.py`: Interfaccia web principale (FastAPI + Jinja + HTMX)
- `web_hexagonal.py`: Implementazione alternativa con architettura hexagonal pura
- `templates/`: Template Jinja2 per l'interfaccia web

## Esecuzione con Docker
- Il servizio `srd-parser` espone la web app su `http://localhost:8100`.
- Variabili: `MONGO_URI`, `DB_NAME`, `INPUT_DIR` configurate in `docker-compose.yml`.
  L'interfaccia mostra solo l'istanza (host:port) e il nome DB; eventuali credenziali sono lette da `MONGO_URI` e non vengono esposte.

Ricostruisci l'immagine dopo modifiche al codice:

```
docker compose build srd-parser
docker compose up -d srd-parser
```

## Caratteristiche Chiave

### Domain-Driven Design
- **Rich Domain Model**: Entità con validation e business logic
- **Value Objects**: Oggetti immutabili (Level, Ability, ClassSlug)
- **Domain Services**: Logica di business complessa
- **Aggregates**: Coerenza dei dati attraverso aggregate roots

### Parser Features
- **Modalità dry‑run**: Analizza e mostra i totali senza scrivere su Mongo
- **Upsert operations**: Disattiva dry‑run per scrivere su Mongo (usa `pymongo`)
- **Structured output**: Parser delle classi produce dati strutturati ricchi
- **Web interface**: Interfaccia web user-friendly, nessun CLI

### Architecture Benefits
- **Testabilità**: Domain logic isolata e facilmente testabile
- **Flessibilità**: Adattatori intercambiabili per diverse persistenze
- **Manutenibilità**: Separazione pulita tra business logic e infrastruttura

## Dettagli per le classi
Il parser delle classi produce:
- `core_traits` + `core_traits_md`
- `features_by_level` (privilegi per livello, con `name` e `text`)
- `spellcasting_progression.by_level` (trucchetti, preparati, slot)
- `spell_lists_by_level` (incantesimi per livello)

La vista del visualizzatore (`show_class.html`) sfrutta questi campi per un rendering ricco.
