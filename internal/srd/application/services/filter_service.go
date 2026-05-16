package services

import (
	"fmt"

	appFilters "github.com/emiliopalmerini/quintaedizione.online/internal/srd/application/filters"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/collections"
	domainFilters "github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/filters"
)

type FilterService struct {
	registry         domainFilters.FilterRepository
	predicateBuilder *appFilters.PredicateBuilder
}

func NewFilterService(registry domainFilters.FilterRepository) *FilterService {
	return &FilterService{
		registry:         registry,
		predicateBuilder: appFilters.NewPredicateBuilder(),
	}
}

func (s *FilterService) ParseFilters(collection collections.CollectionName, queryParams map[string]string) (*domainFilters.FilterSet, error) {
	if !collections.IsValid(collection.String()) {
		return nil, fmt.Errorf("invalid collection: %s", collection)
	}

	filterSet := domainFilters.NewFilterSet(collection)

	available, err := s.registry.GetFiltersForCollection(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to get available filters: %w", err)
	}
	byName := make(map[string]domainFilters.FilterDefinition, len(available))
	for _, f := range available {
		byName[f.Name] = f
	}

	for name, value := range queryParams {
		if value == "" {
			continue
		}
		def, ok := byName[name]
		if !ok {
			// Allow collection-agnostic filters (e.g., _source_short) that
			// aren't applicable to this collection to silently no-op.
			continue
		}
		filterSet.AddFilter(domainFilters.FilterValue{Definition: def, Value: value})
	}

	return filterSet, nil
}

func (s *FilterService) BuildFilter(filterSet *domainFilters.FilterSet) (domainFilters.DocumentPredicate, error) {
	if filterSet == nil {
		return nil, nil
	}
	return s.predicateBuilder.BuildPredicate(filterSet)
}

func (s *FilterService) GetAvailableFilters(collection collections.CollectionName) ([]domainFilters.FilterDefinition, error) {
	if !collections.IsValid(collection.String()) {
		return nil, fmt.Errorf("invalid collection: %s", collection)
	}
	return s.registry.GetFiltersForCollection(collection)
}

func (s *FilterService) BuildSearchPredicate(_ collections.CollectionName, searchTerm string) domainFilters.DocumentPredicate {
	return s.predicateBuilder.BuildSearchPredicate(searchTerm)
}

func (s *FilterService) CombinePredicates(predicates ...domainFilters.DocumentPredicate) domainFilters.DocumentPredicate {
	active := make([]domainFilters.DocumentPredicate, 0, len(predicates))
	for _, p := range predicates {
		if p != nil {
			active = append(active, p)
		}
	}
	switch len(active) {
	case 0:
		return nil
	case 1:
		return active[0]
	}
	return func(doc map[string]any) bool {
		for _, p := range active {
			if !p(doc) {
				return false
			}
		}
		return true
	}
}
