# Agents

Questo progetto può essere supportato da agenti (umani o LLM) per attività ripetitive:

- Ingest/normalizzazione dati: controllo consistenza campi, mapping chiavi, validazioni basate su schemi attesi.
- QA traduzione: verifica campi tradotti vs. originali, liste di controllo (terminologia). 
- Suggestioni UI/UX: miglioramenti incrementali dell’editor e delle viste.

Principi operativi
- Cambi piccoli e atomici, con Conventional Commits.
- Non introdurre dipendenze o tool complessi senza ADR.
- Preferire HTMX e server‑render per semplicità; JS solo dove serve.

Sicurezza e privacy
- Nessun dato sensibile; solo contenuti SRD.
- Evitare log di dati non necessari.

