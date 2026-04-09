# ADR-023: Deduplicate Multi-Source Listings

## Status

Accepted

## Context

With multi-source data (5e + 5.5e), collection browse pages and search results show two entries for every document that exists in both editions (e.g., two "Barbaro" rows in `/srd/classi`). ADR-022 added a version switcher on the detail page, so duplicate listings are now redundant; users can switch editions from the detail page.

## Decision

Add a deduplication predicate to the Store that filters out non-preferred duplicate documents during queries. A document is a "non-preferred duplicate" if its slug exists in multiple sources and its source is not the preferred (default) one.

### Inputs

- `collection` (string): collection name
- `preferredSource` (string): the default source short name (e.g., "5.5e")

### Output

A `DocumentPredicate` that, when combined with existing filter/search predicates, excludes non-preferred duplicates from query results.

### Behavior

1. The Store exposes a `DeduplicatePredicate(collection, preferredSource)` method that returns a predicate using the slug index.
2. The predicate logic:
   - If the document's source matches `preferredSource`, always include it.
   - If the document's source is different, include it only when no preferred-source version exists (i.e., the slug has a single version).
3. The ContentService receives `defaultSource` at construction and applies the dedup predicate to:
   - `GetCollectionItems` (collection browse)
   - `GlobalSearch` (search results)
   - `GetFacetCounts` (filter facet counts, to avoid inflated counts)
4. `GetItem` and `GetAdjacentItems` are NOT affected; they operate on specific composite IDs.
5. `GetItemVersions` is NOT affected; it returns all versions for the switcher.

### Edge Cases

- Single-source mode (only 5.5e loaded): dedup predicate is a no-op (all docs pass).
- Document exists only in 5e (not in 5.5e): included because it has no duplicate.
- `preferredSource` is empty: predicate passes all documents (no dedup).

### Error Conditions

None; the predicate is a pure function with no failure mode.

## Consequences

- Collection lists and search show exactly one entry per unique document
- Filter facet counts reflect deduplicated totals
- The preferred version's URL is used in list row links
- No data loss; non-preferred versions are still accessible via the version switcher on the detail page
