package encounter

import "testing"

// 2024 tiers (Low/Moderate/High) and 2014 tiers (Facile/Media/Difficile/Letale)
// cover all bucketing edges. Thresholds chosen for readability, not realism.

func tiers2024() []TierThreshold {
	return []TierThreshold{
		{Label: "Bassa", Value: "Low", XP: 100},
		{Label: "Moderata", Value: "Moderate", XP: 200},
		{Label: "Alta", Value: "High", XP: 400},
	}
}

func tiers2014() []TierThreshold {
	return []TierThreshold{
		{Label: "Facile", Value: "Facile", XP: 50},
		{Label: "Media", Value: "Media", XP: 100},
		{Label: "Difficile", Value: "Difficile", XP: 200},
		{Label: "Letale", Value: "Letale", XP: 400},
	}
}

func TestInferDifficulty_2024(t *testing.T) {
	thresholds := tiers2024()
	cases := []struct {
		name      string
		cost      int
		wantLabel string
		wantValue string
		below     bool
		above     bool
	}{
		{"empty cart", 0, "Cart vuoto", "Empty", false, false},
		{"below low", 50, "Banale", "Trivial", true, false},
		{"exactly low", 100, "Bassa", "Low", false, false},
		{"between low and moderate", 150, "Bassa", "Low", false, false},
		{"exactly moderate", 200, "Moderata", "Moderate", false, false},
		{"between moderate and high", 300, "Moderata", "Moderate", false, false},
		{"exactly high", 400, "Alta", "High", false, false},
		{"above high", 600, "Mortale", "Deadly", false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := InferDifficulty(thresholds, c.cost)
			if got.Label != c.wantLabel {
				t.Errorf("Label = %q, want %q", got.Label, c.wantLabel)
			}
			if got.Value != c.wantValue {
				t.Errorf("Value = %q, want %q", got.Value, c.wantValue)
			}
			if got.IsBelowMin != c.below {
				t.Errorf("IsBelowMin = %v, want %v", got.IsBelowMin, c.below)
			}
			if got.IsAboveMax != c.above {
				t.Errorf("IsAboveMax = %v, want %v", got.IsAboveMax, c.above)
			}
		})
	}
}

func TestInferDifficulty_2014(t *testing.T) {
	thresholds := tiers2014()
	cases := []struct {
		name      string
		cost      int
		wantLabel string
		wantValue string
		below     bool
		above     bool
	}{
		{"empty cart", 0, "Cart vuoto", "Empty", false, false},
		{"below facile", 25, "Banale", "Trivial", true, false},
		{"exactly facile", 50, "Facile", "Facile", false, false},
		{"between facile and media", 75, "Facile", "Facile", false, false},
		{"exactly media", 100, "Media", "Media", false, false},
		{"between media and difficile", 150, "Media", "Media", false, false},
		{"exactly difficile", 200, "Difficile", "Difficile", false, false},
		{"between difficile and letale", 300, "Difficile", "Difficile", false, false},
		{"exactly letale", 400, "Letale", "Letale", false, false},
		{"above letale", 600, "Letale+", "LethalPlus", false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := InferDifficulty(thresholds, c.cost)
			if got.Label != c.wantLabel {
				t.Errorf("Label = %q, want %q", got.Label, c.wantLabel)
			}
			if got.Value != c.wantValue {
				t.Errorf("Value = %q, want %q", got.Value, c.wantValue)
			}
			if got.IsBelowMin != c.below {
				t.Errorf("IsBelowMin = %v, want %v", got.IsBelowMin, c.below)
			}
			if got.IsAboveMax != c.above {
				t.Errorf("IsAboveMax = %v, want %v", got.IsAboveMax, c.above)
			}
		})
	}
}

func TestInferDifficulty_NoThresholds(t *testing.T) {
	got := InferDifficulty(nil, 100)
	// With no thresholds we can't classify; treat as empty/unknown.
	if got.Value != "Empty" {
		t.Errorf("Value = %q, want %q", got.Value, "Empty")
	}
}

func TestInferDifficulty_NegativeCost(t *testing.T) {
	// A negative effective cost (shouldn't happen, but be defensive) is treated
	// as empty/unspent so the badge degrades gracefully rather than reading as
	// "Banale".
	got := InferDifficulty(tiers2024(), -10)
	if got.Value != "Empty" {
		t.Errorf("Value = %q, want %q", got.Value, "Empty")
	}
}

func TestInferDifficulty_ThreeTierRulesetAboveMaxLabel(t *testing.T) {
	// 2024-shape thresholds must use the "Mortale" overflow label, not "Letale+".
	got := InferDifficulty(tiers2024(), 5000)
	if got.Label != "Mortale" || got.Value != "Deadly" {
		t.Errorf("got {%q,%q}, want {Mortale,Deadly}", got.Label, got.Value)
	}
	if !got.IsAboveMax {
		t.Errorf("IsAboveMax should be true")
	}
}

func TestInferDifficulty_FourTierRulesetAboveMaxLabel(t *testing.T) {
	// 2014-shape thresholds (4 tiers) use "Letale+" for the overflow bucket.
	got := InferDifficulty(tiers2014(), 5000)
	if got.Label != "Letale+" || got.Value != "LethalPlus" {
		t.Errorf("got {%q,%q}, want {Letale+,LethalPlus}", got.Label, got.Value)
	}
}
