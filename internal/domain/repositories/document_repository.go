package repositories

import (
	"context"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// DocumentRepository handles all document CRUD operations
// This replaces all entity-specific repositories with a unified interface
type DocumentRepository interface {
	// Create inserts a new document
	Create(ctx context.Context, doc *domain.Document, collection string) error

	// Update updates an existing document
	Update(ctx context.Context, doc *domain.Document, collection string) error

	// Delete removes a document by ID
	Delete(ctx context.Context, id domain.DocumentID, collection string) error

	// FindByID retrieves a document by ID
	FindByID(ctx context.Context, id domain.DocumentID, collection string) (*domain.Document, error)

	// FindAll retrieves all documents in a collection
	FindAll(ctx context.Context, collection string, limit int) ([]*domain.Document, error)

	// FindByFilters retrieves documents matching filters
	FindByFilters(ctx context.Context, collection string, filters map[string]any, limit int) ([]*domain.Document, error)

	// Count returns the number of documents in a collection
	Count(ctx context.Context, collection string) (int64, error)

	// UpsertMany performs bulk upsert operations
	UpsertMany(ctx context.Context, collection string, documents []*domain.Document) (int, error)

	// UpsertManyMaps performs bulk upsert from raw maps (for parser compatibility)
	UpsertManyMaps(ctx context.Context, collection string, uniqueFields []string, docs []map[string]any) (int, error)

	// GetCollectionStats returns statistics for a collection
	GetCollectionStats(ctx context.Context, collection string) (map[string]any, error)

	// GetAdjacentDocuments retrieves the previous and next documents
	GetAdjacentDocuments(ctx context.Context, collection string, currentID domain.DocumentID) (prev *domain.Document, next *domain.Document, err error)
}
