package services

import (
	"fmt"

	appFilters "github.com/emiliopalmerini/quintaedizione.online/internal/application/filters"
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/collections"
	domainFilters "github.com/emiliopalmerini/quintaedizione.online/internal/domain/filters"
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

	availableFilters, err := s.registry.GetFiltersForCollection(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to get available filters: %w", err)
	}

	filterMap := make(map[string]domainFilters.FilterDefinition)
	for _, filter := range availableFilters {
		filterMap[filter.Name] = filter
	}

	for paramName, paramValue := range queryParams {
		if paramValue == "" {
			continue
		}

		filterDef, exists := filterMap[paramName]
		if !exists {
			filterDef, exists = s.registry.GetFilterByName(paramName)
			if !exists {
				continue
			}

			if !filterDef.IsApplicableToCollection(collection) {
				return nil, domainFilters.NewUnsupportedFilterError(paramName, collection)
			}
		}

		if err := filterDef.ValidateValue(paramValue); err != nil {
			return nil, fmt.Errorf("validation failed for filter %s: %w", paramName, err)
		}

		rawValue, err := s.convertValue(paramValue, filterDef.DataType)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value for filter %s: %w", paramName, err)
		}

		filterValue := domainFilters.FilterValue{
			Definition: filterDef,
			Value:      paramValue,
			RawValue:   rawValue,
		}

		filterSet.AddFilter(filterValue)
	}

	return filterSet, nil
}

func (s *FilterService) ValidateFilterSet(filterSet *domainFilters.FilterSet) error {
	if filterSet == nil {
		return fmt.Errorf("filter set cannot be nil")
	}

	if !collections.IsValid(filterSet.Collection.String()) {
		return fmt.Errorf("invalid collection: %s", filterSet.Collection)
	}

	for _, filterValue := range filterSet.Filters {
		if err := filterValue.Definition.ValidateValue(filterValue.Value); err != nil {
			return fmt.Errorf("validation failed for filter %s: %w", filterValue.Definition.Name, err)
		}

		if !filterValue.Definition.IsApplicableToCollection(filterSet.Collection) {
			return domainFilters.NewUnsupportedFilterError(filterValue.Definition.Name, filterSet.Collection)
		}
	}

	return nil
}

func (s *FilterService) BuildFilter(filterSet *domainFilters.FilterSet) (domainFilters.DocumentPredicate, error) {
	if filterSet == nil {
		return nil, nil
	}

	if err := s.ValidateFilterSet(filterSet); err != nil {
		return nil, fmt.Errorf("invalid filter set: %w", err)
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
	// Filter out nil predicates
	var active []domainFilters.DocumentPredicate
	for _, p := range predicates {
		if p != nil {
			active = append(active, p)
		}
	}

	if len(active) == 0 {
		return nil
	}
	if len(active) == 1 {
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

func (s *FilterService) convertValue(value string, dataType domainFilters.FilterDataType) (any, error) {
	return domainFilters.ConvertValue(value, dataType)
}
