package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// CollectionConfig represents configuration for a single collection
type CollectionConfig struct {
	Title         string   `yaml:"title"`
	DisplayFields []string `yaml:"display_fields"`
	SearchFields  []string `yaml:"search_fields"`
}

// CollectionsConfig represents the entire collections configuration
type CollectionsConfig struct {
	Collections map[string]CollectionConfig `yaml:"collections"`
}

// CollectionMetadata provides metadata and configuration for collections
type CollectionMetadata interface {
	GetTitle(collection string) string
	GetDisplayFields(collection string) []string
	GetSearchFields(collection string) []string
	GetAllCollections() map[string]CollectionConfig
	IsValidCollection(collection string) bool
}

type collectionMetadata struct {
	config *CollectionsConfig
}

// NewCollectionMetadata creates a new collection metadata provider
func NewCollectionMetadata() (CollectionMetadata, error) {
	config, err := LoadCollectionsConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load collections config: %w", err)
	}

	return &collectionMetadata{
		config: config,
	}, nil
}

// GetTitle returns the display title for a collection
func (cm *collectionMetadata) GetTitle(collection string) string {
	if config, exists := cm.config.Collections[collection]; exists {
		return config.Title
	}
	// Fallback to collection name if not found
	return collection
}

// GetDisplayFields returns the display fields for a collection
func (cm *collectionMetadata) GetDisplayFields(collection string) []string {
	if config, exists := cm.config.Collections[collection]; exists {
		return config.DisplayFields
	}
	return []string{}
}

// GetSearchFields returns the search fields for a collection
func (cm *collectionMetadata) GetSearchFields(collection string) []string {
	if config, exists := cm.config.Collections[collection]; exists {
		return config.SearchFields
	}
	return []string{}
}

// GetAllCollections returns all configured collections
func (cm *collectionMetadata) GetAllCollections() map[string]CollectionConfig {
	return cm.config.Collections
}

// IsValidCollection checks if a collection is configured
func (cm *collectionMetadata) IsValidCollection(collection string) bool {
	_, exists := cm.config.Collections[collection]
	return exists
}

// LoadCollectionsConfig loads the collections configuration from YAML file
func LoadCollectionsConfig() (*CollectionsConfig, error) {
	// Try to find the config file
	configPaths := []string{
		"configs/collections.yaml",
		"./configs/collections.yaml",
		"../configs/collections.yaml",
		"../../configs/collections.yaml",
	}

	var configData []byte

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			if data, readErr := os.ReadFile(path); readErr == nil {
				configData = data
				break
			}
		}
	}

	if configData == nil {
		return nil, fmt.Errorf("collections.yaml not found in any of the expected paths: %v", configPaths)
	}

	var config CollectionsConfig
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse collections.yaml: %w", err)
	}

	return &config, nil
}

// GetConfigPath returns the absolute path to the config directory
func GetConfigPath() (string, error) {
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Look for configs directory
	configPaths := []string{
		filepath.Join(wd, "configs"),
		filepath.Join(wd, "..", "configs"),
		filepath.Join(wd, "..", "..", "configs"),
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("configs directory not found")
}