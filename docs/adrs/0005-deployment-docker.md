# 0005 – Deployment e Docker Compose

Status: Accepted

Context
- Ambiente locale semplice e riproducibile.

Decisione
- Docker Compose con servizi `mongo`, `editor`, `srd-parser` (nessun profilo).
- L'immagine del parser espone una web app (FastAPI) su porta 8000 (mappata a 8100).
- Montare `seed/` per ripristino iniziale automatico se il volume è vuoto.

Conseguenze
- `docker compose up -d mongo editor srd-parser` avvia DB+editor+parser web; il parsing si gestisce da interfaccia web su `http://localhost:8100`.
- Ripristino seed trasparente al primo avvio.
