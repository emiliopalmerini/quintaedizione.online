package parsers

import (
	"strings"
	"testing"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

func TestRegoleStrategy(t *testing.T) {
	content := []string{
		"# Glossario delle Regole",
		"",
		"Questo glossario usa le seguenti convenzioni:",
		"",
		"## Prova di caratteristica",
		"",
		"Una prova di caratteristica è un Test con d20 che rappresenta l'uso di una delle sei caratteristiche—o di una specifica abilità associata a una caratteristica—per superare una sfida. Vedi anche \"Giocare\" (\"Test con d20\" e \"Competenza\").",
		"",
		"## Vantaggio",
		"",
		"Se hai Vantaggio a un Test con d20, tira due d20 e usa il risultato più alto. Un tiro non può essere influenzato da più di un Vantaggio, e Vantaggio e Svantaggio sullo stesso tiro si annullano a vicenda. Vedi anche \"Giocare\" (\"Test con d20\").",
		"",
		"## Azione [Azione]",
		"",
		"Nel tuo turno puoi compiere un'azione. Scegli quale azione compiere tra quelle sotto o tra le azioni speciali fornite dalle tue caratteristiche. Vedi anche \"Giocare\" (\"Azioni\").",
		"",
		"Queste azioni sono definite altrove in questo glossario:",
		"",
		"|           |         |             |           |",
		"|-----------|---------|-------------|-----------|",
		"| Attacco   | Schivata| Influenzare | Cercare   |",
		"| Scatto    | Aiutare | Magia       | Studiare  |",
	}

	strategy := NewRegoleStrategy()
	context := NewParsingContext("regole.md", "ita")

	entities, err := strategy.Parse(content, context)
	if err != nil {
		t.Fatalf("Failed to parse regole: %v", err)
	}

	if len(entities) != 3 {
		t.Fatalf("Expected 3 regole, got %d", len(entities))
	}

	// Test first rule
	provaCaratteristica, ok := entities[0].(*domain.Regola)
	if !ok {
		t.Fatal("First entity is not a Regola")
	}

	if provaCaratteristica.Nome != "Prova di caratteristica" {
		t.Errorf("Expected name 'Prova di caratteristica', got '%s'", provaCaratteristica.Nome)
	}

	if provaCaratteristica.EntityType() != "regola" {
		t.Errorf("Expected entity type 'regola', got '%s'", provaCaratteristica.EntityType())
	}

	expectedContent := "Una prova di caratteristica è un Test con d20 che rappresenta l'uso di una delle sei caratteristiche—o di una specifica abilità associata a una caratteristica—per superare una sfida. Vedi anche \"Giocare\" (\"Test con d20\" e \"Competenza\").\n"
	if provaCaratteristica.Contenuto != expectedContent {
		t.Errorf("Expected content '%s', got '%s'", expectedContent, provaCaratteristica.Contenuto)
	}

	// Test second rule
	vantaggio, ok := entities[1].(*domain.Regola)
	if !ok {
		t.Fatal("Second entity is not a Regola")
	}

	if vantaggio.Nome != "Vantaggio" {
		t.Errorf("Expected name 'Vantaggio', got '%s'", vantaggio.Nome)
	}

	// Test third rule with tag
	azione, ok := entities[2].(*domain.Regola)
	if !ok {
		t.Fatal("Third entity is not a Regola")
	}

	if azione.Nome != "Azione [Azione]" {
		t.Errorf("Expected name 'Azione [Azione]', got '%s'", azione.Nome)
	}

	// Verify the content includes tables and multiple paragraphs
	if !contains(azione.Contenuto, "Nel tuo turno puoi compiere un'azione") {
		t.Error("Expected content to contain action description")
	}

	if !contains(azione.Contenuto, "| Attacco   | Schivata| Influenzare | Cercare   |") {
		t.Error("Expected content to contain action table")
	}
}

func TestRegoleStrategy_EmptyContent(t *testing.T) {
	strategy := NewRegoleStrategy()
	context := NewParsingContext("regole.md", "ita")

	entities, err := strategy.Parse([]string{}, context)
	if err == nil {
		t.Error("Expected error for empty content")
	}
	if len(entities) != 0 {
		t.Error("Expected no entities for empty content")
	}
}

func TestRegoleStrategy_Interface(t *testing.T) {
	strategy := NewRegoleStrategy()

	if strategy.ContentType() != ContentTypeRegole {
		t.Errorf("Expected content type %s, got %s", ContentTypeRegole, strategy.ContentType())
	}

	if strategy.Name() == "" {
		t.Error("Strategy name should not be empty")
	}

	if strategy.Description() == "" {
		t.Error("Strategy description should not be empty")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}