package persistence

import (
	"context"

	domainsearch "github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/search"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/infrastructure/datastore"
)

type searchRepository struct {
	store           *datastore.Store
	preferredSource string
}

// NewSearchRepository creates a SearchRepository backed by the in-memory store.
// When preferredSource is non-empty, documents that exist in multiple sources
// are deduplicated so only the preferred source's version is indexed.
func NewSearchRepository(store *datastore.Store, preferredSource string) domainsearch.SearchRepository {
	return &searchRepository{store: store, preferredSource: preferredSource}
}

func (r *searchRepository) GetSearchableItems(_ context.Context, collection string) ([]domainsearch.SearchableItem, error) {
	predicate := r.store.DeduplicatePredicate(collection, r.preferredSource)
	docs, _ := r.store.Query(collection, predicate, 0, 0)

	items := make([]domainsearch.SearchableItem, 0, len(docs))
	for _, doc := range docs {
		id, _ := doc["_id"].(string)
		title, _ := doc["title"].(string)
		short, _ := doc["_source_short"].(string)

		// Use composite key so search results can be looked up in the store
		compositeID := id
		if short != "" {
			compositeID = short + "/" + id
		}

		items = append(items, domainsearch.SearchableItem{
			ID:         compositeID,
			Collection: collection,
			Title:      title,
		})
	}

	return items, nil
}
