package parsers

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain"
)

// KeywordGenerator generates keywords.json from embedded JSON data files.
type KeywordGenerator struct {
	fsys       fs.FS
	outputPath string
}

// NewKeywordGenerator creates a new keyword generator that reads from the
// given filesystem (typically the embedded JSON files).
func NewKeywordGenerator(fsys fs.FS, outputPath string) *KeywordGenerator {
	return &KeywordGenerator{
		fsys:       fsys,
		outputPath: outputPath,
	}
}

type keywordSource struct {
	jsonFile   string
	keywordKey string
	urlPrefix  string
	nameField  string
}

// Generate creates keywords.json from the embedded JSON data.
func (kg *KeywordGenerator) Generate() error {
	sources := []keywordSource{
		{"spells.json", "spells", "/incantesimi/", "name"},
		{"monsters.json", "monsters", "/mostri/", "name"},
		{"equipment.json", "equipment", "/equipaggiamenti/", "name"},
		{"magic_items.json", "magic_items", "/oggetti-magici/", "name"},
		{"classes.json", "classes", "/classi/", "name"},
		{"backgrounds.json", "backgrounds", "/backgrounds/", "name"},
		{"feats.json", "feats", "/talenti/", "name"},
	}

	// Read existing keywords to preserve game mechanics categories
	existingKeywords, err := kg.loadExistingKeywords()
	if err != nil {
		return fmt.Errorf("failed to load existing keywords: %w", err)
	}

	keywords := kg.preserveGameMechanics(existingKeywords)

	for _, src := range sources {
		items, err := kg.extractFromJSON(src)
		if err != nil {
			fmt.Printf("Warning: Failed to extract keywords from %s: %v\n", src.jsonFile, err)
			continue
		}
		if len(items) > 0 {
			keywords[src.keywordKey] = items
			fmt.Printf("  - Extracted %d keywords from %s\n", len(items), src.jsonFile)
		}
	}

	return kg.writeKeywordsFile(keywords)
}

func (kg *KeywordGenerator) extractFromJSON(src keywordSource) (map[string]string, error) {
	data, err := fs.ReadFile(kg.fsys, src.jsonFile)
	if err != nil {
		return nil, err
	}

	var items []map[string]any
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	keywords := make(map[string]string)
	for _, item := range items {
		name, _ := item[src.nameField].(string)
		if name == "" {
			continue
		}
		name = strings.TrimRight(strings.TrimSpace(name), ".*")
		if name == "" {
			continue
		}

		slug, err := domain.NewSlug(name)
		if err != nil {
			continue
		}
		keywords[name] = src.urlPrefix + string(slug)
	}

	return keywords, nil
}

func (kg *KeywordGenerator) loadExistingKeywords() (map[string]any, error) {
	keywords := make(map[string]any)

	data, err := os.ReadFile(kg.outputPath)
	if err != nil {
		if os.IsNotExist(err) {
			return keywords, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, &keywords); err != nil {
		return nil, err
	}

	return keywords, nil
}

func (kg *KeywordGenerator) preserveGameMechanics(existing map[string]any) map[string]any {
	preserved := make(map[string]any)

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

func (kg *KeywordGenerator) writeKeywordsFile(keywords map[string]any) error {
	data, err := json.MarshalIndent(keywords, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal keywords: %w", err)
	}

	if err := os.WriteFile(kg.outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write keywords file: %w", err)
	}

	return nil
}
