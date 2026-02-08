package filters

import (
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/collections"
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/filters"
)

type InMemoryFilterRegistry struct {
	filters map[string]filters.FilterDefinition
}

func NewInMemoryFilterRegistry() *InMemoryFilterRegistry {
	return &InMemoryFilterRegistry{
		filters: make(map[string]filters.FilterDefinition),
	}
}

func (r *InMemoryFilterRegistry) AddFilter(filter filters.FilterDefinition) {
	r.filters[filter.Name] = filter
}

func (r *InMemoryFilterRegistry) GetFiltersForCollection(collection collections.CollectionName) ([]filters.FilterDefinition, error) {
	result := make([]filters.FilterDefinition, 0)

	for _, filter := range r.filters {
		if filter.IsApplicableToCollection(collection) {
			result = append(result, filter)
		}
	}

	return result, nil
}

func (r *InMemoryFilterRegistry) GetFilterByName(name string) (filters.FilterDefinition, bool) {
	filter, exists := r.filters[name]
	return filter, exists
}

func (r *InMemoryFilterRegistry) GetAllFilters() ([]filters.FilterDefinition, error) {
	result := make([]filters.FilterDefinition, 0, len(r.filters))
	for _, filter := range r.filters {
		result = append(result, filter)
	}
	return result, nil
}
