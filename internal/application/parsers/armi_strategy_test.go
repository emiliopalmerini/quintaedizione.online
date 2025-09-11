package parsers

import (
	"testing"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

func TestArmiStrategy(t *testing.T) {
	content := []string{
		"# Armi",
		"",
		"## Randello",
		"",
		"**Costo:** 1 ma",
		"**Peso:** 1 kg",
		"**Danno:** 1d4 Contundente",
		"**Categoria:** Semplice da Mischia",
		"**Proprietà:** Leggera",
		"**Maestria:** Rallenta",
		"**Gittata:** 0",
		"**Gittata lunga:** 0",
		"",
		"## Arco Lungo",
		"",
		"**Costo:** 50 mo",
		"**Peso:** 1 kg",
		"**Danno:** 1d8 Perforante",
		"**Categoria:** Marziale a Distanza",
		"**Proprietà:** Munizioni (Freccia), Pesante, A Due Mani",
		"**Maestria:** Rallenta",
		"**Gittata:** 45 m",
		"**Gittata lunga:** 180 m",
	}

	strategy := NewArmiStrategy()
	context := NewParsingContext("armi.md", "ita")

	entities, err := strategy.Parse(content, context)
	if err != nil {
		t.Fatalf("Failed to parse armi: %v", err)
	}

	if len(entities) != 2 {
		t.Fatalf("Expected 2 armi, got %d", len(entities))
	}

	// Test first weapon (melee)
	randello, ok := entities[0].(*domain.Arma)
	if !ok {
		t.Fatal("First entity is not an Arma")
	}

	if randello.Nome != "Randello" {
		t.Errorf("Expected name 'Randello', got '%s'", randello.Nome)
	}

	if randello.EntityType() != "arma" {
		t.Errorf("Expected entity type 'arma', got '%s'", randello.EntityType())
	}

	if randello.Categoria != domain.CategoriaArmaSimpliceMischia {
		t.Errorf("Expected categoria 'Semplice da Mischia', got '%s'", randello.Categoria)
	}

	if len(randello.Proprieta) != 1 || randello.Proprieta[0] != domain.ProprietaLeggera {
		t.Errorf("Expected properties [Leggera], got %v", randello.Proprieta)
	}

	// Test second weapon (ranged with range)
	arcoLungo, ok := entities[1].(*domain.Arma)
	if !ok {
		t.Fatal("Second entity is not an Arma")
	}

	if arcoLungo.Nome != "Arco Lungo" {
		t.Errorf("Expected name 'Arco Lungo', got '%s'", arcoLungo.Nome)
	}

	if arcoLungo.Categoria != domain.CategoriaArmaMarzialeDistanza {
		t.Errorf("Expected categoria 'Marziale a Distanza', got '%s'", arcoLungo.Categoria)
	}

	if arcoLungo.Gittata == nil {
		t.Fatal("Expected gittata to be set for ranged weapon")
	}

	if arcoLungo.Gittata.Normale != "45 m" {
		t.Errorf("Expected normale range '45 m', got '%s'", arcoLungo.Gittata.Normale)
	}

	if arcoLungo.Gittata.Lunga != "180 m" {
		t.Errorf("Expected lunga range '180 m', got '%s'", arcoLungo.Gittata.Lunga)
	}
}

func TestArmiStrategy_EmptyContent(t *testing.T) {
	strategy := NewArmiStrategy()
	context := NewParsingContext("armi.md", "ita")

	entities, err := strategy.Parse([]string{}, context)
	if err == nil {
		t.Error("Expected error for empty content")
	}
	if len(entities) != 0 {
		t.Error("Expected no entities for empty content")
	}
}

func TestArmiStrategy_Interface(t *testing.T) {
	strategy := NewArmiStrategy()

	if strategy.ContentType() != ContentTypeArmi {
		t.Errorf("Expected content type %s, got %s", ContentTypeArmi, strategy.ContentType())
	}

	if strategy.Name() == "" {
		t.Error("Strategy name should not be empty")
	}

	if strategy.Description() == "" {
		t.Error("Strategy description should not be empty")
	}
}

func TestArmiStrategy_PropertiesParsing(t *testing.T) {
	content := []string{
		"# Armi",
		"",
		"## Test Weapon",
		"",
		"**Costo:** 5 mo",
		"**Peso:** 1 kg",
		"**Danno:** 1d6 Tagliente",
		"**Categoria:** Semplice da Mischia",
		"**Proprietà:** Accurata, Leggera, Da Lancio",
		"**Maestria:** Colpisci",
		"**Gittata:** 6 m",
		"**Gittata lunga:** 18 m",
	}

	strategy := NewArmiStrategy()
	context := NewParsingContext("armi.md", "ita")

	entities, err := strategy.Parse(content, context)
	if err != nil {
		t.Fatalf("Failed to parse armi: %v", err)
	}

	if len(entities) != 1 {
		t.Fatalf("Expected 1 arma, got %d", len(entities))
	}

	arma, ok := entities[0].(*domain.Arma)
	if !ok {
		t.Fatal("Entity is not an Arma")
	}

	expectedProperties := []domain.ProprietaArma{
		domain.ProprietaAccurata,
		domain.ProprietaLeggera,
		domain.ProprietaDaLancio,
	}

	if len(arma.Proprieta) != len(expectedProperties) {
		t.Fatalf("Expected %d properties, got %d", len(expectedProperties), len(arma.Proprieta))
	}

	for i, expected := range expectedProperties {
		if arma.Proprieta[i] != expected {
			t.Errorf("Expected property %s at index %d, got %s", expected, i, arma.Proprieta[i])
		}
	}
}