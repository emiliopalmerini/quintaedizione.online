package parsers

import (
	"testing"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

func TestOggettiMagiciStrategy(t *testing.T) {
	content := []string{
		"# Oggetti Magici",
		"",
		"## Spada Affilata +1",
		"",
		"*Arma (qualsiasi spada), non comune*",
		"",
		"Hai un bonus di +1 ai tiri per colpire e ai danni effettuati con quest'arma magica.",
		"",
		"## Anello della Protezione",
		"",
		"*Anello, raro (richiede sintonia)*",
		"",
		"Mentre indossi questo anello, ottieni un bonus di +1 alla CA e ai tiri salvezza.",
	}

	strategy := NewOggettiMagiciStrategy()
	context := NewParsingContext("oggetti_magici.md", "ita")

	entities, err := strategy.Parse(content, context)
	if err != nil {
		t.Fatalf("Failed to parse oggetti magici: %v", err)
	}

	if len(entities) != 2 {
		t.Fatalf("Expected 2 oggetti magici, got %d", len(entities))
	}

	// Test first item (no attunement)
	spada, ok := entities[0].(*domain.OggettoMagico)
	if !ok {
		t.Fatal("First entity is not an OggettoMagico")
	}

	if spada.Nome != "Spada Affilata +1" {
		t.Errorf("Expected nome 'Spada Affilata +1', got '%s'", spada.Nome)
	}

	if spada.EntityType() != "oggetto_magico" {
		t.Errorf("Expected entity type 'oggetto_magico', got '%s'", spada.EntityType())
	}

	if spada.Rarita != domain.RaritaNonComune {
		t.Errorf("Expected rarita 'non comune', got %s", spada.Rarita)
	}

	if spada.Sintonizzazione {
		t.Error("Expected sintonizzazione to be false for first item")
	}

	// Test second item (with attunement)
	anello, ok := entities[1].(*domain.OggettoMagico)
	if !ok {
		t.Fatal("Second entity is not an OggettoMagico")
	}

	if anello.Rarita != domain.RaritaRara {
		t.Errorf("Expected rarita 'raro', got %s", anello.Rarita)
	}

	if !anello.Sintonizzazione {
		t.Error("Expected sintonizzazione to be true for second item")
	}
}

func TestOggettiMagiciStrategy_EmptyContent(t *testing.T) {
	strategy := NewOggettiMagiciStrategy()
	context := NewParsingContext("oggetti_magici.md", "ita")

	entities, err := strategy.Parse([]string{}, context)
	if err == nil {
		t.Error("Expected error for empty content")
	}
	if len(entities) != 0 {
		t.Error("Expected no entities for empty content")
	}
}

func TestOggettiMagiciStrategy_Interface(t *testing.T) {
	strategy := NewOggettiMagiciStrategy()

	if strategy.ContentType() != ContentTypeOggettiMagici {
		t.Errorf("Expected content type %s, got %s", ContentTypeOggettiMagici, strategy.ContentType())
	}

	if strategy.Name() == "" {
		t.Error("Strategy name should not be empty")
	}

	if strategy.Description() == "" {
		t.Error("Strategy description should not be empty")
	}
}
