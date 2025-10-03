package parsers

import (
	"testing"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

func TestServiziStrategy(t *testing.T) {
	content := []string{
		"# Servizi",
		"",
		"## Alloggio alla Locanda (per notte)",
		"",
		"**Costo:** 5 ma",
		"**Categoria:** Alloggio",
		"**Descrizione:** Una notte di riposo in una locanda modesta.",
		"",
		"## Viaggio via Nave (per miglio)",
		"",
		"**Costo:** 1 mo",
		"**Categoria:** Trasporto",
		"**Descrizione:** Trasporto via nave mercantile.",
	}

	strategy := NewServiziStrategy()
	context := NewParsingContext("servizi.md", "ita")

	entities, err := strategy.Parse(content, context)
	if err != nil {
		t.Fatalf("Failed to parse servizi: %v", err)
	}

	if len(entities) != 2 {
		t.Fatalf("Expected 2 servizi, got %d", len(entities))
	}

	// Test first service
	alloggio, ok := entities[0].(*domain.Servizio)
	if !ok {
		t.Fatal("First entity is not a Servizio")
	}

	if alloggio.Nome != "Alloggio alla Locanda (per notte)" {
		t.Errorf("Expected nome 'Alloggio alla Locanda (per notte)', got '%s'", alloggio.Nome)
	}

	if alloggio.EntityType() != "servizio" {
		t.Errorf("Expected entity type 'servizio', got '%s'", alloggio.EntityType())
	}

	if alloggio.Costo.Valore != 5 || alloggio.Costo.Valuta != "ma" {
		t.Errorf("Expected costo 5 ma, got %d %s", alloggio.Costo.Valore, alloggio.Costo.Valuta)
	}

	if alloggio.Categoria != domain.CategoriaAlloggio {
		t.Errorf("Expected categoria Alloggio, got %s", alloggio.Categoria)
	}

	// Test second service
	viaggio, ok := entities[1].(*domain.Servizio)
	if !ok {
		t.Fatal("Second entity is not a Servizio")
	}

	if viaggio.Categoria != domain.CategoriaTrasporto {
		t.Errorf("Expected categoria Trasporto, got %s", viaggio.Categoria)
	}
}

func TestServiziStrategy_EmptyContent(t *testing.T) {
	strategy := NewServiziStrategy()
	context := NewParsingContext("servizi.md", "ita")

	entities, err := strategy.Parse([]string{}, context)
	if err == nil {
		t.Error("Expected error for empty content")
	}
	if len(entities) != 0 {
		t.Error("Expected no entities for empty content")
	}
}

func TestServiziStrategy_Interface(t *testing.T) {
	strategy := NewServiziStrategy()

	if strategy.ContentType() != ContentTypeServizi {
		t.Errorf("Expected content type %s, got %s", ContentTypeServizi, strategy.ContentType())
	}

	if strategy.Name() == "" {
		t.Error("Strategy name should not be empty")
	}

	if strategy.Description() == "" {
		t.Error("Strategy description should not be empty")
	}
}
