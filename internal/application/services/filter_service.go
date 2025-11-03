package services

import (
	"fmt"
	"strconv"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/filters"
	domainFilters "github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain/filters"
	"go.mongodb.org/mongo-driver/bson"
)

// FilterService implements the domain FilterService interface
type FilterService struct {
	registry     domainFilters.FilterRepository
	mongoBuilder *filters.MongoFilterBuilder
}

// NewFilterService creates a new filter service
func NewFilterService(registry domainFilters.FilterRepository) *FilterService {
	return &FilterService{
		registry:     registry,
		mongoBuilder: filters.NewMongoFilterBuilder(),
	}
}

// ParseFilters parses query parameters into a validated filter set
func (s *FilterService) ParseFilters(collection domainFilters.CollectionType, queryParams map[string]string) (*domainFilters.FilterSet, error) {
	if !collection.IsValid() {
		return nil, fmt.Errorf("invalid collection: %s", collection)
	}

	filterSet := domainFilters.NewFilterSet(collection)

	// Get available filters for this collection
	availableFilters, err := s.registry.GetFiltersForCollection(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to get available filters: %w", err)
	}

	// Create a map for quick lookup
	filterMap := make(map[string]domainFilters.FilterDefinition)
	for _, filter := range availableFilters {
		filterMap[filter.Name] = filter
	}

	// Parse each query parameter
	for paramName, paramValue := range queryParams {
		if paramValue == "" {
			continue // Skip empty values
		}

		// Check if this parameter corresponds to a known filter
		filterDef, exists := filterMap[paramName]
		if !exists {
			// Try to find filter by name in registry (for filters that apply to all collections)
			filterDef, exists = s.registry.GetFilterByName(paramName)
			if !exists {
				continue // Skip unknown parameters
			}

			// Check if the found filter applies to this collection
			if !filterDef.IsApplicableToCollection(collection) {
				return nil, domainFilters.NewUnsupportedFilterError(paramName, collection)
			}
		}

		// Validate the value
		if err := filterDef.ValidateValue(paramValue); err != nil {
			return nil, fmt.Errorf("validation failed for filter %s: %w", paramName, err)
		}

		// Convert value based on data type
		rawValue, err := s.convertValue(paramValue, filterDef.DataType)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value for filter %s: %w", paramName, err)
		}

		// Create filter value and add to set
		filterValue := domainFilters.FilterValue{
			Definition: filterDef,
			Value:      paramValue,
			RawValue:   rawValue,
		}

		filterSet.AddFilter(filterValue)
	}

	return filterSet, nil
}

// ValidateFilterSet validates a complete filter set
func (s *FilterService) ValidateFilterSet(filterSet *domainFilters.FilterSet) error {
	if filterSet == nil {
		return fmt.Errorf("filter set cannot be nil")
	}

	if !filterSet.Collection.IsValid() {
		return fmt.Errorf("invalid collection: %s", filterSet.Collection)
	}

	// Validate each filter in the set
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

// BuildMongoFilter builds a MongoDB filter from a filter set
func (s *FilterService) BuildMongoFilter(filterSet *domainFilters.FilterSet) (bson.M, error) {
	if filterSet == nil {
		return bson.M{}, nil
	}

	if err := s.ValidateFilterSet(filterSet); err != nil {
		return nil, fmt.Errorf("invalid filter set: %w", err)
	}

	return s.mongoBuilder.BuildFilter(filterSet)
}

// GetAvailableFilters returns all filters available for a collection
func (s *FilterService) GetAvailableFilters(collection domainFilters.CollectionType) ([]domainFilters.FilterDefinition, error) {
	if !collection.IsValid() {
		return nil, fmt.Errorf("invalid collection: %s", collection)
	}

	return s.registry.GetFiltersForCollection(collection)
}

// BuildSearchFilter builds a text search filter (separate from field filters)
func (s *FilterService) BuildSearchFilter(collection domainFilters.CollectionType, searchTerm string) bson.M {
	return s.mongoBuilder.BuildSearchFilter(collection, searchTerm)
}

// CombineFilters combines field filters and search filters into a single MongoDB query
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

// convertValue converts a string value to the appropriate type
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
