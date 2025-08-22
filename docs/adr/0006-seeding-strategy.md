# 0006 – Strategia di seeding del DB

Status: Accepted

Context
- Serve un seed di dati dev, versionabile e rapido da ripristinare.

Decisione
- Usare archivio compresso unico `seed/dnd.archive.gz` (mongodump --archive --gzip).
- Script `seed/init/restore.sh` eseguito da Mongo quando il volume è vuoto.
- Alternative: dump directory `seed/dump/` o JSON per revisione manuale.

Conseguenze
- Setup dev veloce, seed nel repo (eventuale Git LFS).
- Ripristino idempotente e non invasivo in ambienti già popolati.

