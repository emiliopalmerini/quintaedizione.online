# 0004 – Ricerca e breadcrumb

Status: Accepted

Context
- Ricerca trasversale e navigazione veloce tra documenti filtrati.

Decisione
- Filtri specifici per collezione lato backend + form HTMX lato UI.
- Ordinamento alfabetico consistente (name→term) per lista e prev/next.
- Breadcrumb con quicksearch inline (dropdown risultati) e placeholder col nome documento.

Conseguenze
- Navigazione rapida senza cambiare pagina.
- Consistenza tra viste, filtri preservati via querystring.

