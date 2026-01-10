package services

import (
	"fmt"
	"strconv"

	"github.com/emiliopalmerini/quintaedizione.online/internal/application/filters"
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/collections"
	domainFilters "github.com/emiliopalmerini/quintaedizione.online/internal/domain/filters"
)

type FilterService struct {
	registry     domainFilters.FilterRepository
	mongoBuilder *filters.MongoFilterBuilder
}

func NewFilterService(registry domainFilters.FilterRepository) *FilterService {
	return &FilterService{
		registry:     registry,
		mongoBuilder: filters.NewMongoFilterBuilder(),
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

func (s *FilterService) BuildMongoFilter(filterSet *domainFilters.FilterSet) (map[string]any, error) {
	if filterSet == nil {
		return map[string]any{}, nil
	}

	if err := s.ValidateFilterSet(filterSet); err != nil {
		return nil, fmt.Errorf("invalid filter set: %w", err)
	}

	return s.mongoBuilder.BuildFilter(filterSet)
}

func (s *FilterService) GetAvailableFilters(collection collections.CollectionName) ([]domainFilters.FilterDefinition, error) {
	if !collections.IsValid(collection.String()) {
		return nil, fmt.Errorf("invalid collection: %s", collection)
	}

	return s.registry.GetFiltersForCollection(collection)
}

func (s *FilterService) BuildSearchFilter(collection collections.CollectionName, searchTerm string) map[string]any {
	return s.mongoBuilder.BuildSearchFilter(collection, searchTerm)
}

func (s *FilterService) CombineFilters(fieldFilter, searchFilter map[string]any) map[string]any {
	var conditions []map[string]any

	if len(fieldFilter) > 0 {
		conditions = append(conditions, fieldFilter)
	}

	if len(searchFilter) > 0 {
		conditions = append(conditions, searchFilter)
	}

	if len(conditions) == 0 {
		return map[string]any{}
	} else if len(conditions) == 1 {
		return conditions[0]
	} else {
		return map[string]any{"$and": conditions}
	}
}

func (s *FilterService) convertValue(value string, dataType domainFilters.FilterDataType) (any, error) {
	switch dataType {
	case domainFilters.StringFilter, domainFilters.EnumFilter:
		return value, nil
	case domainFilters.NumberFilter:
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue, nil
		}
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue, nil
		}
		return nil, fmt.Errorf("invalid number format: %s", value)
	case domainFilters.BooleanFilter:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("invalid boolean format: %s", value)
		}
		return boolValue, nil
	default:
		return value, nil
	}
}
