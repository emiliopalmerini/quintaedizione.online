package parsers

import (
	"testing"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

func TestRegoleStrategy(t *testing.T) {
	content := []string{
		"# Regole",
		"",
		"## Condizioni di convenzione",
		"",
		"Questo glossario definisce le convenzioni usate in questo manuale.",
		"",
		"## Abbreviazioni",
		"",
		"CA = Classe Armatura",
		"",
		"## Vantaggio/Svantaggio",
		"",
		"A volte, un'abilità speciale o un incantesimo ti dice che hai Vantaggio o Svantaggio su un tiro di caratteristica.",
		"",
		"Quando ciò accade, tiri un secondo d20. Se hai Vantaggio, usa il risultato più alto.",
		"",
		"## Condizioni",
		"",
		"Le condizioni alterano le capacità di una creatura in vari modi.",
	}

	strategy := NewRegoleStrategy()
	context := NewParsingContext("regole.md", "ita")

	entities, err := strategy.Parse(content, context)
	if err != nil {
		t.Fatalf("Failed to parse regole: %v", err)
	}

	// Should skip convention sections and only parse glossary entries
	if len(entities) < 1 {
		t.Fatalf("Expected at least 1 regola, got %d", len(entities))
	}

	// Find the Vantaggio/Svantaggio rule
	var vantaggioRule *domain.Regola
	for _, entity := range entities {
		if rule, ok := entity.(*domain.Regola); ok {
			if rule.Nome == "Vantaggio/Svantaggio" {
				vantaggioRule = rule
				break
			}
		}
	}

	if vantaggioRule == nil {
		t.Fatal("Could not find Vantaggio/Svantaggio rule")
	}

	if vantaggioRule.EntityType() != "regola" {
		t.Errorf("Expected entity type 'regola', got '%s'", vantaggioRule.EntityType())
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
