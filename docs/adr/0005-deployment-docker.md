# 0005 – Deployment e Docker Compose

Status: Accepted

Context
- Ambiente locale semplice e riproducibile.

Decisione
- Docker Compose con servizi `mongo`, `editor`, `srd-tui` (nessun profilo).
- L'immagine del parser usa come entrypoint la TUI (`python -m srd_parser.tui`).
- Montare `seed/` per ripristino iniziale automatico se il volume è vuoto.

Conseguenze
- `docker compose up -d mongo editor` avvia DB+editor; il parsing si lancia manualmente via TUI con `docker compose run --rm srd-tui`.
- Nessun rischio di parsing automatico non intenzionale; l'utente conferma dalla TUI.
- Ripristino seed trasparente al primo avvio.
