package filters

import (
	"fmt"
	"os"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/filters"
	"gopkg.in/yaml.v3"
)

type ConfigFilterDefinition struct {
	Name        string   `yaml:"name"`
	FieldPath   string   `yaml:"field_path"`
	DataType    string   `yaml:"data_type"`
	Operator    string   `yaml:"operator"`
	Collections []string `yaml:"collections"`
	EnumValues  []string `yaml:"enum_values,omitempty"`
	Required    bool     `yaml:"required,omitempty"`
	Description string   `yaml:"description,omitempty"`
}

type FilterConfig struct {
	Filters []ConfigFilterDefinition `yaml:"filters"`
}

type YAMLFilterRegistry struct {
	filters    []filters.FilterDefinition
	configPath string
}

func NewYAMLFilterRegistry(configPath string) (*YAMLFilterRegistry, error) {
	registry := &YAMLFilterRegistry{
		configPath: configPath,
		filters:    make([]filters.FilterDefinition, 0),
	}

	if err := registry.loadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load filter config: %w", err)
	}

	return registry, nil
}

func (r *YAMLFilterRegistry) GetFiltersForCollection(collection filters.CollectionType) ([]filters.FilterDefinition, error) {
	result := make([]filters.FilterDefinition, 0)

	for _, filter := range r.filters {
		if filter.IsApplicableToCollection(collection) {
			result = append(result, filter)
		}
	}

	return result, nil
}

func (r *YAMLFilterRegistry) GetFilterByName(name string) (filters.FilterDefinition, bool) {
	for _, filter := range r.filters {
		if filter.Name == name {
			return filter, true
		}
	}
	return filters.FilterDefinition{}, false
}

func (r *YAMLFilterRegistry) GetAllFilters() ([]filters.FilterDefinition, error) {
	return r.filters, nil
}

func (r *YAMLFilterRegistry) loadConfig() error {
	data, err := os.ReadFile(r.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", r.configPath, err)
	}

	var config FilterConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	r.filters = make([]filters.FilterDefinition, 0, len(config.Filters))

	for _, configFilter := range config.Filters {
		filter, err := r.convertConfigToFilter(configFilter)
		if err != nil {
			return fmt.Errorf("failed to convert filter %s: %w", configFilter.Name, err)
		}
		r.filters = append(r.filters, filter)
	}

	return nil
}

func (r *YAMLFilterRegistry) convertConfigToFilter(config ConfigFilterDefinition) (filters.FilterDefinition, error) {
	dataType, err := parseDataType(config.DataType)
	if err != nil {
		return filters.FilterDefinition{}, fmt.Errorf("invalid data type %s: %w", config.DataType, err)
	}

	operator, err := parseOperator(config.Operator)
	if err != nil {
		return filters.FilterDefinition{}, fmt.Errorf("invalid operator %s: %w", config.Operator, err)
	}

	collections := make([]filters.CollectionType, 0, len(config.Collections))
	for _, collName := range config.Collections {
		collection := filters.CollectionType(collName)
		if !collection.IsValid() {
			return filters.FilterDefinition{}, fmt.Errorf("invalid collection %s", collName)
		}
		collections = append(collections, collection)
	}

	return filters.FilterDefinition{
		Name:        config.Name,
		FieldPath:   config.FieldPath,
		DataType:    dataType,
		Operator:    operator,
		Collections: collections,
		EnumValues:  config.EnumValues,
		Required:    config.Required,
		Description: config.Description,
	}, nil
}

func parseDataType(dataType string) (filters.FilterDataType, error) {
	switch dataType {
	case "string":
		return filters.StringFilter, nil
	case "number":
		return filters.NumberFilter, nil
	case "boolean":
		return filters.BooleanFilter, nil
	case "enum":
		return filters.EnumFilter, nil
	default:
		return 0, fmt.Errorf("unknown data type: %s", dataType)
	}
}

func parseOperator(operator string) (filters.FilterOperator, error) {
	switch operator {
	case "exact":
		return filters.ExactMatch, nil
	case "regex":
		return filters.RegexMatch, nil
	case "range":
		return filters.RangeMatch, nil
	case "in":
		return filters.InMatch, nil
	default:
		return 0, fmt.Errorf("unknown operator: %s", operator)
	}
}

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

func (r *InMemoryFilterRegistry) GetFiltersForCollection(collection filters.CollectionType) ([]filters.FilterDefinition, error) {
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
