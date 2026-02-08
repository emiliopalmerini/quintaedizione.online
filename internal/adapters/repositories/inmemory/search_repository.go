package inmemory

import (
	"context"
	"fmt"

	domainsearch "github.com/emiliopalmerini/quintaedizione.online/internal/domain/search"
	"github.com/emiliopalmerini/quintaedizione.online/internal/infrastructure/datastore"
)

type searchRepository struct {
	store *datastore.Store
}

// NewSearchRepository creates a SearchRepository backed by the in-memory store.
func NewSearchRepository(store *datastore.Store) domainsearch.SearchRepository {
	return &searchRepository{store: store}
}

func (r *searchRepository) GetSearchableItems(_ context.Context, collection string) ([]domainsearch.SearchableItem, error) {
	docs, _ := r.store.Query(collection, nil, 0, 0)

	items := make([]domainsearch.SearchableItem, 0, len(docs))
	for _, doc := range docs {
		id, _ := doc["_id"].(string)
		title, _ := doc["title"].(string)

		keywords := extractKeywords(doc)
		items = append(items, domainsearch.SearchableItem{
			ID:         id,
			Collection: collection,
			Title:      title,
			Keywords:   keywords,
		})
	}

	return items, nil
}

func extractKeywords(doc map[string]any) []string {
	keywordFields := []string{
		"scuola", "classe", "tipo", "categoria",
		"rarita", "livello", "taglia", "allineamento",
		"tipo_danno", "proprieta", "ambiente",
	}

	var keywords []string
	for _, field := range keywordFields {
		val, ok := doc[field]
		if !ok || val == nil {
			continue
		}

		switch v := val.(type) {
		case string:
			if v != "" {
				keywords = append(keywords, v)
			}
		case []any:
			for _, item := range v {
				if s, ok := item.(string); ok && s != "" {
					keywords = append(keywords, s)
				}
			}
		default:
			s := fmt.Sprintf("%v", v)
			if s != "" && s != "0" {
				keywords = append(keywords, s)
			}
		}
	}

	return keywords
}
