package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type CollectionConfig struct {
	Title         string   `yaml:"title"`
	DisplayFields []string `yaml:"display_fields"`
	SearchFields  []string `yaml:"search_fields"`
}

type CollectionsConfig struct {
	Collections map[string]CollectionConfig `yaml:"collections"`
}

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

func NewCollectionMetadata() (CollectionMetadata, error) {
	config, err := LoadCollectionsConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load collections config: %w", err)
	}

	return &collectionMetadata{
		config: config,
	}, nil
}

func (cm *collectionMetadata) GetTitle(collection string) string {
	if config, exists := cm.config.Collections[collection]; exists {
		return config.Title
	}

	return collection
}

func (cm *collectionMetadata) GetDisplayFields(collection string) []string {
	if config, exists := cm.config.Collections[collection]; exists {
		return config.DisplayFields
	}
	return []string{}
}

func (cm *collectionMetadata) GetSearchFields(collection string) []string {
	if config, exists := cm.config.Collections[collection]; exists {
		return config.SearchFields
	}
	return []string{}
}

func (cm *collectionMetadata) GetAllCollections() map[string]CollectionConfig {
	return cm.config.Collections
}

func (cm *collectionMetadata) IsValidCollection(collection string) bool {
	_, exists := cm.config.Collections[collection]
	return exists
}

func LoadCollectionsConfig() (*CollectionsConfig, error) {

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

func GetConfigPath() (string, error) {

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

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
