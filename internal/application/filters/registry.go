package filters

import (
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/collections"
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/filters"
)

type InMemoryFilterRegistry struct {
	filters []filters.FilterDefinition
}

func NewInMemoryFilterRegistry() *InMemoryFilterRegistry {
	return &InMemoryFilterRegistry{}
}

func (r *InMemoryFilterRegistry) AddFilter(filter filters.FilterDefinition) {
	r.filters = append(r.filters, filter)
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
	for _, filter := range r.filters {
		if filter.Name == name {
			return filter, true
		}
	}
	return filters.FilterDefinition{}, false
}

func (r *InMemoryFilterRegistry) GetAllFilters() ([]filters.FilterDefinition, error) {
	result := make([]filters.FilterDefinition, 0, len(r.filters))
	result = append(result, r.filters...)
	return result, nil
}
