# 0002 – Stored sort key and UI indexes

**Status:** Accepted

## Context

* Le liste UI e la navigazione “vicini” ordinano con una chiave derivata e lowercased calcolata via `$addFields` a runtime. Comodo ma costoso.
* La homepage interroga `documenti` per `numero_di_pagina`; serve un indice dedicato.

## Decision

* Calcolare e **salvare** in ingest `_sortkey_alpha` lowercased usando il primo campo disponibile in questo ordine:
  `slug` → `name` → `term` → `title` → `titolo` → `nome`.
* In query e pipeline usare `_sortkey_alpha` quando presente; fallback alla vecchia espressione quando assente.
* Creare indici UI “best-effort” all’avvio applicazione.

## Implementazione

### Algoritmo `_sortkey_alpha`

Pseudocodice:

```js
function computeSortKey(doc) {
  const first = doc.slug ?? doc.name ?? doc.term ?? doc.title ?? doc.titolo ?? doc.nome ?? "";
  return first.toString().trim().toLowerCase();
}
```

MongoDB update per ingest/backfill:

```js
db.COLLECTION.updateMany(
  {},
  [{
    $set: {
      _sortkey_alpha: {
        $toLower: {
          $trim: { input: { $ifNull: [
            "$slug", {$ifNull:["$name", {$ifNull:["$term", {$ifNull:["$title", {$ifNull:["$titolo", {$ifNull:["$nome", ""]}]}]}]}]}
          ]}}
        }
      }
    }
  }]
);
```

### Query di ordinamento (preferita)

```js
db.COLLECTION.find(filter).sort({ _sortkey_alpha: 1, _id: 1 });
```

Fallback se mancante:

```js
db.COLLECTION.aggregate([
  { $addFields: { _k: { $toLower: { $ifNull: ["$slug","$name","$term","$title","$titolo","$nome",""] } } } },
  { $sort: { _k: 1, _id: 1 } }
]);
```

## Index plan

### documenti

```js
db.documenti.createIndex({ numero_di_pagina: 1 });
db.documenti.createIndex({ _sortkey_alpha: 1 });
```

### Collezioni SRD (tutte con ordinamento alfabetico UI)

Applicare l’indice `{ _sortkey_alpha: 1 }` alle seguenti:

* `classi`
* `backgrounds`
* `incantesimi`
* `armi`
* `armature`
* `strumenti`
* `servizi`
* `equipaggiamento`
* `oggetti_magici`
* `mostri`
* `animali`
* `talenti`
* `cavalcature_e_veicoli`

Esempio:

```js
for (const c of [
  "classi","backgrounds","incantesimi","armi","armature","strumenti",
  "servizi","equipaggiamento","oggetti_magici","mostri","animali",
  "talenti","cavalcature_e_veicoli"
]) {
  db.getCollection(c).createIndex({ _sortkey_alpha: 1 });
}
```

### Indici consigliati aggiuntivi

* Lookup veloce per dettaglio:

```js
db.COLLECTION.createIndex({ slug: 1 }, { unique: true, sparse: true });
```

* Homepage e paginazioni alfabetiche:

```js
db.COLLECTION.createIndex({ _sortkey_alpha: 1, _id: 1 });
```

## Migrazione / Backfill

1. Per ogni collection: eseguire l’aggiornamento che imposta `_sortkey_alpha`.
2. Creare gli indici. Creazioni idempotenti e tolleranti a collezioni assenti.
3. Lasciare attivo il fallback per i documenti legacy non ancora ricalcolati.

## Consequences

* Liste e “vicini” più rapide con `_sortkey_alpha`.
* Indici idempotenti e sicuri; nessun hard-fail se la collection non esiste.
* Dati esistenti senza `_sortkey_alpha` restano funzionanti via fallback.

