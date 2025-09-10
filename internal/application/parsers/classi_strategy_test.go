package parsers

import (
	"testing"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

func TestClassiStrategy(t *testing.T) {
	content := []string{
		"# Classi",
		"",
		"## Barbaro",
		"",
		"Tabella: Tratti base del Barbaro",
		"",
		"|                               |                                                                                            |",
		"|-------------------------------|--------------------------------------------------------------------------------------------|",
		"| Caratteristica primaria       | Forza                                                                                      |",
		"| Dado Punti Ferita             | D12 per livello da Barbaro                                                                 |",
		"| Tiri salvezza competenti      | Forza e Costituzione                                                                       |",
		"| Abilità competenti            | Scegli 2: Addestrare Animali, Atletica, Intimidire, Natura, Percezione o Sopravvivenza     |",
		"| Armi competenti               | Armi semplici e da guerra                                                                  |",
		"| Armature addestramento        | Armature leggere e medie e Scudi                                                           |",
		"| Equipaggiamento iniziale      | Scegli A o B: (A) Ascia bipenne, 4 Asce da lancio, Zaino da esploratore e 15 mo; oppure (B) 75 mo |",
		"",
		"### Diventare un Barbaro …",
		"",
		"#### Come personaggio di 1° livello",
		"",
		"• Ottieni tutti i tratti della tabella Tratti base del Barbaro.",
		"",
		"• Ottieni i privilegi di classe di 1° livello del Barbaro, elencati nella tabella Privilegi del Barbaro.",
		"",
		"#### Come personaggio multiclasse",
		"",
		"- Ottieni i seguenti tratti dalla tabella Tratti base del Barbaro: Dado Punti Ferita, competenza con le armi da guerra e addestramento con gli Scudi.",
		"- Ottieni i privilegi di classe di 1° livello del Barbaro, elencati nella tabella Privilegi del Barbaro.",
		"",
		"### Privilegi di classe del Barbaro",
		"",
		"Come Barbaro, ottieni i seguenti privilegi di classe quando raggiungi i livelli indicati da Barbaro. Questi privilegi sono elencati nella tabella Privilegi del Barbaro.",
		"",
		"Tabella: Privilegi del Barbaro",
		"",
		"| Livello | Bonus competenza     | Privilegi di classe                       | Ira | Danni da Ira | Maestria nelle armi |",
		"|---------|----------------------|------------------------------------------|-------|----------------|---------------------|",
		"| 1       | +2                   | Ira, Difesa senza armatura, Maestria nelle armi | 2     | +2             | 2                   |",
		"| 2       | +2                   | Senso del pericolo, Attacco sconsiderato | 2     | +2             | 2                   |",
		"| 3       | +2                   | Sottoclasse del Barbaro, Conoscenza primordiale | 3     | +2             | 2                   |",
		"",
		"#### 1° livello: Ira",
		"",
		"Puoi imbeverti di una forza primordiale detta Ira, che ti dona potenza e resistenza straordinarie.",
		"",
		"## Bardo",
		"",
		"Tabella: Tratti base del Bardo",
		"",
		"|                               |                                                                                            |",
		"|-------------------------------|--------------------------------------------------------------------------------------------|",
		"| Caratteristica primaria       | Carisma                                                                                    |",
		"| Dado Punti Ferita             | D8 per livello da Bardo                                                                    |",
		"| Tiri salvezza competenti      | Destrezza e Carisma                                                                        |",
		"| Abilità competenti            | Scegli 3: qualunque abilità                                                               |",
		"| Armi competenti               | Armi semplici, Balestre a mano, Spade lunghe, Stocchi e Spade corte                       |",
		"| Armature addestramento        | Armature leggere                                                                           |",
		"| Equipaggiamento iniziale      | Scegli A o B: (A) Stocco, Balestra a mano con 20 quadrelli, Zaino da intrattenitore e 11 mo; oppure (B) 110 mo |",
	}

	strategy := NewClassiStrategy()
	context := NewParsingContext("classi.md", "ita")

	entities, err := strategy.Parse(content, context)
	if err != nil {
		t.Fatalf("Failed to parse classi: %v", err)
	}

	if len(entities) != 2 {
		t.Fatalf("Expected 2 classi, got %d", len(entities))
	}

	// Test first classe (Barbaro)
	barbaro, ok := entities[0].(*domain.Classe)
	if !ok {
		t.Fatal("First entity is not a Classe")
	}

	if barbaro.Nome != "Barbaro" {
		t.Errorf("Expected name 'Barbaro', got '%s'", barbaro.Nome)
	}

	if barbaro.EntityType() != "classe" {
		t.Errorf("Expected entity type 'classe', got '%s'", barbaro.EntityType())
	}

	// Test that we have parsed the basic attributes
	expectedD12 := domain.NewDado(1, 12, 0)
	if barbaro.DadoVita.Facce != expectedD12.Facce {
		t.Errorf("Expected dado vita with 12 faces, got %v", barbaro.DadoVita.Facce)
	}

	// Test second classe (Bardo)
	bardo, ok := entities[1].(*domain.Classe)
	if !ok {
		t.Fatal("Second entity is not a Classe")
	}

	if bardo.Nome != "Bardo" {
		t.Errorf("Expected name 'Bardo', got '%s'", bardo.Nome)
	}

	expectedD8 := domain.NewDado(1, 8, 0)
	if bardo.DadoVita.Facce != expectedD8.Facce {
		t.Errorf("Expected dado vita with 8 faces, got %v", bardo.DadoVita.Facce)
	}
}

func TestClassiStrategy_EmptyContent(t *testing.T) {
	strategy := NewClassiStrategy()
	context := NewParsingContext("classi.md", "ita")

	entities, err := strategy.Parse([]string{}, context)
	if err == nil {
		t.Error("Expected error for empty content")
	}
	if len(entities) != 0 {
		t.Error("Expected no entities for empty content")
	}
}

func TestClassiStrategy_Interface(t *testing.T) {
	strategy := NewClassiStrategy()

	if strategy.ContentType() != ContentTypeClassi {
		t.Errorf("Expected content type %s, got %s", ContentTypeClassi, strategy.ContentType())
	}

	if strategy.Name() == "" {
		t.Error("Strategy name should not be empty")
	}

	if strategy.Description() == "" {
		t.Error("Strategy description should not be empty")
	}
}