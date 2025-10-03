package parsers

import (
	"testing"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

func TestStrumentiStrategy(t *testing.T) {
	content := []string{
		"# Strumenti",
		"",
		"## Strumenti da Alchimista (50 MO)",
		"",
		"**Peso:** 3,5 kg",
		"**Abilità:** Natura",
		"**Utilizzo:** Identificare una sostanza (CD 15) o Creare un acido (CD 20)",
		"**Creazioni:** Acido, Fuoco dell'alchimista, Antitossina",
		"",
		"Questi strumenti permettono a un personaggio di produrre composti alchemici.",
		"",
		"## Attrezzi da Ladro (25 MO)",
		"",
		"**Peso:** 0,5 kg",
		"**Abilità:** Furtività",
		"**Utilizzo:** Scassinare una serratura (CD 15)",
		"",
		"Questo set include grimaldelli e piccoli strumenti.",
	}

	strategy := NewStrumentiStrategy()
	context := NewParsingContext("strumenti.md", "ita")

	entities, err := strategy.Parse(content, context)
	if err != nil {
		t.Fatalf("Failed to parse strumenti: %v", err)
	}

	if len(entities) != 2 {
		t.Fatalf("Expected 2 strumenti, got %d", len(entities))
	}

	// Test first tool
	alchimista, ok := entities[0].(*domain.Strumento)
	if !ok {
		t.Fatal("First entity is not a Strumento")
	}

	if alchimista.Nome != "Strumenti da Alchimista" {
		t.Errorf("Expected nome 'Strumenti da Alchimista', got '%s'", alchimista.Nome)
	}

	if alchimista.EntityType() != "strumento" {
		t.Errorf("Expected entity type 'strumento', got '%s'", alchimista.EntityType())
	}

	if alchimista.Costo.Valore != 50 || alchimista.Costo.Valuta != domain.ValutaOro {
		t.Errorf("Expected costo 50 mo, got %d %s", alchimista.Costo.Valore, alchimista.Costo.Valuta)
	}

	if alchimista.AbilitaAssociata != "natura" {
		t.Errorf("Expected abilita 'natura', got '%s'", alchimista.AbilitaAssociata)
	}

	if len(alchimista.Utilizzi) < 1 {
		t.Error("Expected at least one utilizzo")
	}

	if len(alchimista.Creazioni) != 3 {
		t.Errorf("Expected 3 creazioni, got %d", len(alchimista.Creazioni))
	}

	// Test second tool
	ladro, ok := entities[1].(*domain.Strumento)
	if !ok {
		t.Fatal("Second entity is not a Strumento")
	}

	if ladro.AbilitaAssociata != "furtivita" {
		t.Errorf("Expected abilita 'furtivita', got '%s'", ladro.AbilitaAssociata)
	}
}

func TestStrumentiStrategy_EmptyContent(t *testing.T) {
	strategy := NewStrumentiStrategy()
	context := NewParsingContext("strumenti.md", "ita")

	entities, err := strategy.Parse([]string{}, context)
	if err == nil {
		t.Error("Expected error for empty content")
	}
	if len(entities) != 0 {
		t.Error("Expected no entities for empty content")
	}
}

func TestStrumentiStrategy_Interface(t *testing.T) {
	strategy := NewStrumentiStrategy()

	if strategy.ContentType() != ContentTypeStrumenti {
		t.Errorf("Expected content type %s, got %s", ContentTypeStrumenti, strategy.ContentType())
	}

	if strategy.Name() == "" {
		t.Error("Strategy name should not be empty")
	}

	if strategy.Description() == "" {
		t.Error("Strategy description should not be empty")
	}
}
