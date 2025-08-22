#!/usr/bin/env sh
set -e

# This script runs inside the mongo container when the data dir is EMPTY
# (docker-entrypoint will execute scripts in /docker-entrypoint-initdb.d).

echo "[seed] Checking for seed archives..."

if [ -f /seed/dnd.archive.gz ]; then
  echo "[seed] Restoring /seed/dnd.archive.gz"
  mongorestore --gzip --archive=/seed/dnd.archive.gz --drop
  exit 0
fi

if [ -d /seed/dump ]; then
  echo "[seed] Restoring dump directory /seed/dump"
  mongorestore --dir /seed/dump --drop
  exit 0
fi

echo "[seed] No seed found; skipping restore."

