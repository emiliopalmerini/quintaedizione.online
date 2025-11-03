package parsers

import (
	"testing"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// MockDocumentStrategy is a test implementation of DocumentParsingStrategy
type MockDocumentStrategy struct {
	contentType ContentType
	name        string
	description string
}

func (m *MockDocumentStrategy) ParseDocument(content []string, context *ParsingContext) ([]*domain.Document, error) {
	return []*domain.Document{}, nil
}

func (m *MockDocumentStrategy) ContentType() ContentType {
	return m.contentType
}

func (m *MockDocumentStrategy) Name() string {
	return m.name
}

func (m *MockDocumentStrategy) Description() string {
	return m.description
}

func (m *MockDocumentStrategy) Validate(content []string) error {
	return nil
}

func TestNewDocumentRegistry(t *testing.T) {
	registry := NewDocumentRegistry()

	if registry == nil {
		t.Error("Expected non-nil registry")
	}

	if registry.strategies == nil {
		t.Error("Expected initialized strategies map")
	}

	if len(registry.strategies) != 0 {
		t.Errorf("Expected empty registry, got %d strategies", len(registry.strategies))
	}
}

func TestDocumentRegistry_Register(t *testing.T) {
	registry := NewDocumentRegistry()
	strategy := &MockDocumentStrategy{
		contentType: ContentTypeIncantesimi,
		name:        "Mock Spells",
		description: "Mock spell parser",
	}

	// Test successful registration
	err := registry.Register("spells_ita", strategy)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test duplicate registration
	err = registry.Register("spells_ita", strategy)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}

	// Test registration with empty key
	err = registry.Register("", strategy)
	if err == nil {
		t.Error("Expected error for empty key")
	}

	// Test registration with nil strategy
	err = registry.Register("nil_strategy", nil)
	if err == nil {
		t.Error("Expected error for nil strategy")
	}
}

func TestDocumentRegistry_GetStrategyByKey(t *testing.T) {
	registry := NewDocumentRegistry()
	strategy := &MockDocumentStrategy{
		contentType: ContentTypeIncantesimi,
		name:        "Mock Spells",
		description: "Mock spell parser",
	}

	key := "spells_ita"
	registry.Register(key, strategy)

	// Test successful retrieval
	retrieved, err := registry.GetStrategyByKey(key)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if retrieved != strategy {
		t.Error("Expected retrieved strategy to match registered strategy")
	}

	// Test retrieval of non-existent key
	_, err = registry.GetStrategyByKey("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent key")
	}
}

func TestDocumentRegistry_GetStrategy(t *testing.T) {
	registry := NewDocumentRegistry()
	strategy := &MockDocumentStrategy{
		contentType: ContentTypeIncantesimi,
		name:        "Mock Spells",
		description: "Mock spell parser",
	}

	key := "incantesimi_ita"
	registry.Register(key, strategy)

	// Test successful retrieval
	retrieved, err := registry.GetStrategy(ContentTypeIncantesimi, Italian)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if retrieved != strategy {
		t.Error("Expected retrieved strategy to match registered strategy")
	}

	// Test retrieval of non-existent strategy
	_, err = registry.GetStrategy(ContentType("nonexistent"), Italian)
	if err == nil {
		t.Error("Expected error for non-existent strategy")
	}
}

func TestDocumentRegistry_ListKeys(t *testing.T) {
	registry := NewDocumentRegistry()

	// Register multiple strategies
	strategies := map[string]*MockDocumentStrategy{
		"spells_ita": {
			contentType: ContentTypeIncantesimi,
			name:        "Spells",
		},
		"weapons_ita": {
			contentType: ContentTypeArmi,
			name:        "Weapons",
		},
		"armor_ita": {
			contentType: ContentTypeArmature,
			name:        "Armor",
		},
	}

	for key, strategy := range strategies {
		registry.Register(key, strategy)
	}

	keys := registry.ListKeys()

	if len(keys) != len(strategies) {
		t.Errorf("Expected %d keys, got %d", len(strategies), len(keys))
	}

	// Verify all expected keys are present
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[key] = true
	}

	for expectedKey := range strategies {
		if !keyMap[expectedKey] {
			t.Errorf("Expected key %s not found in list", expectedKey)
		}
	}
}

func TestDocumentRegistry_Count(t *testing.T) {
	registry := NewDocumentRegistry()

	// Initially empty
	if count := registry.Count(); count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// Add strategies
	strategy := &MockDocumentStrategy{
		contentType: ContentTypeIncantesimi,
		name:        "Spells",
	}

	registry.Register("spells_ita", strategy)
	if count := registry.Count(); count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	registry.Register("weapons_ita", strategy)
	if count := registry.Count(); count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestCreateDocumentRegistry(t *testing.T) {
	registry, err := CreateDocumentRegistry()
	if err != nil {
		t.Fatalf("Expected no error creating registry, got %v", err)
	}

	if registry == nil {
		t.Error("Expected non-nil registry")
	}

	// Should have registered strategies for all content types
	expectedStrategies := 14 // Based on the CreateDocumentRegistry function

	if count := registry.Count(); count != expectedStrategies {
		t.Errorf("Expected %d strategies, got %d", expectedStrategies, count)
	}

	// Test that we can retrieve a specific strategy
	strategy, err := registry.GetStrategy(ContentTypeIncantesimi, Italian)
	if err != nil {
		t.Errorf("Expected no error retrieving spells strategy, got %v", err)
	}
	if strategy == nil {
		t.Error("Expected non-nil spells strategy")
	}
	if strategy.ContentType() != ContentTypeIncantesimi {
		t.Errorf("Expected content type %s, got %s", ContentTypeIncantesimi, strategy.ContentType())
	}
}
