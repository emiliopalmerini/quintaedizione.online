# SRD Data Layout (IT/EN)

This folder contains both Italian and English SRD sources, split by major sections.

- `ita/..`: Italian SRD (5e 2024) files
- `eng/..`: English SRD 5.2 files

## Parser support and collections

The web parser currently supports:

- Italian full pages → Mongo `documenti`
- Italian classes → Mongo `classi`
- Italian backgrounds → Mongo `backgrounds`
- English full pages → Mongo `documenti_en`

Other English collections (e.g., spells, magic items as itemized docs) are not yet parsed — they appear as full pages in `documenti_en`.

The editor UI can switch homepage documents between IT/EN via `?lang=it|en`.

## Notes

- Filenames encode an optional leading page number (e.g., `01_...`) used as `numero_di_pagina`.
- Slugs are derived from filenames; titles come from the first H1 in the file, or from the slug when H1 is missing.

## License

All content is licensed under the Creative Commons Attribution 4.0 International License (CC-BY-4.0) as specified in the legal information sections.
