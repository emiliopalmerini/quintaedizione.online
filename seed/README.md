# Seed

Helper directory mounted into the MongoDB container at `/seed`.

Use the Makefile targets to dump/restore the development database:

- `make seed-dump` → writes `/seed/dnd.archive.gz`
- `make seed-restore` → restores from `/seed/dnd.archive.gz`
- `make seed-dump-dir` → writes `/seed/dump/`
- `make seed-restore-dir` → restores from `/seed/dump/`

You can copy these files out of the container via the mounted volume.

