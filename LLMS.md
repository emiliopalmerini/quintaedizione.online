# LLMs

Linee guida per l'uso di LLM a supporto del progetto.

## Cosa chiedere agli LLM
- Generare/raffinare template Jinja2/HTMX, CSS leggeri.
- Suggerire refactor limitati e localizzati.
- Scrivere documentazione (README, ADR) e checklist QA.

## Cosa evitare
- Non generare migrazioni strutturali senza ADR.
- Non introdurre librerie senza motivazione e consenso.
- Non manipolare direttamente dati di produzione.

## Policy commit
- Conventional Commits, PR piccole, test manuale ove possibile.
- Descrivere sempre il razionale nel body del commit o nell’ADR.

