package templates

import (
	"bytes"
	"strings"
	"testing"
)

func TestHomeUsesNativeEncounterForm(t *testing.T) {
	data := HomeData{
		Editions:    []EditionOption{{ShortName: "5.5e", Ruleset: "2024", IsDefault: true}},
		Ruleset:     "2024",
		Party:       []int{3, 3, 3, 3},
		SourceShort: "5.5e",
		Cart:        []CartSeed{{ID: "goblin", Source: "5.5e", Quantity: 3}},
	}
	var rendered bytes.Buffer
	if err := Home(data).Render(t.Context(), &rendered); err != nil {
		t.Fatalf("render home: %v", err)
	}
	body := rendered.String()
	for _, expected := range []string{
		`<h1`,
		`action="/combattimenti"`,
		`method="post"`,
		`type="submit"`,
		`name="monsters[]" value="goblin@5.5e:3"`,
		`for="character-level-1"`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("expected home to contain %q", expected)
		}
	}
}
