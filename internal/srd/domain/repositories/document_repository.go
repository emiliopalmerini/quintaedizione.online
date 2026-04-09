package repositories

import (
	"context"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain"
)

// DocumentPredicate is a function that tests whether a raw document map matches a filter.
// Predicates operate on map[string]any because they run at the Store level.
type DocumentPredicate = func(map[string]any) bool

// DocumentReader provides read operations for documents.
type DocumentReader interface {
	FindByID(ctx context.Context, collection string, id string) (*domain.Document, error)
	FindByPredicate(ctx context.Context, collection string, match DocumentPredicate, skip int64, limit int64) ([]*domain.Document, int64, error)
}

// DocumentStatistics provides statistics and counting operations.
type DocumentStatistics interface {
	Count(ctx context.Context, collection string) (int64, error)
	GetAllCollectionStats(ctx context.Context) ([]map[string]any, error)
	AggregateField(ctx context.Context, collection, fieldPath string, match DocumentPredicate) (map[string]int64, error)
}

// DocumentNavigation provides navigation between adjacent documents.
type DocumentNavigation interface {
	GetAdjacent(ctx context.Context, collection string, currentID string) (prevID *string, nextID *string, position int, total int, err error)
}

// DocumentVersions provides access to all editions of a document by slug.
type DocumentVersions interface {
	FindVersions(ctx context.Context, collection, slug string) ([]domain.VersionInfo, error)
	DeduplicatePredicate(collection, preferredSource string) DocumentPredicate
}

// DocumentRepository composes all document-related read operations.
type DocumentRepository interface {
	DocumentReader
	DocumentStatistics
	DocumentNavigation
	DocumentVersions
}
