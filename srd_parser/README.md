# SRD Parser

Parser e ingest dei contenuti SRD in MongoDB.

## Obiettivo
- Convertire le fonti SRD in documenti normalizzati per le collezioni del DB.
- Estrarre strutture utili (es. `features_by_level` per le classi) per una resa migliore nello show.

## Struttura
- `parsers/*.py`: parser specifici per dominio (incantesimi, mostri, classi, ...)
- `ingest.py`: entrypoint di import, orchestrazione
- `utils.py`: helpers comuni (split, normalizzazione, label sorgente)

## Esecuzione con Docker
```
docker compose --profile parser up srd-parser
```
Usa `MONGO_URI`, `DB_NAME` e `INPUT_DIR` dal `docker-compose.yml`.

## Decisivi per le classi
Il parser delle classi produce:
- `core_traits` + `core_traits_md`
- `features_by_level` (privilegi per livello, con `name` e `text`)
- `spellcasting_progression.by_level` (trucchetti, preparati, slot)
- `spell_lists_by_level` (incantesimi per livello)

La vista `show_class.html` usa questi campi per un rendering ricco.

