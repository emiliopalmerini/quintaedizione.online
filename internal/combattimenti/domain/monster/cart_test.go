package monster

import "testing"

func constantMultiplier(v float64) func(int) float64 {
	return func(int) float64 { return v }
}

func stepMultiplier(pairs ...struct {
	maxCount   int
	multiplier float64
}) func(int) float64 {
	return func(count int) float64 {
		for _, p := range pairs {
			if count <= p.maxCount {
				return p.multiplier
			}
		}
		if len(pairs) == 0 {
			return 1
		}
		return pairs[len(pairs)-1].multiplier
	}
}

func TestCart_EmptySubtotalIsZero(t *testing.T) {
	c := Cart{}
	if got := c.Subtotal(); got != 0 {
		t.Fatalf("empty subtotal = %d, want 0", got)
	}
}

func TestCart_SubtotalSumsUnitXP(t *testing.T) {
	c := Cart{Entries: []CartEntry{
		{UnitXP: 100},
		{UnitXP: 450},
		{UnitXP: 50},
	}}
	if got, want := c.Subtotal(), 600; got != want {
		t.Fatalf("subtotal = %d, want %d", got, want)
	}
}

func TestCart_EffectiveCost2024IsSubtotal(t *testing.T) {
	c := Cart{Entries: []CartEntry{{UnitXP: 100}, {UnitXP: 200}}}
	if got, want := c.EffectiveCost(constantMultiplier(1.0)), 300; got != want {
		t.Fatalf("effective = %d, want %d", got, want)
	}
}

func TestCart_EffectiveCost2014AppliesMultiplier(t *testing.T) {
	c := Cart{Entries: []CartEntry{{UnitXP: 100}, {UnitXP: 200}}}
	mult := stepMultiplier(
		struct {
			maxCount   int
			multiplier float64
		}{1, 1.0},
		struct {
			maxCount   int
			multiplier float64
		}{2, 1.5},
	)
	if got, want := c.EffectiveCost(mult), 450; got != want {
		t.Fatalf("effective = %d, want %d", got, want)
	}
}

func TestCart_EffectiveCost_EmptyCartIsZero(t *testing.T) {
	c := Cart{}
	if got := c.EffectiveCost(constantMultiplier(5.0)); got != 0 {
		t.Fatalf("empty effective = %d, want 0", got)
	}
}

func TestCart_Remaining_NegativeWhenOverspent(t *testing.T) {
	c := Cart{Entries: []CartEntry{{UnitXP: 500}, {UnitXP: 500}}}
	if got, want := c.Remaining(400, constantMultiplier(1.0)), -600; got != want {
		t.Fatalf("remaining = %d, want %d", got, want)
	}
}
