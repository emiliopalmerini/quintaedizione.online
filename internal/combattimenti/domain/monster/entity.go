package monster

import "context"

// Monster is the narrow read model needed for encounter-budget pricing.
type Monster struct {
	ID     string
	Source string
	Name   string
	CR     string
	XP     int
}

// SearchQuery describes a monster search filtered by edition and budget.
type SearchQuery struct {
	Source     string
	Query      string
	MaxXP      int
	OnlyAfford bool
	Limit      int
}

// Reader is the port the combattimenti slice uses to read monster data.
// Implementations live in infrastructure; the canonical one wraps the SRD
// DocumentRepository so the combattimenti slice does not own monster data.
type Reader interface {
	Search(ctx context.Context, q SearchQuery) ([]Monster, error)
	FindByID(ctx context.Context, source, id string) (Monster, error)
}
