package monster

// CartEntry is one (Source, ID) monster group in the encounter cart. Quantity
// encodes "three goblins" as a single entry rather than three duplicate rows.
type CartEntry struct {
	ID       string
	Source   string
	Name     string
	CR       string
	UnitXP   int
	Quantity int
}

// Cart holds the monsters chosen for an encounter and prices them against the
// party's budget using a provided multiplier function.
//
// The multiplier callback encapsulates the 2014 count-based multiplier; for
// 2024 rulesets the caller passes a constant 1.0 function.
type Cart struct {
	Entries []CartEntry
}

// Size returns the total monster count (sum of per-entry quantities). This is
// what the 2014 multiplier table is keyed on — not the number of unique chips.
func (c Cart) Size() int {
	total := 0
	for _, e := range c.Entries {
		total += e.Quantity
	}
	return total
}

// Subtotal is the raw sum of (unit XP × quantity) across cart entries.
func (c Cart) Subtotal() int {
	total := 0
	for _, e := range c.Entries {
		total += e.UnitXP * e.Quantity
	}
	return total
}

// EffectiveCost applies the ruleset multiplier (1.0 for 2024, variable for
// 2014 based on cart size) to the raw subtotal.
func (c Cart) EffectiveCost(multiplier func(count int) float64) int {
	if c.Size() == 0 {
		return 0
	}
	m := multiplier(c.Size())
	return int(float64(c.Subtotal()) * m)
}

// Remaining returns budget minus effective cost. Can be negative when the
// cart exceeds the budget; the UI displays this as an overspend.
func (c Cart) Remaining(budget int, multiplier func(count int) float64) int {
	return budget - c.EffectiveCost(multiplier)
}
