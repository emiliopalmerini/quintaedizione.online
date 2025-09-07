package parsers

import (
	"regexp"
	"testing"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// MockStrategy implements a simple parser for testing Template Method
type MockStrategy struct {
	*BaseParser
	parseError bool
}

func NewMockStrategy(parseError bool) *MockStrategy {
	return &MockStrategy{
		BaseParser: NewBaseParserWithLanguage(
			ContentTypeSpells,
			"Mock Parser",
			"Test parser for Template Method pattern",
			Italian,
			&LanguageConfig{
				SectionDelimiter: "##", // Explicitly set section delimiter for tests
				FieldMappings: map[string]string{
					"name": "nome",
					"level": "livello",
				},
				Patterns: map[string]*regexp.Regexp{
					"level_pattern": regexp.MustCompile(`Livello\\s+(\\d+)`),
				},
			},
		),
		parseError: parseError,
	}
}

// Mock domain object for testing
type MockEntity struct {
	Name string
	Type string
}

func (m *MockEntity) GetID() string { return "mock-id" }
func (m *MockEntity) GetContent() string { return "mock content" }
func (m *MockEntity) EntityType() string { return "mock" }

// Implement the parseSection hook method
func (m *MockStrategy) parseSection(section Section, context *ParsingContext) (domain.ParsedEntity, error) {
	if m.parseError {
		return nil, ErrMissingSectionTitle
	}
	
	return &MockEntity{
		Name: section.Title,
		Type: "Mock",
	}, nil
}

// Test BaseParser Template Method
func TestBaseParser_TemplateMethod(t *testing.T) {
	testCases := []struct {
		name        string
		content     []string
		parseError  bool
		expectError bool
		expectCount int
	}{
		{
			name: "successful parsing",
			content: []string{
				"## First Item",
				"Content for first item",
				"",
				"## Second Item",
				"Content for second item",
			},
			parseError:  false,
			expectError: false,
			expectCount: 2,
		},
		{
			name: "empty content should fail validation",
			content: []string{},
			parseError:  false,
			expectError: true,
			expectCount: 0,
		},
		{
			name: "parse section errors should be handled gracefully",
			content: []string{
				"## First Item",
				"Content for first item",
			},
			parseError:  true,
			expectError: false,
			expectCount: 0, // Errors are logged but parsing continues
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			strategy := NewMockStrategy(tc.parseError)
			context := &ParsingContext{}

			result, err := strategy.Parse(tc.content, context)

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(result) != tc.expectCount {
				t.Errorf("Expected %d entities, got %d", tc.expectCount, len(result))
			}

			// Verify entities are of correct type
			for i, entity := range result {
				mockEntity, ok := entity.(*MockEntity)
				if !ok {
					t.Errorf("Entity %d is not MockEntity type", i)
					continue
				}

				if mockEntity.Type != "Mock" {
					t.Errorf("Entity %d type: expected 'Mock', got %s", i, mockEntity.Type)
				}
			}
		})
	}
}

func TestBaseParser_LanguageAwareFieldExtraction(t *testing.T) {
	// Create parser with Italian configuration
	italianConfig := &LanguageConfig{
		FieldMappings: map[string]string{
			"name":        "nome",
			"level":       "livello",
			"description": "descrizione",
		},
	}
	
	parser := NewBaseParserWithLanguage(
		ContentTypeSpells,
		"Test Parser",
		"Test parser for language-aware field extraction",
		Italian,
		italianConfig,
	)

	testCases := []struct {
		name          string
		content       []string
		fieldName     string
		expectedValue string
	}{
		{
			name: "extract Italian field",
			content: []string{
				"**Nome:** Palla di Fuoco",
				"**Livello:** 3",
				"**Descrizione:** Una potente palla di fuoco",
			},
			fieldName:     "name",
			expectedValue: "Palla di Fuoco",
		},
		{
			name: "extract level field",
			content: []string{
				"**Nome:** Magic Missile",
				"**Livello:** 1",
				"**Scuola:** Evocazione",
			},
			fieldName:     "level",
			expectedValue: "1",
		},
		{
			name: "field not found",
			content: []string{
				"**Nome:** Test Spell",
				"**Scuola:** Evocazione",
			},
			fieldName:     "description",
			expectedValue: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parser.extractFieldFromLines(tc.content, tc.fieldName)
			
			if result != tc.expectedValue {
				t.Errorf("Expected '%s', got '%s'", tc.expectedValue, result)
			}
		})
	}
}

func TestBaseParser_GetFieldName(t *testing.T) {
	// Test Italian mappings
	italianConfig := &LanguageConfig{
		FieldMappings: map[string]string{
			"name":  "nome",
			"level": "livello",
		},
	}
	
	italianParser := NewBaseParserWithLanguage(
		ContentTypeSpells,
		"Italian Parser",
		"Test Italian parser",
		Italian,
		italianConfig,
	)

	// Test English mappings (direct mapping)
	englishConfig := &LanguageConfig{
		FieldMappings: map[string]string{
			"name":  "name",
			"level": "level",
		},
	}
	
	englishParser := NewBaseParserWithLanguage(
		ContentTypeSpells,
		"English Parser", 
		"Test English parser",
		English,
		englishConfig,
	)

	testCases := []struct {
		parser        *BaseParser
		englishField  string
		expectedField string
	}{
		{italianParser, "name", "nome"},
		{italianParser, "level", "livello"},
		{italianParser, "unknown_field", "unknown_field"}, // fallback
		{englishParser, "name", "name"},
		{englishParser, "level", "level"},
	}

	for _, tc := range testCases {
		t.Run(tc.englishField, func(t *testing.T) {
			result := tc.parser.getFieldName(tc.englishField)
			
			if result != tc.expectedField {
				t.Errorf("Expected '%s', got '%s'", tc.expectedField, result)
			}
		})
	}
}

func TestBaseParser_ExtractSectionsWithLanguage(t *testing.T) {
	// Test different section delimiters
	testCases := []struct {
		name      string
		config    *LanguageConfig
		content   []string
		expected  int
	}{
		{
			name: "H2 delimiter (##)",
			config: &LanguageConfig{
				SectionDelimiter: "##",
			},
			content: []string{
				"## Section 1",
				"Content 1",
				"## Section 2", 
				"Content 2",
			},
			expected: 2,
		},
		{
			name: "H3 delimiter (###)",
			config: &LanguageConfig{
				SectionDelimiter: "###",
			},
			content: []string{
				"### Section 1",
				"Content 1",
				"### Section 2",
				"Content 2",
				"## Not a section", // should be ignored
			},
			expected: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewBaseParserWithLanguage(
				ContentTypeSpells,
				"Test Parser",
				"Test section extraction",
				Italian,
				tc.config,
			)

			sections, err := parser.extractSectionsWithLanguage(tc.content)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(sections) != tc.expected {
				t.Errorf("Expected %d sections, got %d", tc.expected, len(sections))
			}
		})
	}
}