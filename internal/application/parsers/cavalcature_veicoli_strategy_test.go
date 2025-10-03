package parsers

import (
	"testing"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

func TestCavalcatureVeicoliStrategy(t *testing.T) {
	content := []string{
		"# Cavalcature e Veicoli",
		"",
		"## Cavallo da Sella",
		"",
		"**Tipo:** Cavalcatura",
		"**Costo:** 75 mo",
		"**Velocità:** 18 m",
		"**Capacità Carico:** 225 kg",
		"",
		"Un cavallo addestrato per essere cavalcato.",
		"",
		"## Carrozza",
		"",
		"**Tipo:** Veicolo",
		"**Costo:** 100 mo",
		"**Velocità:** 3 m",
		"**Capacità Carico:** 300 kg",
		"",
		"Un veicolo a quattro ruote trainato da cavalli.",
	}

	strategy := NewCavalcatureVeicoliStrategy()
	context := NewParsingContext("cavalcature_veicoli.md", "ita")

	entities, err := strategy.Parse(content, context)
	if err != nil {
		t.Fatalf("Failed to parse cavalcature/veicoli: %v", err)
	}

	if len(entities) != 2 {
		t.Fatalf("Expected 2 entities, got %d", len(entities))
	}

	// Test first entity (mount)
	cavallo, ok := entities[0].(*domain.CavalcaturaVeicolo)
	if !ok {
		t.Fatal("First entity is not a CavalcaturaVeicolo")
	}

	if cavallo.Nome != "Cavallo da Sella" {
		t.Errorf("Expected nome 'Cavallo da Sella', got '%s'", cavallo.Nome)
	}

	if cavallo.EntityType() != "cavalcatura_veicolo" {
		t.Errorf("Expected entity type 'cavalcatura_veicolo', got '%s'", cavallo.EntityType())
	}

	if cavallo.Tipo != domain.TipoCavalcatura {
		t.Errorf("Expected tipo Cavalcatura, got %s", cavallo.Tipo)
	}

	// Test second entity (vehicle)
	carrozza, ok := entities[1].(*domain.CavalcaturaVeicolo)
	if !ok {
		t.Fatal("Second entity is not a CavalcaturaVeicolo")
	}

	if carrozza.Tipo != domain.TipoVeicolo {
		t.Errorf("Expected tipo Veicolo, got %s", carrozza.Tipo)
	}
}

func TestCavalcatureVeicoliStrategy_EmptyContent(t *testing.T) {
	strategy := NewCavalcatureVeicoliStrategy()
	context := NewParsingContext("cavalcature_veicoli.md", "ita")

	entities, err := strategy.Parse([]string{}, context)
	if err == nil {
		t.Error("Expected error for empty content")
	}
	if len(entities) != 0 {
		t.Error("Expected no entities for empty content")
	}
}

func TestCavalcatureVeicoliStrategy_Interface(t *testing.T) {
	strategy := NewCavalcatureVeicoliStrategy()

	if strategy.ContentType() != ContentTypeCavalcatureVeicoli {
		t.Errorf("Expected content type %s, got %s", ContentTypeCavalcatureVeicoli, strategy.ContentType())
	}

	if strategy.Name() == "" {
		t.Error("Strategy name should not be empty")
	}

	if strategy.Description() == "" {
		t.Error("Strategy description should not be empty")
	}
}
