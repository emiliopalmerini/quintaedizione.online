package search

import (
	"sync"
	"time"

	domainsearch "github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/search"
)

type SearchIndex struct {
	mu         sync.RWMutex
	items      map[string][]domainsearch.SearchableItem
	lastUpdate time.Time
	ttl        time.Duration
}

func NewSearchIndex(ttl time.Duration) *SearchIndex {
	return &SearchIndex{
		items: make(map[string][]domainsearch.SearchableItem),
		ttl:   ttl,
	}
}

func (idx *SearchIndex) Get(collection string) []domainsearch.SearchableItem {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.items[collection]
}

func (idx *SearchIndex) Set(collection string, items []domainsearch.SearchableItem) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.items[collection] = items
	idx.lastUpdate = time.Now()
}

func (idx *SearchIndex) NeedsRefresh() bool {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.items) == 0 || time.Since(idx.lastUpdate) > idx.ttl
}

func (idx *SearchIndex) GetAll() map[string][]domainsearch.SearchableItem {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	result := make(map[string][]domainsearch.SearchableItem, len(idx.items))
	for k, v := range idx.items {
		result[k] = v
	}
	return result
}

func (idx *SearchIndex) Clear() {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.items = make(map[string][]domainsearch.SearchableItem)
	idx.lastUpdate = time.Time{}
}
