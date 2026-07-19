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
		Tiers:        []DifficultyTier{{Label: "Moderata", Value: "Moderate", XP: 600}},
		Cart:         CartView{Entries: []monster.CartEntry{{ID: "goblin", Source: "5.5e", Name: "Goblin", Quantity: 2, UnitXP: 50}}, Subtotal: 100, EffectiveCost: 100, Remaining: 500},
		InferredTier: app.InferredTier{Label: "Moderata", Value: "Moderate"}, HasCartItems: true,
	}
	var rendered bytes.Buffer
	if err := Result(data).Render(t.Context(), &rendered); err != nil {
		t.Fatalf("render result: %v", err)
	}
	body := rendered.String()
	for _, expected := range []string{
		`role="status"`,
		`aria-live="polite"`,
		`<dl class="result-cart-totals"`,
		`href="/srd/mostri/5.5e/goblin"`,
		`name="cart_action" value="increment:goblin@5.5e"`,
		`aria-hidden="true"`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("expected result to contain %q", expected)
		}
	}
	if strings.Contains(body, `hx-swap-oob`) || strings.Count(body, `id="monster-picker"`) > 0 {
		t.Error("result must not render an out-of-band monster picker")
	}
}
