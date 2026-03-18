package encounter

// Repository defines the interface for accessing encounter data
type Repository interface {
	// GetXPFor2024 returns the XP amount for a given level and difficulty in 2024 rules
	GetXPFor2024(level int, difficulty Difficulty) (int, error)

	// GetThresholdFor2014 returns the XP threshold for a given level and difficulty in 2014 rules
	GetThresholdFor2014(level int, difficulty Difficulty) (int, error)

	// GetMultiplierFor2014 returns the encounter multiplier based on number of monsters
	GetMultiplierFor2014(numMonsters int) (float64, error)

	// GetAllDifficultiesFor2024 returns all available difficulties for 2024 ruleset
	GetAllDifficultiesFor2024() []Difficulty

	// GetAllDifficultiesFor2014 returns all available difficulties for 2014 ruleset
	GetAllDifficultiesFor2014() []Difficulty

	// GetSupportedLevels returns all supported character levels
	GetSupportedLevels() []int
}

// MultiplierRange represents a range for encounter multipliers
type MultiplierRange struct {
	MaxMonsters int
	Multiplier  float64
}
