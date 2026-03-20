# ADR-014: Random Generators (Generatori Casuali)

**Status**: Proposed

## Context

We want to add a "Generatori" section to the site with random generators based on Sly Flourish's Lazy GM Resource Document tables. Each generator is a simple random table: the user clicks a button and gets a random result from a predefined list.

## Decision

### Architecture

Follow the existing hexagonal feature-module pattern (like `combattimenti` and `mappe`):

```
data/generatori/
  embed.go              # embed.FS for JSON tables
  *.json                # one JSON file per generator table
internal/generatori/
  domain/
    types.go            # Generator, Table domain types
  application/
    service.go          # random selection logic
  infrastructure/
    web/
      handlers/
        generator_handler.go  # HTTP handlers
      templates/
        home.templ       # landing page listing all generators
        generator.templ  # single generator page with roll button
        result.templ     # HTMX partial for random result
```

### Data Format

Each JSON file represents one generator table:

```json
{
  "id": "connettori",
  "name": "Connettori Casuali",
  "description": "Connettori tra stanze di un dungeon.",
  "die": "1d20",
  "items": [
    "Corridoio stretto",
    "Scale a spirale",
    ...
  ]
}
```

### Routes

- `GET /generatori/` — home page listing all generators
- `GET /generatori/{slug}` — single generator page
- `POST /generatori/{slug}/roll` — HTMX endpoint returning a random result

### UX

- Landing page shows a card grid of all available generators
- Each generator page has: title, description, a "Tira" (roll) button
- Button uses HTMX to fetch a random result inline (no page reload)
- Result displays with a brief animation

### Navigation

Add "Generatori" to the site nav bar.

## Inputs

- JSON table files embedded at compile time
- User clicks roll button

## Outputs

- Random item from the selected table

## Edge Cases

- Table with 0 items → should not happen (validated at load time)
- Multiple rapid clicks → each returns an independent random result

## Consequences

- New nav item in the site header
- New embedded data directory
- Tables will be added incrementally (one by one, translated to Italian)
