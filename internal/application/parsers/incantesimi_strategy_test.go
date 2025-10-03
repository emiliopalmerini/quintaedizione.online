package parsers

import (
	"testing"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

func TestIncantesimiStrategy(t *testing.T) {
	content := []string{
		"# Incantesimi",
		"",
		"## Palla di Fuoco",
		"",
		"*Livello 3 Invocazione (Mago, Stregone)*",
		"",
		"**Tempo di Lancio:** Azione",
		"**Gittata:** 45 m",
		"**Componenti:** V, S, M (una piccola sfera di guano di pipistrello e zolfo)",
		"**Durata:** Istantanea",
		"",
		"Una striscia luminosa si dirige verso un punto.",
		"",
		"## Luce",
		"",
		"*Trucchetto di Invocazione (Bardo, Chierico, Mago, Stregone)*",
		"",
		"**Tempo di Lancio:** Azione",
		"**Gittata:** Contatto",
		"**Componenti:** V, M (una lucciola o muschio fosforescente)",
		"**Durata:** 1 ora",
		"",
		"Tocchi un oggetto che emette luce.",
	}

	strategy := NewIncantesimiStrategy()
	context := NewParsingContext("incantesimi.md", "ita")

	entities, err := strategy.Parse(content, context)
	if err != nil {
		t.Fatalf("Failed to parse incantesimi: %v", err)
	}

	if len(entities) != 2 {
		t.Fatalf("Expected 2 incantesimi, got %d", len(entities))
	}

	// Test first spell (3rd level)
	pallaDiFuoco, ok := entities[0].(*domain.Incantesimo)
	if !ok {
		t.Fatal("First entity is not an Incantesimo")
	}

	if pallaDiFuoco.Nome != "Palla di Fuoco" {
		t.Errorf("Expected nome 'Palla di Fuoco', got '%s'", pallaDiFuoco.Nome)
	}

	if pallaDiFuoco.EntityType() != "incantesimo" {
		t.Errorf("Expected entity type 'incantesimo', got '%s'", pallaDiFuoco.EntityType())
	}

	if pallaDiFuoco.Livello != 3 {
		t.Errorf("Expected livello 3, got %d", pallaDiFuoco.Livello)
	}

	if pallaDiFuoco.Scuola != "invocazione" {
		t.Errorf("Expected scuola 'invocazione', got '%s'", pallaDiFuoco.Scuola)
	}

	// Test second spell (cantrip)
	luce, ok := entities[1].(*domain.Incantesimo)
	if !ok {
		t.Fatal("Second entity is not an Incantesimo")
	}

	if luce.Livello != 0 {
		t.Errorf("Expected livello 0 for cantrip, got %d", luce.Livello)
	}
}

func TestIncantesimiStrategy_EmptyContent(t *testing.T) {
	strategy := NewIncantesimiStrategy()
	context := NewParsingContext("incantesimi.md", "ita")

	entities, err := strategy.Parse([]string{}, context)
	if err == nil {
		t.Error("Expected error for empty content")
	}
	if len(entities) != 0 {
		t.Error("Expected no entities for empty content")
	}
}

func TestIncantesimiStrategy_Interface(t *testing.T) {
	strategy := NewIncantesimiStrategy()

	if strategy.ContentType() != ContentTypeIncantesimi {
		t.Errorf("Expected content type %s, got %s", ContentTypeIncantesimi, strategy.ContentType())
	}

	if strategy.Name() == "" {
		t.Error("Strategy name should not be empty")
	}

	if strategy.Description() == "" {
		t.Error("Strategy description should not be empty")
	}
}
