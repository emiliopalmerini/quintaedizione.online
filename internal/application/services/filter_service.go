package services

import (
	"fmt"
	"strconv"

	"github.com/emiliopalmerini/quintaedizione.online/internal/application/filters"
	domainFilters "github.com/emiliopalmerini/quintaedizione.online/internal/domain/filters"
	"go.mongodb.org/mongo-driver/bson"
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

func (s *FilterService) ParseFilters(collection domainFilters.CollectionType, queryParams map[string]string) (*domainFilters.FilterSet, error) {
	if !collection.IsValid() {
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

	if !filterSet.Collection.IsValid() {
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

func (s *FilterService) BuildMongoFilter(filterSet *domainFilters.FilterSet) (bson.M, error) {
	if filterSet == nil {
		return bson.M{}, nil
	}

	if err := s.ValidateFilterSet(filterSet); err != nil {
		return nil, fmt.Errorf("invalid filter set: %w", err)
	}

	return s.mongoBuilder.BuildFilter(filterSet)
}

func (s *FilterService) GetAvailableFilters(collection domainFilters.CollectionType) ([]domainFilters.FilterDefinition, error) {
	if !collection.IsValid() {
		return nil, fmt.Errorf("invalid collection: %s", collection)
	}

	return s.registry.GetFiltersForCollection(collection)
}

func (s *FilterService) BuildSearchFilter(collection domainFilters.CollectionType, searchTerm string) bson.M {
	return s.mongoBuilder.BuildSearchFilter(collection, searchTerm)
}

func (s *FilterService) CombineFilters(fieldFilter, searchFilter bson.M) bson.M {
	var conditions []bson.M

	if len(fieldFilter) > 0 {
		conditions = append(conditions, fieldFilter)
	}

	if len(searchFilter) > 0 {
		conditions = append(conditions, searchFilter)
	}

	if len(conditions) == 0 {
		return bson.M{}
	} else if len(conditions) == 1 {
		return conditions[0]
	} else {
		return bson.M{"$and": conditions}
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
