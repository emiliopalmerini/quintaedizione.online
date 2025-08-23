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
Il container del parser avvia di default la TUI interattiva.
Usa le variabili `MONGO_URI`, `DB_NAME` e `INPUT_DIR` dal `docker-compose.yml`.

Se modifichi il codice Python, ricostruisci l'immagine per propagare i cambi:

```
make tui-build
make tui
```

## TUI per parsing selettivo
È disponibile una semplice TUI a terminale per lanciare i parser in modalità selettiva (es. solo "classes") e in dry‑run o con upsert su Mongo.

Esecuzione locale:

```
python -m srd_parser.tui
```

Con Docker:

```
docker compose run --rm srd-tui
# oppure via Makefile
make tui       # avvia solo la TUI
make tui-up    # avvia mongo e poi la TUI
```

Suggerimento: per provare rapidamente le modifiche senza Docker, usa:

```
make tui-local
```

Tasti principali:
- Frecce/j,k: muovi • Spazio: seleziona • a: tutto • n: niente
- d: dry‑run ON/OFF • u: upsert (dry‑run OFF)
- e: cambia input dir • m: Mongo URI • b: DB name
- Invio/r: esegui • q: esci

Nota: per l'upsert è necessario `pymongo`. In dry‑run non serve.

## Decisivi per le classi
Il parser delle classi produce:
- `core_traits` + `core_traits_md`
- `features_by_level` (privilegi per livello, con `name` e `text`)
- `spellcasting_progression.by_level` (trucchetti, preparati, slot)
- `spell_lists_by_level` (incantesimi per livello)

La vista `show_class.html` usa questi campi per un rendering ricco.
