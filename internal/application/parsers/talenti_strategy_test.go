package parsers

import (
	"testing"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

func TestTalentiStrategy(t *testing.T) {
	content := []string{
		"# Talenti",
		"",
		"## Allerta",
		"",
		"*Talento di Origine*",
		"",
		"Ottieni i seguenti benefici:",
		"",
		"**Bonus all'Iniziativa.** Ottieni un bonus di +5 all'iniziativa.",
		"",
		"**Non Colto di Sorpresa.** Non puoi essere sorpreso.",
		"",
		"## Maestria nelle Armi",
		"",
		"*Combattimento (Requisito: Competenza con almeno un'arma)*",
		"",
		"Ottieni i seguenti benefici:",
		"",
		"**Maestria.** Ottieni la maestria con tre armi a tua scelta.",
	}

	strategy := NewTalentiStrategy()
	context := NewParsingContext("talenti.md", "ita")

	entities, err := strategy.Parse(content, context)
	if err != nil {
		t.Fatalf("Failed to parse talenti: %v", err)
	}

	if len(entities) != 2 {
		t.Fatalf("Expected 2 talenti, got %d", len(entities))
	}

	// Test first talent (origin feat)
	allerta, ok := entities[0].(*domain.Talento)
	if !ok {
		t.Fatal("First entity is not a Talento")
	}

	if allerta.Nome != "Allerta" {
		t.Errorf("Expected nome 'Allerta', got '%s'", allerta.Nome)
	}

	if allerta.EntityType() != "talento" {
		t.Errorf("Expected entity type 'talento', got '%s'", allerta.EntityType())
	}

	if allerta.Categoria != domain.CategoriaTalentoOrigine {
		t.Errorf("Expected categoria Origine, got %s", allerta.Categoria)
	}

	if len(allerta.Benefici) < 1 {
		t.Error("Expected at least one beneficio")
	}

	// Test second talent (with prerequisites)
	maestria, ok := entities[1].(*domain.Talento)
	if !ok {
		t.Fatal("Second entity is not a Talento")
	}

	if maestria.Categoria != domain.CategoriaTalentoCombat {
		t.Errorf("Expected categoria Combattimento, got %s", maestria.Categoria)
	}

	if maestria.Prerequisiti == "" {
		t.Error("Expected prerequisiti to be set")
	}
}

func TestTalentiStrategy_EmptyContent(t *testing.T) {
	strategy := NewTalentiStrategy()
	context := NewParsingContext("talenti.md", "ita")

	entities, err := strategy.Parse([]string{}, context)
	if err == nil {
		t.Error("Expected error for empty content")
	}
	if len(entities) != 0 {
		t.Error("Expected no entities for empty content")
	}
}

func TestTalentiStrategy_Interface(t *testing.T) {
	strategy := NewTalentiStrategy()

	if strategy.ContentType() != ContentTypeTalenti {
		t.Errorf("Expected content type %s, got %s", ContentTypeTalenti, strategy.ContentType())
	}

	if strategy.Name() == "" {
		t.Error("Strategy name should not be empty")
	}

	if strategy.Description() == "" {
		t.Error("Strategy description should not be empty")
	}
}
