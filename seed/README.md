Seed database for local development

This folder holds a small MongoDB seed you can restore locally or in Docker.

Recommended: single archive dump (keeps types, single file)
- Create: `mongodump --db dnd --gzip --archive=seed/dnd.archive.gz`
- Restore: `mongorestore --gzip --archive=seed/dnd.archive.gz --drop`

Docker helpers
- Restore in the running mongo container:
  - `docker compose exec mongo sh -lc 'mongorestore --gzip --archive=/seed/dnd.archive.gz --drop'`

Automatic restore on first start
- Place `dnd.archive.gz` in this folder.
- The `seed/init/restore.sh` script (mounted into /docker-entrypoint-initdb.d)
  will restore the archive the first time the MongoDB volume is empty.

Alternative layouts
- Directory dump: put a BSON dump under `seed/dump/` and the init script will use it.
- JSON exports for review (types lost): keep them under `seed/json/` and import manually.

Notes
- Use Git LFS if the archive grows large (>100MB).
- Do NOT commit sensitive data. This seed should contain only SRD content.

