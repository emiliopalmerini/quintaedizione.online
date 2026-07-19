package templates

import (
	"bytes"
	"strings"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/monster"
)

func TestMonsterPickerUsesNativeSearchLinksAndActions(t *testing.T) {
	data := PickerData{Source: "5.5e", Monsters: []monster.Monster{{ID: "goblin", Source: "5.5e", Name: "Goblin", CR: "1/4", XP: 50}}}
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
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("expected picker to contain %q", expected)
		}
	}
	if strings.Contains(body, "Caricamento scheda") {
		t.Error("picker must not contain a JavaScript-only stat block placeholder")
	}
}
