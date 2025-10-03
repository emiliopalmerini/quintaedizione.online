package parsers

import (
	"testing"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

func TestAnimaliStrategy(t *testing.T) {
	content := []string{
		"# Animali",
		"",
		"## Cavallo da Guerra",
		"",
		"*Bestia Grande, Non Allineato*",
		"",
		"- **Classe Armatura:** 11",
		"- **Punti Ferita:** 19 (3d10 + 3)",
		"- **Velocità:** 18 m",
		"",
		"| Caratteristica | Valore | Modificatore | Tiro Salvezza |",
		"|----------------|--------|--------------|---------------|",
		"| FOR | 18 | +4 | +4 |",
		"| DES | 12 | +1 | +1 |",
		"| COS | 13 | +1 | +1 |",
		"| INT | 2 | -4 | -4 |",
		"| SAG | 12 | +1 | +1 |",
		"| CAR | 7 | -2 | -2 |",
		"",
		"- **Sensi:** Percezione Passiva 11",
		"- **GS:** 1/2 (PE 100; PB +2)",
		"",
		"### Tratti",
		"",
		"***Carica Travolgente.*** Se il cavallo si muove di almeno 6 m in linea retta.",
		"",
		"### Azioni",
		"",
		"***Zoccoli.*** Attacco con Arma da Mischia: +6 al tiro per colpire.",
	}

	strategy := NewAnimaliStrategy()
	context := NewParsingContext("animali.md", "ita")

	entities, err := strategy.Parse(content, context)
	if err != nil {
		t.Fatalf("Failed to parse animali: %v", err)
	}

	if len(entities) != 1 {
		t.Fatalf("Expected 1 animale, got %d", len(entities))
	}

	animale, ok := entities[0].(*domain.Animale)
	if !ok {
		t.Fatal("Entity is not an Animale")
	}

	if animale.Nome != "Cavallo da Guerra" {
		t.Errorf("Expected nome 'Cavallo da Guerra', got '%s'", animale.Nome)
	}

	if animale.EntityType() != "animale" {
		t.Errorf("Expected entity type 'animale', got '%s'", animale.EntityType())
	}

	if animale.Taglia != domain.TagliaGrande {
		t.Errorf("Expected taglia Grande, got %s", animale.Taglia)
	}

	if animale.Tipo != domain.TipoAnimaleBestia {
		t.Errorf("Expected tipo Bestia, got %s", animale.Tipo)
	}
}

func TestAnimaliStrategy_EmptyContent(t *testing.T) {
	strategy := NewAnimaliStrategy()
	context := NewParsingContext("animali.md", "ita")

	entities, err := strategy.Parse([]string{}, context)
	if err == nil {
		t.Error("Expected error for empty content")
	}
	if len(entities) != 0 {
		t.Error("Expected no entities for empty content")
	}
}

func TestAnimaliStrategy_Interface(t *testing.T) {
	strategy := NewAnimaliStrategy()

	if strategy.ContentType() != ContentTypeAnimali {
		t.Errorf("Expected content type %s, got %s", ContentTypeAnimali, strategy.ContentType())
	}

	if strategy.Name() == "" {
		t.Error("Strategy name should not be empty")
	}

	if strategy.Description() == "" {
		t.Error("Strategy description should not be empty")
	}
}
