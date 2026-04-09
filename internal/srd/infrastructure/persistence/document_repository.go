package persistence

import (
	"context"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/repositories"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/infrastructure/datastore"
)

type documentRepository struct {
	store *datastore.Store
}

// NewDocumentRepository creates a DocumentRepository backed by the in-memory store.
func NewDocumentRepository(store *datastore.Store) repositories.DocumentRepository {
	return &documentRepository{store: store}
}

func (r *documentRepository) FindByID(_ context.Context, collection string, id string) (*domain.Document, error) {
	m, err := r.store.Get(collection, id)
	if err != nil {
		return nil, err
	}
	return domain.DocumentFromMap(m), nil
}

func (r *documentRepository) FindByPredicate(_ context.Context, collection string, match repositories.DocumentPredicate, skip int64, limit int64) ([]*domain.Document, int64, error) {
	maps, total := r.store.Query(collection, match, skip, limit)

	docs := make([]*domain.Document, 0, len(maps))
	for _, m := range maps {
		docs = append(docs, domain.DocumentFromMap(m))
	}
	return docs, total, nil
}

func (r *documentRepository) Count(_ context.Context, collection string) (int64, error) {
	return r.store.Count(collection), nil
}

func (r *documentRepository) GetAllCollectionStats(_ context.Context) ([]map[string]any, error) {
	var stats []map[string]any
	for _, name := range r.store.Collections() {
		stats = append(stats, map[string]any{
			"collection": name,
			"count":      r.store.Count(name),
		})
	}
	return stats, nil
}

func (r *documentRepository) AggregateField(_ context.Context, collection, fieldPath string, match repositories.DocumentPredicate) (map[string]int64, error) {
	return r.store.Aggregate(collection, fieldPath, match), nil
}

func (r *documentRepository) GetAdjacent(_ context.Context, collection string, currentID string) (prevID *string, nextID *string, position int, total int, err error) {
	prev, next, pos, tot := r.store.Adjacent(collection, currentID)
	return prev, next, pos, tot, nil
}

func (r *documentRepository) DeduplicatePredicate(collection, preferredSource string) repositories.DocumentPredicate {
	return r.store.DeduplicatePredicate(collection, preferredSource)
}

func (r *documentRepository) FindVersions(_ context.Context, collection, slug string) ([]domain.VersionInfo, error) {
	docs := r.store.GetBySlug(collection, slug)
	versions := make([]domain.VersionInfo, 0, len(docs))
	for _, doc := range docs {
		sourceShort, _ := doc["_source_short"].(string)
		id, _ := doc["_id"].(string)
		compositeID := id
		if sourceShort != "" {
			compositeID = sourceShort + "/" + id
		}
		versions = append(versions, domain.VersionInfo{
			SourceShort: sourceShort,
			CompositeID: compositeID,
		})
	}
	return versions, nil
}
