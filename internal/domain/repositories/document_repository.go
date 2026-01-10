package repositories

import (
	"context"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain"
)

// DocumentReader provides read operations for documents.
type DocumentReader interface {
	FindByID(ctx context.Context, id domain.DocumentID, collection string) (*domain.Document, error)
	FindAll(ctx context.Context, collection string, limit int) ([]*domain.Document, error)
	FindByFilters(ctx context.Context, collection string, filters map[string]any, limit int) ([]*domain.Document, error)
	FindMapByID(ctx context.Context, collection string, id string) (map[string]any, error)
	FindMaps(ctx context.Context, collection string, filter map[string]any, skip int64, limit int64) ([]map[string]any, int64, error)
}

// DocumentWriter provides write operations for documents.
type DocumentWriter interface {
	Create(ctx context.Context, doc *domain.Document, collection string) error
	Update(ctx context.Context, doc *domain.Document, collection string) error
	Delete(ctx context.Context, id domain.DocumentID, collection string) error
}

// DocumentBulkOperations provides bulk operations for documents.
type DocumentBulkOperations interface {
	UpsertMany(ctx context.Context, collection string, documents []*domain.Document) (int, error)
	UpsertManyMaps(ctx context.Context, collection string, uniqueFields []string, docs []map[string]any) (int, error)
}

// DocumentStatistics provides statistics and counting operations.
type DocumentStatistics interface {
	Count(ctx context.Context, collection string) (int64, error)
	CountWithFilter(ctx context.Context, collection string, filter map[string]any) (int64, error)
	GetCollectionStats(ctx context.Context, collection string) (map[string]any, error)
	GetAllCollectionStats(ctx context.Context) ([]map[string]any, error)
}

// DocumentNavigation provides navigation between adjacent documents.
type DocumentNavigation interface {
	GetAdjacentDocuments(ctx context.Context, collection string, currentID domain.DocumentID) (prev *domain.Document, next *domain.Document, err error)
	GetAdjacentMaps(ctx context.Context, collection string, currentID string) (prevID *string, nextID *string, err error)
}

// DocumentCollectionManager provides collection management operations.
type DocumentCollectionManager interface {
	DropCollection(ctx context.Context, collection string) error
}

// DocumentRepository composes all document-related operations.
// This interface exists for backward compatibility and convenience.
// Prefer using specific interfaces (DocumentReader, DocumentWriter, etc.) when possible.
type DocumentRepository interface {
	DocumentReader
	DocumentWriter
	DocumentBulkOperations
	DocumentStatistics
	DocumentNavigation
	DocumentCollectionManager
}
