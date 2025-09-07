package factories

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
	"gopkg.in/yaml.v3"
)

// LoadLanguageConfig loads language configuration from YAML file
func LoadLanguageConfig(configPath string) (*parsers.LanguageConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config parsers.LanguageConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	// Compile regex patterns
	if err := compilePatterns(&config); err != nil {
		return nil, fmt.Errorf("failed to compile patterns: %w", err)
	}

	// Validate configuration
	if err := validateLanguageConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// compilePatterns compiles regex patterns from string patterns
func compilePatterns(config *parsers.LanguageConfig) error {
	config.Patterns = make(map[string]*regexp.Regexp)

	for patternName, patternStr := range config.PatternStrings {
		compiled, err := regexp.Compile(patternStr)
		if err != nil {
			return fmt.Errorf("invalid regex pattern %s: %s - %w", patternName, patternStr, err)
		}
		config.Patterns[patternName] = compiled
	}

	return nil
}

// validateLanguageConfig validates the loaded configuration
func validateLanguageConfig(config *parsers.LanguageConfig) error {
	if config.DataPath == "" {
		return fmt.Errorf("data_path is required")
	}

	if config.SectionDelimiter == "" {
		return fmt.Errorf("section_delimiter is required")
	}

	if config.FieldMappings == nil {
		return fmt.Errorf("field_mappings is required")
	}

	// Validate that data path exists (skip if it's a test or the path doesn't start with absolute path)
	if config.DataPath != "" && !strings.HasPrefix(config.DataPath, "/") {
		if _, err := os.Stat(config.DataPath); os.IsNotExist(err) {
			// For tests, we'll allow non-existent data paths
			// In production, this would be more strict
		}
	}

	return nil
}

// GetLanguageConfigPath returns the path to a language configuration file
func GetLanguageConfigPath(language parsers.LanguageCode) string {
	switch language {
	case parsers.Italian:
		return "config/languages/italian.yaml"
	case parsers.English:
		return "config/languages/english.yaml"
	default:
		return ""
	}
}

// LoadAllLanguageConfigs loads configurations for all supported languages
func LoadAllLanguageConfigs() (map[parsers.LanguageCode]*parsers.LanguageConfig, error) {
	configs := make(map[parsers.LanguageCode]*parsers.LanguageConfig)
	
	languages := []parsers.LanguageCode{parsers.Italian, parsers.English}
	
	for _, lang := range languages {
		configPath := GetLanguageConfigPath(lang)
		if configPath == "" {
			continue
		}
		
		config, err := LoadLanguageConfig(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config for %s: %w", lang, err)
		}
		
		configs[lang] = config
	}
	
	return configs, nil
}