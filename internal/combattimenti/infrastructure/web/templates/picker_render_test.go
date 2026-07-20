package templates

import (
	"bytes"
	"strings"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/monster"
)

func TestMonsterPickerUsesNativeSearchLinksAndActions(t *testing.T) {
	data := PickerData{Source: "5.5e", Limit: 20, TotalMatched: 42, Monsters: []monster.Monster{{ID: "goblin", Source: "5.5e", Name: "Goblin", CR: "1/4", XP: 50}}}
	var rendered bytes.Buffer
	if err := MonsterPicker(data).Render(t.Context(), &rendered); err != nil {
		t.Fatalf("render picker: %v", err)
	}
	body := rendered.String()
	for _, expected := range []string{
		`action="/combattimenti"`,
		`method="get"`,
		`type="submit"`,
		`href="/srd/mostri/5.5e/goblin"`,
		`form="encounter-form"`,
		`name="cart_action" value="add:goblin@5.5e"`,
		`1 di 42 mostri`,
		`Carica altri`,
		`limit=40`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("expected picker to contain %q", expected)
		}
	}
	if strings.Contains(body, "Caricamento scheda") {
		t.Error("picker must not contain a JavaScript-only stat block placeholder")
	}
}

func TestMonsterPickerSwapsOnlyResultsToKeepSearchFocused(t *testing.T) {
	data := PickerData{Source: "5.5e", Limit: 20, TotalMatched: 42, Monsters: []monster.Monster{{ID: "goblin", Source: "5.5e", Name: "Goblin"}}}
	var rendered bytes.Buffer
	if err := MonsterPicker(data).Render(t.Context(), &rendered); err != nil {
		t.Fatalf("render picker: %v", err)
	}

	body := rendered.String()
	if got := strings.Count(body, `hx-target="#monster-picker-results"`); got != 2 {
		t.Errorf("expected search and pagination to target results, got %d occurrences", got)
	}
	if got := strings.Count(body, `hx-select="#monster-picker-results"`); got != 2 {
		t.Errorf("expected search and pagination to select results, got %d occurrences", got)
	}
	if strings.Contains(body, `hx-target="#monster-picker"`) {
		t.Error("picker controls must not be replaced during search")
	}
}
