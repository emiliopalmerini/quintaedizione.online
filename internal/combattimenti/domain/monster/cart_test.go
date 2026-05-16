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
		{UnitXP: 100, Quantity: 1},
		{UnitXP: 450, Quantity: 1},
		{UnitXP: 50, Quantity: 1},
	}}
	if got, want := c.Subtotal(), 600; got != want {
		t.Fatalf("subtotal = %d, want %d", got, want)
	}
}

func TestCart_SubtotalMultipliesUnitXPByQuantity(t *testing.T) {
	c := Cart{Entries: []CartEntry{
		{UnitXP: 50, Quantity: 3},  // 3x Goblin = 150
		{UnitXP: 100, Quantity: 2}, // 2x Orc = 200
	}}
	if got, want := c.Subtotal(), 350; got != want {
		t.Fatalf("subtotal = %d, want %d", got, want)
	}
}

func TestCart_SizeSumsQuantities(t *testing.T) {
	c := Cart{Entries: []CartEntry{
		{UnitXP: 50, Quantity: 3},
		{UnitXP: 100, Quantity: 2},
	}}
	if got, want := c.Size(), 5; got != want {
		t.Fatalf("size = %d, want %d", got, want)
	}
}

func TestCart_EffectiveCost2024IsSubtotal(t *testing.T) {
	c := Cart{Entries: []CartEntry{{UnitXP: 100, Quantity: 1}, {UnitXP: 200, Quantity: 1}}}
	if got, want := c.EffectiveCost(constantMultiplier(1.0)), 300; got != want {
		t.Fatalf("effective = %d, want %d", got, want)
	}
}

func TestCart_EffectiveCost2014AppliesMultiplier(t *testing.T) {
	c := Cart{Entries: []CartEntry{{UnitXP: 100, Quantity: 1}, {UnitXP: 200, Quantity: 1}}}
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

// With 2x Goblin + 1x Orc the multiplier must see total count (3), not unique
// chip count (2). At 3 monsters the table jumps to ×2.0.
func TestCart_EffectiveCost2014MultiplierUsesSummedCount(t *testing.T) {
	c := Cart{Entries: []CartEntry{
		{UnitXP: 50, Quantity: 2},  // 2x Goblin = 100
		{UnitXP: 100, Quantity: 1}, // 1x Orc = 100
	}}
	mult := stepMultiplier(
		struct {
			maxCount   int
			multiplier float64
		}{1, 1.0},
		struct {
			maxCount   int
			multiplier float64
		}{2, 1.5},
		struct {
			maxCount   int
			multiplier float64
		}{3, 2.0},
	)
	// subtotal=200, 3 monsters → ×2.0 → 400
	if got, want := c.EffectiveCost(mult), 400; got != want {
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
	c := Cart{Entries: []CartEntry{{UnitXP: 500, Quantity: 1}, {UnitXP: 500, Quantity: 1}}}
	if got, want := c.Remaining(400, constantMultiplier(1.0)), -600; got != want {
		t.Fatalf("remaining = %d, want %d", got, want)
	}
}
