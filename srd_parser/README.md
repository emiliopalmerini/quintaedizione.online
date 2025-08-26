# SRD Parser (Web)

Parser e ingest dei contenuti SRD in MongoDB, con interfaccia web minimale.

## Obiettivo
- Convertire le fonti SRD in documenti normalizzati per le collezioni del DB.
- Estrarre strutture utili (es. `features_by_level` per le classi) per una resa migliore nello show.

## Struttura
- `parsers/*.py`: parser specifici per dominio (incantesimi, mostri, classi, ...)
- `ingest.py`: funzioni di upsert e chiavi univoche
- `work.py`: elenco delle collezioni e file sorgente
- `web.py` + `templates/parser_form.html`: interfaccia web (FastAPI + Jinja + HTMX)

## Esecuzione con Docker
- Il servizio `srd-parser` espone la web app su `http://localhost:8100`.
- Variabili: `MONGO_URI`, `DB_NAME`, `INPUT_DIR` configurate in `docker-compose.yml`.
  L'interfaccia mostra solo l'istanza (host:port) e il nome DB; eventuali credenziali sono lette da `MONGO_URI` e non vengono esposte.

Ricostruisci l'immagine dopo modifiche al codice:

```
docker compose build srd-parser
docker compose up -d srd-parser
```

## Note
- Modalità dry‑run: analizza e mostra i totali senza scrivere su Mongo.
- Upsert: disattiva dry‑run per scrivere su Mongo (usa `pymongo`).

## Dettagli per le classi
Il parser delle classi produce:
- `core_traits` + `core_traits_md`
- `features_by_level` (privilegi per livello, con `name` e `text`)
- `spellcasting_progression.by_level` (trucchetti, preparati, slot)
- `spell_lists_by_level` (incantesimi per livello)

La vista dell'editor (`show_class.html`) sfrutta questi campi per un rendering ricco.
