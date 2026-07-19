package templates

import (
	"bytes"
	"strings"
	"testing"

	app "github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/application/encounter"
	domain "github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/encounter"
	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/monster"
)

func TestResultUsesAccessibleNativeCart(t *testing.T) {
	data := ResultData{
		Result:       &app.CalculateXPResponse{XPCalculationResult: domain.XPCalculationResult{Ruleset: domain.Ruleset2024, PartySize: 4, CharacterLevels: []int{3, 3, 3, 3}}},
		Tiers:        []DifficultyTier{{Label: "Moderata", Value: "Moderate", XP: 600, Selected: true}},
		Cart:         CartView{Entries: []monster.CartEntry{{ID: "goblin", Source: "5.5e", Name: "Goblin", Quantity: 2, UnitXP: 350}}, Subtotal: 700, EffectiveCost: 700, Budget: 600, Remaining: -100},
		InferredTier: app.InferredTier{Label: "Moderata", Value: "Moderate"}, HasCartItems: true, IsOverspent: true,
	}
	var rendered bytes.Buffer
	if err := Result(data).Render(t.Context(), &rendered); err != nil {
		t.Fatalf("render result: %v", err)
	}
	body := rendered.String()
	for _, expected := range []string{
		`role="status"`,
		`aria-live="polite"`,
		`href="/srd/mostri/5.5e/goblin"`,
		`name="cart_action" value="increment:goblin@5.5e"`,
		`aria-hidden="true"`,
		`4 PG · livello 3`,
		`Difficoltà attuale`,
		`100 PE oltre il target Moderata`,
		`Mostri · 2`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("expected result to contain %q", expected)
		}
	}
	if strings.Contains(body, `hx-swap-oob`) || strings.Count(body, `id="monster-picker"`) > 0 {
		t.Error("result must not render an out-of-band monster picker")
	}
	for _, duplicate := range []string{`Target: Moderata`, `Budget target`, `<dl class="result-cart-totals"`} {
		if strings.Contains(body, duplicate) {
			t.Errorf("2024 result should not repeat %q", duplicate)
		}
	}
}

func TestResultKeepsDistinct2014XPValues(t *testing.T) {
	data := ResultData{
		Result:       &app.CalculateXPResponse{XPCalculationResult: domain.XPCalculationResult{Ruleset: domain.Ruleset2014, PartySize: 4, CharacterLevels: []int{3, 4, 4, 5}}},
		Tiers:        []DifficultyTier{{Label: "Media", Value: "Media", XP: 600, Selected: true}},
		Cart:         CartView{Entries: []monster.CartEntry{{ID: "goblin", Source: "5e", Name: "Goblin", Quantity: 4, UnitXP: 50}}, Subtotal: 200, EffectiveCost: 400, Budget: 600, Remaining: 200, Multiplier: true},
		InferredTier: app.InferredTier{Label: "Facile", Value: "Facile"}, HasCartItems: true,
	}
	var rendered bytes.Buffer
	if err := Result(data).Render(t.Context(), &rendered); err != nil {
		t.Fatalf("render result: %v", err)
	}
	body := rendered.String()
	for _, expected := range []string{`4 PG · livelli 3, 4, 4, 5`, `PE base`, `200`, `PE effettivi`, `400`} {
		if !strings.Contains(body, expected) {
			t.Errorf("expected 2014 result to contain %q", expected)
		}
	}
}
