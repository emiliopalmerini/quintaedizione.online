# ADR-026: Permanent Content Route Registry

## Status

Accepted

## Context

Canonical entity identities are permanent UUIDs, while names can be corrected or
retranslated. Website paths are a consumer concern and must remain stable for
bookmarks, inbound links, and search engines. Deriving paths from the current
name on every build would silently break URLs after legitimate data corrections.

## Decision

The website owns a permanent route registry keyed by canonical entity UUID.
Each active mapping contains one canonical absolute path and zero or more
historical aliases.

The registry:

- accepts only identities present in the compiled canonical catalog;
- permits exactly one canonical route per entity;
- rejects relative, non-normalized, query-bearing, and fragment-bearing paths;
- rejects collisions across canonical paths and aliases;
- resolves aliases to their canonical path for permanent redirects;
- returns defensive copies so validated mappings cannot be mutated at runtime.

Names and slugs are never used as relational or canonical identities. Tooling may
propose a slug for a newly discovered entity, but the resulting UUID-to-path
mapping becomes reviewed website configuration and does not change
automatically when canonical content changes.

Tombstoned identities cannot receive active routes. Their historical behavior
will be defined explicitly during migration, using a replacement relation or a
retired-content response rather than silently reassigning the URL.

## Consequences

- Translation corrections do not change public URLs.
- Existing `/srd` paths can remain aliases while a new information architecture
  uses canonical `/compendio` paths.
- Route changes are explicit, reviewable, and testable.
- Publication tooling must report active canonical entities without a route and
  stale mappings that reference retired or absent identities.
