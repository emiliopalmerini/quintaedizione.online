package filters

import (
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/collections"
)

// FilterDefinition describes a faceted filter exposed for a set of collections.
// All filters share the same semantics: case-insensitive membership match against
// a (possibly nested or slice-valued) document field, with comma-separated values
// in the query string treated as OR.
type FilterDefinition struct {
	Name        string
	FieldPath   string
	Collections []collections.CollectionName
	EnumValues  []string
	Description string
}

type FilterValue struct {
	Definition FilterDefinition
	Value      string
}

type FilterSet struct {
	Collection collections.CollectionName
	Filters    []FilterValue
}

type FilterRepository interface {
	GetFiltersForCollection(collection collections.CollectionName) ([]FilterDefinition, error)
	GetFilterByName(name string) (FilterDefinition, bool)
	GetAllFilters() ([]FilterDefinition, error)
}

// DocumentPredicate tests whether a document matches a filter condition.
type DocumentPredicate = func(map[string]any) bool

type FilterService interface {
	ParseFilters(collection collections.CollectionName, queryParams map[string]string) (*FilterSet, error)
	BuildFilter(filterSet *FilterSet) (DocumentPredicate, error)
	GetAvailableFilters(collection collections.CollectionName) ([]FilterDefinition, error)

	BuildSearchPredicate(collection collections.CollectionName, searchTerm string) DocumentPredicate
	CombinePredicates(predicates ...DocumentPredicate) DocumentPredicate
}

func NewFilterSet(collection collections.CollectionName) *FilterSet {
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

func (fd FilterDefinition) IsApplicableToCollection(collection collections.CollectionName) bool {
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
