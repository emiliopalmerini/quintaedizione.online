# 0002 – Stored sort key and UI indexes

Status: Accepted

Context
- The UI lists and neighbors navigation sort by a derived lowercase key built at query time via `$addFields`. This is convenient but adds per‑request CPU.
- The homepage queries `documenti` by `numero_di_pagina` and would benefit from a dedicated index.

Decision
- Compute and store a normalized `_sortkey_alpha` at ingest time using the first available of `slug`, `name`, `term`, `title`, `titolo`, `nome`, lower‑cased.
- Prefer the stored `_sortkey_alpha` in sorting pipelines, falling back to the previous expression if missing.
- Create the following indexes at application startup (best‑effort):
  - `documenti.numero_di_pagina` (asc)
  - `documenti_en.numero_di_pagina` (asc)
  - `documenti._sortkey_alpha` (asc)
  - `documenti_en._sortkey_alpha` (asc)

Consequences
- Faster list and neighbor queries when `_sortkey_alpha` is present.
- Index creation is idempotent and safe; absence of a collection is tolerated.
- Existing data without `_sortkey_alpha` still work via the original fallback.

