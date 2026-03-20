package search

import "context"

type SearchableItem struct {
	ID         string
	Collection string
	Title      string
	Keywords   []string
}

type SearchResult struct {
	ID         string
	Collection string
	Title      string
	Score      int
}

type SearchResultSet struct {
	Collection string
	Results    []SearchResult
	Total      int64
}

type SearchRepository interface {
	GetSearchableItems(ctx context.Context, collection string) ([]SearchableItem, error)
}

type SearchService interface {
	Search(ctx context.Context, query string, limitPerCollection int) ([]SearchResultSet, error)
	SearchCollection(ctx context.Context, collection, query string, limit int) ([]SearchResult, error)
	RefreshIndex(ctx context.Context) error
}
