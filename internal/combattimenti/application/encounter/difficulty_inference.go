package encounter

// TierThreshold is one difficulty tier of the active ruleset, with the XP cost
// at which the encounter is considered to hit that tier for the current party.
// Label is the localized display name; Value is the canonical ruleset key
// ("Low"/"Moderate"/"High" for 2024; "Facile"/"Media"/"Difficile"/"Letale" for
// 2014).
type TierThreshold struct {
	Label string
	Value string
	XP    int
}

// InferredTier is the inference output: which tier the cart's effective cost
// currently falls into, plus overflow/underflow flags so the UI can render
// edge buckets ("Banale", "Mortale", "Letale+") and overspend cues.
type InferredTier struct {
	Label      string
	Value      string
	IsBelowMin bool // effective cost is below the lowest tier threshold
	IsAboveMax bool // effective cost exceeds the highest tier threshold
}

// Sentinel tier values for empty cart, under-threshold ("Banale"), and
// over-max ("Mortale" for 3-tier rulesets like 2024, "Letale+" for 4-tier
// rulesets like 2014). The Value field uses stable canonical keys so CSS
// classes and equality checks don't depend on localized labels.
const (
	tierEmptyValue      = "Empty"
	tierEmptyLabel      = "Cart vuoto"
	tierTrivialValue    = "Trivial"
	tierTrivialLabel    = "Banale"
	tierDeadlyValue     = "Deadly"
	tierDeadlyLabel     = "Mortale"
	tierLethalPlusValue = "LethalPlus"
	tierLethalPlusLabel = "Letale+"
)

// InferDifficulty buckets an effective XP cost into one of the supplied
// thresholds. Thresholds are assumed to be ordered from lowest XP to highest;
// callers (the handler) build them via buildDifficultyTiers, which preserves
// ruleset order.
//
// Rules:
//   - effectiveCost <= 0 or len(thresholds) == 0 → empty bucket (no badge).
//   - effectiveCost < thresholds[0].XP → "Banale" (below-min overflow).
//   - thresholds[i].XP <= effectiveCost < thresholds[i+1].XP → tier i.
//   - effectiveCost >= last threshold → that last tier, unless effectiveCost
//     strictly exceeds it: 3-tier → "Mortale"; 4-tier (2014 already has
//     Letale at top) → "Letale+".
func InferDifficulty(thresholds []TierThreshold, effectiveCost int) InferredTier {
	if effectiveCost <= 0 || len(thresholds) == 0 {
		return InferredTier{Label: tierEmptyLabel, Value: tierEmptyValue}
	}

	if effectiveCost < thresholds[0].XP {
		return InferredTier{Label: tierTrivialLabel, Value: tierTrivialValue, IsBelowMin: true}
	}

	last := thresholds[len(thresholds)-1]
	if effectiveCost > last.XP {
		// Above the top tier: pick an overflow label sized to the ruleset.
		// 4-tier rulesets (2014) already include "Letale" as the top tier, so
		// the overflow becomes "Letale+"; 3-tier rulesets (2024) go to
		// "Mortale".
		if len(thresholds) >= 4 {
			return InferredTier{Label: tierLethalPlusLabel, Value: tierLethalPlusValue, IsAboveMax: true}
		}
		return InferredTier{Label: tierDeadlyLabel, Value: tierDeadlyValue, IsAboveMax: true}
	}

	// Highest tier whose threshold <= effectiveCost. Walk from the top down so
	// we naturally pick the right bucket without off-by-one comparisons.
	for i := len(thresholds) - 1; i >= 0; i-- {
		if effectiveCost >= thresholds[i].XP {
			return InferredTier{Label: thresholds[i].Label, Value: thresholds[i].Value}
		}
	}

	// Unreachable given the guards above; defensive fallback.
	return InferredTier{Label: tierEmptyLabel, Value: tierEmptyValue}
}
