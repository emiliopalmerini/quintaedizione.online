package filters

import (
	"fmt"
	"os"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain/filters"
	"gopkg.in/yaml.v3"
)

// ConfigFilterDefinition represents a filter definition in the YAML config
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

// FilterConfig represents the structure of the filters.yaml file
type FilterConfig struct {
	Filters []ConfigFilterDefinition `yaml:"filters"`
}

// YAMLFilterRegistry implements FilterRepository using YAML configuration
type YAMLFilterRegistry struct {
	filters    []filters.FilterDefinition
	configPath string
}

// NewYAMLFilterRegistry creates a new YAML-based filter registry
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

// GetFiltersForCollection returns all filters applicable to a collection
func (r *YAMLFilterRegistry) GetFiltersForCollection(collection filters.CollectionType) ([]filters.FilterDefinition, error) {
	result := make([]filters.FilterDefinition, 0)

	for _, filter := range r.filters {
		if filter.IsApplicableToCollection(collection) {
			result = append(result, filter)
		}
	}

	return result, nil
}

// GetFilterByName returns a filter definition by name
func (r *YAMLFilterRegistry) GetFilterByName(name string) (filters.FilterDefinition, bool) {
	for _, filter := range r.filters {
		if filter.Name == name {
			return filter, true
		}
	}
	return filters.FilterDefinition{}, false
}

// GetAllFilters returns all available filter definitions
func (r *YAMLFilterRegistry) GetAllFilters() ([]filters.FilterDefinition, error) {
	return r.filters, nil
}

// loadConfig loads filter definitions from the YAML configuration file
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

// convertConfigToFilter converts a config filter to a domain filter definition
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

// parseDataType converts string data type to enum
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

// parseOperator converts string operator to enum
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

// InMemoryFilterRegistry provides an in-memory implementation for testing
type InMemoryFilterRegistry struct {
	filters map[string]filters.FilterDefinition
}

// NewInMemoryFilterRegistry creates a new in-memory filter registry
func NewInMemoryFilterRegistry() *InMemoryFilterRegistry {
	return &InMemoryFilterRegistry{
		filters: make(map[string]filters.FilterDefinition),
	}
}

// AddFilter adds a filter definition to the registry
func (r *InMemoryFilterRegistry) AddFilter(filter filters.FilterDefinition) {
	r.filters[filter.Name] = filter
}

// GetFiltersForCollection returns all filters applicable to a collection
func (r *InMemoryFilterRegistry) GetFiltersForCollection(collection filters.CollectionType) ([]filters.FilterDefinition, error) {
	result := make([]filters.FilterDefinition, 0)

	for _, filter := range r.filters {
		if filter.IsApplicableToCollection(collection) {
			result = append(result, filter)
		}
	}

	return result, nil
}

// GetFilterByName returns a filter definition by name
func (r *InMemoryFilterRegistry) GetFilterByName(name string) (filters.FilterDefinition, bool) {
	filter, exists := r.filters[name]
	return filter, exists
}

// GetAllFilters returns all available filter definitions
func (r *InMemoryFilterRegistry) GetAllFilters() ([]filters.FilterDefinition, error) {
	result := make([]filters.FilterDefinition, 0, len(r.filters))
	for _, filter := range r.filters {
		result = append(result, filter)
	}
	return result, nil
}