# 0005 – Deployment e Docker Compose

Status: Accepted

Context
- Ambiente locale semplice e riproducibile, parser opzionale.

Decisione
- Docker Compose con servizi `mongo`, `editor`, `srd-parser` dietro profilo `parser`.
- Montare `seed/` per ripristino iniziale automatico se il volume è vuoto.

Conseguenze
- `docker compose up` avvia solo DB+editor; il parser si abilita con `--profile parser`.
- Ripristino seed trasparente al primo avvio.

