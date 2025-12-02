package repositories

import (
	"context"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain"
)

type DocumentRepository interface {
	Create(ctx context.Context, doc *domain.Document, collection string) error

	Update(ctx context.Context, doc *domain.Document, collection string) error

	Delete(ctx context.Context, id domain.DocumentID, collection string) error

	FindByID(ctx context.Context, id domain.DocumentID, collection string) (*domain.Document, error)

	FindAll(ctx context.Context, collection string, limit int) ([]*domain.Document, error)

	FindByFilters(ctx context.Context, collection string, filters map[string]any, limit int) ([]*domain.Document, error)

	Count(ctx context.Context, collection string) (int64, error)

	UpsertMany(ctx context.Context, collection string, documents []*domain.Document) (int, error)

	UpsertManyMaps(ctx context.Context, collection string, uniqueFields []string, docs []map[string]any) (int, error)

	GetCollectionStats(ctx context.Context, collection string) (map[string]any, error)

	GetAdjacentDocuments(ctx context.Context, collection string, currentID domain.DocumentID) (prev *domain.Document, next *domain.Document, err error)

	FindMapByID(ctx context.Context, collection string, id string) (map[string]any, error)

	FindMaps(ctx context.Context, collection string, filter map[string]any, skip int64, limit int64) ([]map[string]any, int64, error)

	CountWithFilter(ctx context.Context, collection string, filter map[string]any) (int64, error)

	GetAdjacentMaps(ctx context.Context, collection string, currentID string) (prevID *string, nextID *string, err error)

	GetAllCollectionStats(ctx context.Context) ([]map[string]any, error)

	DropCollection(ctx context.Context, collection string) error
}
