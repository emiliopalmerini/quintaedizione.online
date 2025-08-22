# 0001 – Stack applicativo e stile UI

Status: Accepted

Context
- Serve un’app veloce per cercare/visualizzare/modificare SRD con UX semplice.

Decisione
- FastAPI + Motor (MongoDB) per backend asincrono semplice.
- Jinja2 + HTMX per server‑render con interazioni progressive.
- Tailwind via CDN + CSS custom per evitare build tool complessi.

Conseguenze
- Sviluppo rapido, pochi layer di complessità.
- JS minimo; alcune funzioni (Markdown) via CDN.
- Trade‑off: meno SPA, ma migliore semplicità/robustezza.

