package parsers

import (
	"testing"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

func TestEquipaggiamentiStrategy(t *testing.T) {
	content := []string{
		"# Equipaggiamento",
		"",
		"## Corda di Canapa (15 m) (1 mo)",
		"",
		"**Peso:** 5 kg",
		"",
		"Una corda robusta per arrampicarsi o legare.",
		"",
		"## Zaino (2 mo)",
		"",
		"**Peso:** 2,5 kg",
		"**Capacità:** 30 l",
		"",
		"Uno zaino per trasportare equipaggiamento.",
	}

	strategy := NewEquipaggiamentiStrategy()
	context := NewParsingContext("equipaggiamenti.md", "ita")

	entities, err := strategy.Parse(content, context)
	if err != nil {
		t.Fatalf("Failed to parse equipaggiamenti: %v", err)
	}

	if len(entities) != 2 {
		t.Fatalf("Expected 2 equipaggiamenti, got %d", len(entities))
	}

	// Test first equipment
	corda, ok := entities[0].(*domain.Equipaggiamento)
	if !ok {
		t.Fatal("First entity is not an Equipaggiamento")
	}

	if corda.Nome != "Corda di Canapa (15 m)" {
		t.Errorf("Expected nome 'Corda di Canapa (15 m)', got '%s'", corda.Nome)
	}

	if corda.EntityType() != "equipaggiamento" {
		t.Errorf("Expected entity type 'equipaggiamento', got '%s'", corda.EntityType())
	}

	if corda.Costo.Valore != 1 || corda.Costo.Valuta != domain.ValutaOro {
		t.Errorf("Expected costo 1 mo, got %d %s", corda.Costo.Valore, corda.Costo.Valuta)
	}

	// Test second equipment with capacity
	zaino, ok := entities[1].(*domain.Equipaggiamento)
	if !ok {
		t.Fatal("Second entity is not an Equipaggiamento")
	}

	if zaino.Capacita == nil {
		t.Error("Expected capacita to be set for zaino")
	}
}

func TestEquipaggiamentiStrategy_EmptyContent(t *testing.T) {
	strategy := NewEquipaggiamentiStrategy()
	context := NewParsingContext("equipaggiamenti.md", "ita")

	entities, err := strategy.Parse([]string{}, context)
	if err == nil {
		t.Error("Expected error for empty content")
	}
	if len(entities) != 0 {
		t.Error("Expected no entities for empty content")
	}
}

func TestEquipaggiamentiStrategy_Interface(t *testing.T) {
	strategy := NewEquipaggiamentiStrategy()

	if strategy.ContentType() != ContentTypeEquipaggiamenti {
		t.Errorf("Expected content type %s, got %s", ContentTypeEquipaggiamenti, strategy.ContentType())
	}

	if strategy.Name() == "" {
		t.Error("Strategy name should not be empty")
	}

	if strategy.Description() == "" {
		t.Error("Strategy description should not be empty")
	}
}
