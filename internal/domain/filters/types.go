package filters

import "go.mongodb.org/mongo-driver/bson"

type FilterDataType int

const (
	StringFilter FilterDataType = iota
	NumberFilter
	BooleanFilter
	EnumFilter
)

type FilterOperator int

const (
	ExactMatch FilterOperator = iota
	RegexMatch
	RangeMatch
	InMatch
)

type FilterDefinition struct {
	Name        string
	FieldPath   string
	DataType    FilterDataType
	Operator    FilterOperator
	Collections []CollectionType
	EnumValues  []string
	Required    bool
	Description string
}

type FilterValue struct {
	Definition FilterDefinition
	Value      string
	RawValue   any
}

type FilterSet struct {
	Collection CollectionType
	Filters    []FilterValue
}

type FilterRepository interface {
	GetFiltersForCollection(collection CollectionType) ([]FilterDefinition, error)
	GetFilterByName(name string) (FilterDefinition, bool)
	GetAllFilters() ([]FilterDefinition, error)
}

type FilterService interface {
	ParseFilters(collection CollectionType, queryParams map[string]string) (*FilterSet, error)
	ValidateFilterSet(filterSet *FilterSet) error
	BuildMongoFilter(filterSet *FilterSet) (bson.M, error)
	GetAvailableFilters(collection CollectionType) ([]FilterDefinition, error)

	BuildSearchFilter(collection CollectionType, searchTerm string) bson.M
	CombineFilters(fieldFilter, searchFilter bson.M) bson.M
}

func NewFilterSet(collection CollectionType) *FilterSet {
	return &FilterSet{
		Collection: collection,
		Filters:    make([]FilterValue, 0),
	}
}

func (fs *FilterSet) AddFilter(filterValue FilterValue) {
	fs.Filters = append(fs.Filters, filterValue)
}

func (fs *FilterSet) HasFilters() bool {
	return len(fs.Filters) > 0
}

func (fs *FilterSet) GetFilter(name string) (FilterValue, bool) {
	for _, filter := range fs.Filters {
		if filter.Definition.Name == name {
			return filter, true
		}
	}
	return FilterValue{}, false
}

func (fd FilterDefinition) IsApplicableToCollection(collection CollectionType) bool {
	if len(fd.Collections) == 0 {
		return true
	}

	for _, c := range fd.Collections {
		if c == collection {
			return true
		}
	}
	return false
}

func (fd FilterDefinition) ValidateValue(value string) error {
	if fd.Required && value == "" {
		return NewValidationError(fd.Name, "value is required")
	}

	if value == "" {
		return nil
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
