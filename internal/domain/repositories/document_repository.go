package repositories

import (
	"context"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain"
)

// DocumentPredicate is a function that tests whether a document matches a filter.
type DocumentPredicate = func(map[string]any) bool

// DocumentReader provides read operations for documents.
type DocumentReader interface {
	FindByID(ctx context.Context, id domain.DocumentID, collection string) (*domain.Document, error)
	FindAll(ctx context.Context, collection string, limit int) ([]*domain.Document, error)
	FindByFilters(ctx context.Context, collection string, filters map[string]any, limit int) ([]*domain.Document, error)
	FindMapByID(ctx context.Context, collection string, id string) (map[string]any, error)
	FindMaps(ctx context.Context, collection string, match DocumentPredicate, skip int64, limit int64) ([]map[string]any, int64, error)
}

// DocumentStatistics provides statistics and counting operations.
type DocumentStatistics interface {
	Count(ctx context.Context, collection string) (int64, error)
	GetAllCollectionStats(ctx context.Context) ([]map[string]any, error)
}

// DocumentNavigation provides navigation between adjacent documents.
type DocumentNavigation interface {
	GetAdjacentMaps(ctx context.Context, collection string, currentID string) (prevID *string, nextID *string, err error)
}

// DocumentRepository composes all document-related read operations.
type DocumentRepository interface {
	DocumentReader
	DocumentStatistics
	DocumentNavigation
}
