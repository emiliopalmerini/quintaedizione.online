# 0007 – Rendering Markdown lato client

Status: Accepted

Context
- Molti testi SRD sono markdown. Serve una resa leggibile senza pipeline di build.

Decisione
- Usare Marked.js via CDN, trasformando i blocchi con `data-markdown`.
- Tipografia CSS dedicata per titoli, liste, citazioni, codice, tabelle, link.
- Pulsante “Copia Markdown” per recuperare il sorgente.

Conseguenze
- Zero build tool aggiuntivi, resa consistente.
- Fallback testuale se la libreria non carica.

