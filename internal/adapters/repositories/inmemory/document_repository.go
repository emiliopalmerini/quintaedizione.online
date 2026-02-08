package inmemory

import (
	"context"
	"fmt"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain"
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/repositories"
	"github.com/emiliopalmerini/quintaedizione.online/internal/infrastructure/datastore"
)

type documentRepository struct {
	store *datastore.Store
}

// NewDocumentRepository creates a DocumentRepository backed by the in-memory store.
func NewDocumentRepository(store *datastore.Store) repositories.DocumentRepository {
	return &documentRepository{store: store}
}

func (r *documentRepository) FindByID(_ context.Context, id domain.DocumentID, collection string) (*domain.Document, error) {
	doc, err := r.store.Get(collection, string(id))
	if err != nil {
		return nil, err
	}

	title, _ := doc["title"].(string)
	content, _ := doc["content"].(string)
	rawContent, _ := doc["raw_content"].(string)

	return &domain.Document{
		ID:         id,
		Title:      title,
		Content:    domain.HTMLContent(content),
		RawContent: domain.MarkdownContent(rawContent),
	}, nil
}

func (r *documentRepository) FindAll(_ context.Context, collection string, limit int) ([]*domain.Document, error) {
	docs, _ := r.store.Query(collection, nil, 0, int64(limit))

	result := make([]*domain.Document, 0, len(docs))
	for _, doc := range docs {
		id, _ := doc["_id"].(string)
		title, _ := doc["title"].(string)
		content, _ := doc["content"].(string)
		rawContent, _ := doc["raw_content"].(string)
		result = append(result, &domain.Document{
			ID:         domain.DocumentID(id),
			Title:      title,
			Content:    domain.HTMLContent(content),
			RawContent: domain.MarkdownContent(rawContent),
		})
	}
	return result, nil
}

func (r *documentRepository) FindByFilters(_ context.Context, collection string, filters map[string]any, limit int) ([]*domain.Document, error) {
	match := func(doc map[string]any) bool {
		for key, value := range filters {
			if fmt.Sprintf("%v", doc[key]) != fmt.Sprintf("%v", value) {
				return false
			}
		}
		return true
	}

	docs, _ := r.store.Query(collection, match, 0, int64(limit))

	result := make([]*domain.Document, 0, len(docs))
	for _, doc := range docs {
		id, _ := doc["_id"].(string)
		title, _ := doc["title"].(string)
		content, _ := doc["content"].(string)
		rawContent, _ := doc["raw_content"].(string)
		result = append(result, &domain.Document{
			ID:         domain.DocumentID(id),
			Title:      title,
			Content:    domain.HTMLContent(content),
			RawContent: domain.MarkdownContent(rawContent),
		})
	}
	return result, nil
}

func (r *documentRepository) FindMapByID(_ context.Context, collection string, id string) (map[string]any, error) {
	return r.store.Get(collection, id)
}

func (r *documentRepository) FindMaps(_ context.Context, collection string, match repositories.DocumentPredicate, skip int64, limit int64) ([]map[string]any, int64, error) {
	docs, total := r.store.Query(collection, match, skip, limit)
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

func (r *documentRepository) GetAdjacentMaps(_ context.Context, collection string, currentID string) (prevID *string, nextID *string, err error) {
	prev, next := r.store.Adjacent(collection, currentID)
	return prev, next, nil
}
