# ADR-025: Canonical Data Release Contract

## Status

Accepted

## Context

`quintaedizione-data-ita` owns canonical SRD identities, concepts, domain values,
relationships, structured content, provenance, corrections, and validation. The
website must consume immutable releases without depending on the producer's Go
packages, repository layout, search indexes, storage, or legacy embedded JSON
files.

The canonical domain deliberately does not prescribe transport or artifact
layout. The producer and this consumer therefore need a small, versioned release
projection at their boundary.

## Decision

Canonical releases are exposed to the website as an `fs.FS` containing a strict
`manifest.json` and checksummed JSON resources. The filesystem may be backed by
an embedded module, a directory, or an archive; that delivery choice is not
visible to application code.

### Manifest

```json
{
  "formatVersion": 1,
  "schemaVersion": "1.0.0",
  "datasetVersion": "2026.07.1",
  "resources": [
    {
      "path": "entities/spells.json",
      "recordKind": "spell",
      "mediaType": "application/json",
      "sha256": "d1c5..."
    }
  ]
}
```

- `formatVersion` versions the release envelope and resource packaging.
- `schemaVersion` is the canonical model's semantic version.
- `datasetVersion` identifies an immutable snapshot as `YYYY.MM.REVISION`.
- `resources` declares every payload consumed from the release.
- `recordKind` identifies the canonical record type contained in a resource;
  consumers discover resources by this value rather than by their paths.
- Every resource payload is a JSON array containing records of its declared
  `recordKind`; object and `null` payloads are invalid.
- Resource paths are unique, relative, slash-separated paths without traversal.
- Resources are ordered lexicographically by path to keep publication
  deterministic.
- `sha256` is the lowercase hexadecimal digest of the resource bytes exactly as
  published.
- JSON objects reject unknown fields. Contract expansion therefore requires a
  compatible format revision rather than being silently ignored.

The release itself is immutable. Publishing different bytes under an existing
dataset version is invalid even if the manifest checksum is updated. Resource
digests prove internal consistency, not authenticity or version uniqueness;
deployments also pin the producer module revision or archive digest.

### Compatibility

The website declares the release format versions and canonical schema major
versions it supports. Startup or compilation fails before building projections
when either is unsupported, a manifest is malformed, a resource path is unsafe,
or a checksum differs.

Minor and patch schema changes remain acceptable within a supported major
version. Entity adapters may still reject a release when newly required domain
data cannot be projected safely.

### Consumer Projections

After verification, the website compiles release resources into disposable,
website-owned projections:

- permanent route mappings keyed by canonical entity UUID;
- concept and edition indexes keyed by `conceptId`;
- collection and facet views;
- rendered content;
- search indexes;
- SEO metadata and sitemaps;
- a read-optimized runtime artifact.

The website does not correct canonical records. Data defects are fixed and
republished by `quintaedizione-data-ita` as a new dataset version.

## Consequences

- The website is independent of the producer's implementation language.
- Embedded modules, downloaded archives, and local fixtures use the same
  consumer API.
- Corrupt or incompatible releases fail early and visibly.
- Resource layout is versioned as transport, not mistaken for canonical domain
  structure.
- A producer release projection must be published before the website can replace
  its legacy data adapters.
