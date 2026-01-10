package parsers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain"
)

// KeywordGenerator generates keywords.json from markdown files
type KeywordGenerator struct {
	dataBasePath string
	outputPath   string
}

// NewKeywordGenerator creates a new keyword generator
func NewKeywordGenerator(dataBasePath, outputPath string) *KeywordGenerator {
	return &KeywordGenerator{
		dataBasePath: dataBasePath,
		outputPath:   outputPath,
	}
}

// collectionKeywordConfig defines how to generate keywords for a collection
type collectionKeywordConfig struct {
	workItem     WorkItem
	keywordKey   string
	urlPrefix    string
	titleCleaner func(string) string
}

// Generate creates keywords.json from markdown files
func (kg *KeywordGenerator) Generate() error {
	// Define collections to process
	configs := []collectionKeywordConfig{
		{WorkItem{"ita/lists/incantesimi.md", "incantesimi", Italian}, "spells", "/incantesimi/", cleanSpellTitle},
		{WorkItem{"ita/lists/mostri.md", "mostri", Italian}, "monsters", "/mostri/", cleanGenericTitle},
		{WorkItem{"ita/lists/equipaggiamenti.md", "equipaggiamenti", Italian}, "equipment", "/equipaggiamenti/", cleanGenericTitle},
		{WorkItem{"ita/lists/oggetti_magici.md", "oggetti_magici", Italian}, "magic_items", "/oggetti-magici/", cleanGenericTitle},
		{WorkItem{"ita/lists/armi.md", "armi", Italian}, "weapons", "/armi/", cleanGenericTitle},
		{WorkItem{"ita/lists/backgrounds.md", "backgrounds", Italian}, "backgrounds", "/backgrounds/", cleanGenericTitle},
		{WorkItem{"ita/lists/armature.md", "armature", Italian}, "armor", "/armature/", cleanGenericTitle},
		{WorkItem{"ita/lists/classi.md", "classi", Italian}, "classes", "/classi/", cleanGenericTitle},
	}

	// Read existing keywords to preserve game mechanics categories
	existingKeywords, err := kg.loadExistingKeywords()
	if err != nil {
		return fmt.Errorf("failed to load existing keywords: %w", err)
	}

	// Start with preserved categories
	keywords := kg.preserveGameMechanics(existingKeywords)

	// Generate keywords for each collection
	for _, config := range configs {
		collectionKeywords, err := kg.extractKeywordsFromWorkItem(config)
		if err != nil {
			// Log warning but continue with other collections
			fmt.Printf("Warning: Failed to extract keywords from %s: %v\n", config.workItem.Filename, err)
			continue
		}

		if len(collectionKeywords) > 0 {
			keywords[config.keywordKey] = collectionKeywords
			fmt.Printf("  - Extracted %d keywords from %s\n", len(collectionKeywords), config.workItem.Collection)
		}
	}

	// Write to file
	return kg.writeKeywordsFile(keywords)
}

// loadExistingKeywords reads the current keywords.json file
func (kg *KeywordGenerator) loadExistingKeywords() (map[string]interface{}, error) {
	keywords := make(map[string]interface{})

	data, err := os.ReadFile(kg.outputPath)
	if err != nil {
		if os.IsNotExist(err) {
			return keywords, nil // File doesn't exist yet, return empty map
		}
		return nil, err
	}

	if err := json.Unmarshal(data, &keywords); err != nil {
		return nil, err
	}

	return keywords, nil
}

// preserveGameMechanics extracts existing game mechanics categories
func (kg *KeywordGenerator) preserveGameMechanics(existing map[string]interface{}) map[string]interface{} {
	preserved := make(map[string]interface{})

	// Preserve existing game mechanics categories
	categoriesToPreserve := []string{
		"damage_types",
		"conditions",
		"core_mechanics",
		"creature_types",
	}

	for _, category := range categoriesToPreserve {
		if value, ok := existing[category]; ok {
			preserved[category] = value
		}
	}

	return preserved
}

// extractKeywordsFromWorkItem extracts keywords from a single markdown file
func (kg *KeywordGenerator) extractKeywordsFromWorkItem(config collectionKeywordConfig) (map[string]string, error) {
	filePath := filepath.Join(kg.dataBasePath, config.workItem.Filename)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	keywords := make(map[string]string)
	scanner := bufio.NewScanner(file)

	// Regex to match level 2 headers (## Title)
	headerRegex := regexp.MustCompile(`^##\s+(.+)$`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check if line is a level 2 header
		matches := headerRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		// Extract and clean title
		title := config.titleCleaner(matches[1])
		if title == "" {
			continue
		}

		// Generate URL slug
		slug, err := domain.NewSlug(title)
		if err != nil {
			fmt.Printf("Warning: Failed to generate slug for '%s': %v\n", title, err)
			continue
		}

		// Add to keywords map
		url := config.urlPrefix + string(slug)
		keywords[title] = url
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return keywords, nil
}

// writeKeywordsFile writes the keywords map to JSON
func (kg *KeywordGenerator) writeKeywordsFile(keywords map[string]interface{}) error {
	data, err := json.MarshalIndent(keywords, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal keywords: %w", err)
	}

	if err := os.WriteFile(kg.outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write keywords file: %w", err)
	}

	return nil
}

// cleanSpellTitle removes trailing periods and asterisks from spell names
func cleanSpellTitle(title string) string {
	title = strings.TrimSpace(title)
	// Remove trailing periods and asterisks (spell-specific formatting)
	title = strings.TrimRight(title, ".*")
	return strings.TrimSpace(title)
}

// cleanGenericTitle trims whitespace only
func cleanGenericTitle(title string) string {
	return strings.TrimSpace(title)
}
