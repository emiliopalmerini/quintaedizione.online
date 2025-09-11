package parsers

import (
	"testing"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

func TestArmatureStrategy(t *testing.T) {
	content := []string{
		"# Armature",
		"",
		"## Armatura Imbottita",
		"",
		"**Costo:** 5 mo",
		"**Peso:** 3,5 kg",
		"**Categoria:** Leggera",
		"**CA Base:** 11",
		"**CA + Des:** sì",
		"**Limite Des:** —",
		"**Forza richiesta:** —",
		"**Svantaggio Furtività:** sì",
		"",
		"## Corazza a Scaglie",
		"",
		"**Costo:** 50 mo",
		"**Peso:** 20 kg",
		"**Categoria:** Media",
		"**CA Base:** 14",
		"**CA + Des:** sì",
		"**Limite Des:** 2",
		"**Forza richiesta:** —",
		"**Svantaggio Furtività:** sì",
	}

	strategy := NewArmatureStrategy()
	context := NewParsingContext("armature.md", "ita")

	entities, err := strategy.Parse(content, context)
	if err != nil {
		t.Fatalf("Failed to parse armature: %v", err)
	}

	if len(entities) != 2 {
		t.Fatalf("Expected 2 armature, got %d", len(entities))
	}

	// Test first armor
	imbottita, ok := entities[0].(*domain.Armatura)
	if !ok {
		t.Fatal("First entity is not an Armatura")
	}

	if imbottita.Nome != "Armatura Imbottita" {
		t.Errorf("Expected name 'Armatura Imbottita', got '%s'", imbottita.Nome)
	}

	if imbottita.EntityType() != "armatura" {
		t.Errorf("Expected entity type 'armatura', got '%s'", imbottita.EntityType())
	}

	// Test second armor - has limits
	scaglie, ok := entities[1].(*domain.Armatura)
	if !ok {
		t.Fatal("Second entity is not an Armatura")
	}

	if scaglie.Nome != "Corazza a Scaglie" {
		t.Errorf("Expected name 'Corazza a Scaglie', got '%s'", scaglie.Nome)
	}
}

func TestArmatureStrategy_EmptyContent(t *testing.T) {
	strategy := NewArmatureStrategy()
	context := NewParsingContext("armature.md", "ita")

	entities, err := strategy.Parse([]string{}, context)
	if err == nil {
		t.Error("Expected error for empty content")
	}
	if len(entities) != 0 {
		t.Error("Expected no entities for empty content")
	}
}

func TestArmatureStrategy_Interface(t *testing.T) {
	strategy := NewArmatureStrategy()

	if strategy.ContentType() != ContentTypeArmature {
		t.Errorf("Expected content type %s, got %s", ContentTypeArmature, strategy.ContentType())
	}

	if strategy.Name() == "" {
		t.Error("Strategy name should not be empty")
	}

	if strategy.Description() == "" {
		t.Error("Strategy description should not be empty")
	}
}