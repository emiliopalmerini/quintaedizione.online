# quintaedizione.online

Sito web italiano per contenuti SRD di D&D 5a Edizione, strumenti per combattimenti, mappe e generatori casuali. I dati sono forniti dal modulo `quintaedizione-data-ita`, caricati da JSON strutturato e serviti da un'applicazione Go con store in memoria.

## Prerequisiti

- Go 1.25.2
- [templ](https://templ.guide): `go install github.com/a-h/templ/cmd/templ@latest`

## Setup

```bash
cp .env.example .env  # configura PORT (default: 8000)
make run              # genera template, formatta, avvia il server
```

## Comandi

| Comando              | Descrizione                        |
|----------------------|------------------------------------|
| `make run`           | Genera template, formatta, avvia   |
| `make test`          | Esegue i test                      |
| `make templ-generate`| Genera solo i template templ       |
| `make format`        | Formatta il codice                 |
| `make clean`         | Pulisce gli artefatti di build     |

## Come funziona

```
quintaedizione-data-ita → JSON embed.FS → Loader Go → Store in memoria → net/http → templ
```

I file JSON vengono embeddati nel modulo dati a compile time. Nessun database esterno: le sezioni dell'app usano repository in memoria costruiti all'avvio.

**Contenuti**: incantesimi, mostri, classi, background, equipaggiamenti, oggetti magici, talenti, regole, glossario, specie, mappe, calcolatore combattimenti e generatori casuali.

## Stack

- `net/http` della standard library come router e server HTTP
- [a-h/templ](https://github.com/a-h/templ) per i template server-side
- [gomarkdown/markdown](https://github.com/gomarkdown/markdown) per il rendering markdown
- HTMX e JavaScript vanilla per interazioni progressive
- CSS locale in `web/static/css`

## Struttura del progetto

```
cmd/app/             Entrypoint HTTP
internal/
  combattimenti/     Calcolatore incontri
  generatori/        Tabelle casuali
  mappe/             Galleria mappe
  srd/               Lettura, ricerca e rendering contenuti SRD
pkg/                 Helper condivisi per template, web e datastore
web/
  static/            Asset statici
  templates/         Template templ (.templ)
```

## Licenza

- **Codice**: [BSD 3-Clause](LICENSE)
- **Contenuto SRD**: [Creative Commons Attribution 4.0 (CC-BY-4.0)](https://creativecommons.org/licenses/by/4.0/) di Wizards of the Coast
