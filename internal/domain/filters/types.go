package filters

import "go.mongodb.org/mongo-driver/bson"

// FilterDataType represents the data type of a filter value
type FilterDataType int

const (
	StringFilter FilterDataType = iota
	NumberFilter
	BooleanFilter
	EnumFilter
)

// FilterOperator represents how a filter should be applied
type FilterOperator int

const (
	ExactMatch FilterOperator = iota
	RegexMatch
	RangeMatch
	InMatch
)

// FilterDefinition defines metadata for a single filter
type FilterDefinition struct {
	Name        string
	FieldPath   string
	DataType    FilterDataType
	Operator    FilterOperator
	Collections []CollectionType
	EnumValues  []string // For EnumFilter type
	Required    bool
	Description string
}

// FilterValue represents a single filter value with its definition
type FilterValue struct {
	Definition FilterDefinition
	Value      string
	RawValue   any // Original typed value
}

// FilterSet represents a collection of active filters
type FilterSet struct {
	Collection CollectionType
	Filters    []FilterValue
}

// FilterRepository defines the interface for filter data access
type FilterRepository interface {
	GetFiltersForCollection(collection CollectionType) ([]FilterDefinition, error)
	GetFilterByName(name string) (FilterDefinition, bool)
	GetAllFilters() ([]FilterDefinition, error)
}

// FilterService defines the interface for filter business logic
type FilterService interface {
	ParseFilters(collection CollectionType, queryParams map[string]string) (*FilterSet, error)
	ValidateFilterSet(filterSet *FilterSet) error
	BuildMongoFilter(filterSet *FilterSet) (bson.M, error)
	GetAvailableFilters(collection CollectionType) ([]FilterDefinition, error)
	// BuildSearchFilter creates a MongoDB $text search filter from a search term.
	// Uses MongoDB text indexes for efficient full-text search with support for:
	// - Word boundaries (words match from start, not fuzzy)
	// - Stemming and stop words (language-aware)
	// - Phrase queries: "exact phrase" (with quotes)
	// - Negation: -word (exclude from results)
	// - Multiple words: word1 word2 (implicit AND)
	BuildSearchFilter(collection CollectionType, searchTerm string) bson.M
	CombineFilters(fieldFilter, searchFilter bson.M) bson.M
}

// NewFilterSet creates a new filter set for a collection
func NewFilterSet(collection CollectionType) *FilterSet {
	return &FilterSet{
		Collection: collection,
		Filters:    make([]FilterValue, 0),
	}
}

// AddFilter adds a filter value to the set
func (fs *FilterSet) AddFilter(filterValue FilterValue) {
	fs.Filters = append(fs.Filters, filterValue)
}

// HasFilters returns true if the filter set contains any filters
func (fs *FilterSet) HasFilters() bool {
	return len(fs.Filters) > 0
}

// GetFilter returns a filter by name if it exists
func (fs *FilterSet) GetFilter(name string) (FilterValue, bool) {
	for _, filter := range fs.Filters {
		if filter.Definition.Name == name {
			return filter, true
		}
	}
	return FilterValue{}, false
}

// IsApplicableToCollection checks if a filter definition applies to a collection
func (fd FilterDefinition) IsApplicableToCollection(collection CollectionType) bool {
	if len(fd.Collections) == 0 {
		return true // No restrictions means applies to all
	}

	for _, c := range fd.Collections {
		if c == collection {
			return true
		}
	}
	return false
}

// ValidateValue validates a value against the filter definition
func (fd FilterDefinition) ValidateValue(value string) error {
	if fd.Required && value == "" {
		return NewValidationError(fd.Name, "value is required")
	}

	if value == "" {
		return nil // Empty values are valid for non-required fields
	}

	switch fd.DataType {
	case EnumFilter:
		if len(fd.EnumValues) > 0 {
			for _, enumValue := range fd.EnumValues {
				if enumValue == value {
					return nil
				}
			}
			return NewValidationError(fd.Name, "invalid enum value")
		}
	}

	return nil
}
