package monster

import "context"

// Monster is the narrow read model needed for encounter-budget pricing.
type Monster struct {
	ID     string
	Source string
	Name   string
	Type   string // creature type (Drago, Non morto, Umanoide, ...)
	Size   string // Piccola, Media, Grande, ...
	CR     string
	XP     int
	HP     string // raw "150 (20d10 + 40)" — kept as string so dice notation survives
	AC     string // raw "17" or "17 (armatura naturale)"
	Speed  string // raw "3 m, nuoto 12 m"
}

// SearchQuery describes a monster search filtered by edition and taxonomy.
//
// MinCR and MaxCR are inclusive numeric bounds; 0 means "no bound" on either
// side (so MinCR=0,MaxCR=0 yields all monsters). Type filters by exact (case-
// insensitive) match against Monster.Type.
//
// Note: an Environment filter was considered but the SRD JSON carries no
// habitat metadata, so it is omitted intentionally.
type SearchQuery struct {
	Source string
	Query  string
	MinCR  float64
	MaxCR  float64
	Type   string
	Limit  int
}

// FacetSet enumerates the distinct filterable values present in a source's
// monster collection. Used to populate the picker's dropdowns.
type FacetSet struct {
	Types []string
}

// Reader is the port the combattimenti slice uses to read monster data.
// Implementations live in infrastructure; the canonical one wraps the SRD
// DocumentRepository so the combattimenti slice does not own monster data.
type Reader interface {
	Search(ctx context.Context, q SearchQuery) ([]Monster, error)
	FindByID(ctx context.Context, source, id string) (Monster, error)
	Facets(ctx context.Context, source string) (FacetSet, error)
}
