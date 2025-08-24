# 0002 – Modello dati su MongoDB

Status: Accepted

Context
- I dati SRD hanno struttura variabile (liste, oggetti, campi liberi), servono query flessibili.

Decisione
- Usare MongoDB con collezioni separate (spells, magic_items, monsters, classes, ...).
- Ricerca testuale con `$regex` su campi comuni (`name`, `term`, `title`, `description`).
- Per `classes`, il parser produce campi strutturati (es. `features_by_level`).

Conseguenze
- Schema flessibile e adatto a sorgenti non uniformi.
- Alcuni filtri per‑collezione necessari a livello applicativo.

