package parsers

import (
	"testing"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

func TestBackgroundsStrategy(t *testing.T) {
	content := []string{
		"# Background",
		"",
		"## Accolito",
		"",
		"**Punteggi di Caratteristica:** Intelligenza, Saggezza, Carisma",
		"**Competenze in Abilità:** Intuizione e Religione",
		"**Competenza negli Strumenti:** Nessuna",
		"**Talento:** Iniziato alla Magia (Chierico)",
		"**Equipaggiamento:** *Scegli A o B:* (A) Simbolo Sacro, Libro di Preghiere, 5 Bastoncini d'Incenso, Vesti, Zaino, 8 mo; oppure (B) 50 mo",
		"",
		"Hai servito come accolito in un tempio.",
	}

	strategy := NewBackgroundsStrategy()
	context := NewParsingContext("backgrounds.md", "ita")

	entities, err := strategy.Parse(content, context)
	if err != nil {
		t.Fatalf("Failed to parse backgrounds: %v", err)
	}

	if len(entities) != 1 {
		t.Fatalf("Expected 1 background, got %d", len(entities))
	}

	background, ok := entities[0].(*domain.Background)
	if !ok {
		t.Fatal("Entity is not a Background")
	}

	if background.Nome != "Accolito" {
		t.Errorf("Expected nome 'Accolito', got '%s'", background.Nome)
	}

	if background.EntityType() != "background" {
		t.Errorf("Expected entity type 'background', got '%s'", background.EntityType())
	}

	if len(background.Caratteristiche) != 3 {
		t.Errorf("Expected 3 caratteristiche, got %d", len(background.Caratteristiche))
	}

	if len(background.CompetenzeAbilita) != 2 {
		t.Errorf("Expected 2 competenze abilita, got %d", len(background.CompetenzeAbilita))
	}
}

func TestBackgroundsStrategy_EmptyContent(t *testing.T) {
	strategy := NewBackgroundsStrategy()
	context := NewParsingContext("backgrounds.md", "ita")

	entities, err := strategy.Parse([]string{}, context)
	if err == nil {
		t.Error("Expected error for empty content")
	}
	if len(entities) != 0 {
		t.Error("Expected no entities for empty content")
	}
}

func TestBackgroundsStrategy_Interface(t *testing.T) {
	strategy := NewBackgroundsStrategy()

	if strategy.ContentType() != ContentTypeBackgrounds {
		t.Errorf("Expected content type %s, got %s", ContentTypeBackgrounds, strategy.ContentType())
	}

	if strategy.Name() == "" {
		t.Error("Strategy name should not be empty")
	}

	if strategy.Description() == "" {
		t.Error("Strategy description should not be empty")
	}
}
