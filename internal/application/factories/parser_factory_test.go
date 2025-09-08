package factories

import (
	"testing"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
)

func TestParserFactory_LoadLanguageConfig(t *testing.T) {
	factory := NewParserFactory(parsers.NewRegistry(), &parsers.NoOpLogger{})

	// Test loading Italian config
	err := factory.LoadLanguageConfig(parsers.Italian, "../../../config/languages/italian.yaml")
	if err != nil {
		t.Fatalf("Failed to load Italian config: %v", err)
	}

	// Test loading English config
	err = factory.LoadLanguageConfig(parsers.English, "../../../config/languages/english.yaml")
	if err != nil {
		t.Fatalf("Failed to load English config: %v", err)
	}

	// Verify configs are loaded
	if len(factory.languageConfigs) != 2 {
		t.Errorf("Expected 2 language configs, got %d", len(factory.languageConfigs))
	}

	// Verify Italian config
	italianConfig := factory.languageConfigs[parsers.Italian]
	if italianConfig == nil {
		t.Fatal("Italian config not found")
	}
	if italianConfig.DataPath != "data/ita" {
		t.Errorf("Italian config data path: expected 'data/ita', got %s", italianConfig.DataPath)
	}

	// Verify English config
	englishConfig := factory.languageConfigs[parsers.English]
	if englishConfig == nil {
		t.Fatal("English config not found")
	}
	if englishConfig.DataPath != "data/eng" {
		t.Errorf("English config data path: expected 'data/eng', got %s", englishConfig.DataPath)
	}
}

func TestParserFactory_CreateParser(t *testing.T) {
	factory := NewParserFactory(parsers.NewRegistry(), &parsers.NoOpLogger{})

	// Load language configs
	err := factory.LoadAllLanguageConfigs()
	if err != nil {
		t.Fatalf("Failed to load all language configs: %v", err)
	}

	testCases := []struct {
		contentType parsers.ContentType
		language    parsers.LanguageCode
		expectError bool
	}{
		{parsers.ContentTypeSpells, parsers.Italian, false},
		{parsers.ContentTypeSpells, parsers.English, false},
		{parsers.ContentTypeMonsters, parsers.Italian, false},
		{parsers.ContentTypeMonsters, parsers.English, false},
		{parsers.ContentTypeClasses, parsers.Italian, false},
		{parsers.ContentTypeWeapons, parsers.Italian, false},
		{parsers.ContentTypeArmor, parsers.Italian, false},
		{"invalid_type", parsers.Italian, true},
	}

	for _, tc := range testCases {
		t.Run(string(tc.contentType)+"_"+string(tc.language), func(t *testing.T) {
			parser, err := factory.CreateParser(tc.contentType, tc.language)

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if parser == nil {
				t.Fatal("Parser is nil")
			}

			// Verify parser properties
			if parser.ContentType() != tc.contentType {
				t.Errorf("Parser content type: expected %s, got %s", tc.contentType, parser.ContentType())
			}
		})
	}
}

func TestParserFactory_RegisterParsersForLanguage(t *testing.T) {
	registry := parsers.NewRegistry()
	factory := NewParserFactory(registry, &parsers.NoOpLogger{})

	// Load language configs
	err := factory.LoadAllLanguageConfigs()
	if err != nil {
		t.Fatalf("Failed to load all language configs: %v", err)
	}

	// Register Italian parsers
	err = factory.RegisterParsersForLanguage(parsers.Italian)
	if err != nil {
		t.Fatalf("Failed to register Italian parsers: %v", err)
	}

	// Register English parsers
	err = factory.RegisterParsersForLanguage(parsers.English)
	if err != nil {
		t.Fatalf("Failed to register English parsers: %v", err)
	}

	// Verify registrations
	keys := registry.ListKeys()
	if len(keys) == 0 {
		t.Fatal("No parsers registered")
	}

	// Test specific parser retrieval
	spellsParser, err := registry.GetStrategy(parsers.ContentTypeSpells, parsers.Italian)
	if err != nil {
		t.Fatalf("Failed to get Italian spells parser: %v", err)
	}
	if spellsParser == nil {
		t.Fatal("Italian spells parser is nil")
	}

	// Test English parser
	spellsParserEn, err := registry.GetStrategy(parsers.ContentTypeSpells, parsers.English)
	if err != nil {
		t.Fatalf("Failed to get English spells parser: %v", err)
	}
	if spellsParserEn == nil {
		t.Fatal("English spells parser is nil")
	}

	// Verify they are different instances
	if spellsParser == spellsParserEn {
		t.Error("Italian and English parsers should be different instances")
	}
}

func TestLanguageConfig_FieldMappings(t *testing.T) {
	factory := NewParserFactory(parsers.NewRegistry(), &parsers.NoOpLogger{})

	// Load language configs
	err := factory.LoadAllLanguageConfigs()
	if err != nil {
		t.Fatalf("Failed to load all language configs: %v", err)
	}

	// Test Italian field mappings
	italianConfig := factory.languageConfigs[parsers.Italian]
	if italianConfig.FieldMappings["name"] != "nome" {
		t.Errorf("Italian 'name' mapping: expected 'nome', got %s", italianConfig.FieldMappings["name"])
	}
	if italianConfig.FieldMappings["level"] != "livello" {
		t.Errorf("Italian 'level' mapping: expected 'livello', got %s", italianConfig.FieldMappings["level"])
	}

	// Test English field mappings
	englishConfig := factory.languageConfigs[parsers.English]
	if englishConfig.FieldMappings["name"] != "name" {
		t.Errorf("English 'name' mapping: expected 'name', got %s", englishConfig.FieldMappings["name"])
	}
	if englishConfig.FieldMappings["level"] != "level" {
		t.Errorf("English 'level' mapping: expected 'level', got %s", englishConfig.FieldMappings["level"])
	}
}
