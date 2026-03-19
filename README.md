# quintaedizione.online

Sito web con i contenuti del Systems Reference Document (SRD) di D&D 5a Edizione, tradotti in italiano. Tutto il contenuto è estratto dal PDF ufficiale, convertito in JSON strutturato e servito come applicazione web.

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
PDF SRD → Parser Python → JSON → Binario Go (embed.FS) → Store in-memory → Web
```

I file JSON in `data/ita/json/` vengono embeddati nel binario a compile time. Nessun database esterno: tutto vive in memoria.

**Contenuti**: incantesimi, mostri, animali, classi, background, equipaggiamenti, oggetti magici, armi, armature, strumenti, cavalcature e veicoli, servizi, talenti, regole.

## Stack

- [gin-gonic/gin](https://github.com/gin-gonic/gin) — HTTP framework
- [a-h/templ](https://github.com/a-h/templ) — Template engine
- [gomarkdown/markdown](https://github.com/gomarkdown/markdown) — Rendering markdown
- [due-draghi-design-system](https://github.com/duedraghi/design-system) — CSS

## Struttura del progetto

```
cmd/viewer/          Entrypoint
data/ita/json/       Dati JSON (source of truth)
internal/
  adapters/          Interfacce esterne (web handler, repository in-memory)
  application/       Logica applicativa (servizi, filtri, parser)
  domain/            Modelli di dominio, interfacce repository
  infrastructure/    Config, datastore, loader JSON
web/
  static/            Asset statici
  templates/         Template templ (.templ)
scripts/parse_srd/   Parser Python per estrarre dati dal PDF SRD
```

## Licenza

- **Codice**: [BSD 3-Clause](LICENSE)
- **Contenuto SRD**: [Creative Commons Attribution 4.0 (CC-BY-4.0)](https://creativecommons.org/licenses/by/4.0/) di Wizards of the Coast
