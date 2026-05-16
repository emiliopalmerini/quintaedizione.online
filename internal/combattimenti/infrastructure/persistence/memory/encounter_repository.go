package memory

import (
	"fmt"

	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/encounter"
)

// EncounterRepository implements the encounter.Repository interface using in-memory data
type EncounterRepository struct {
	xpData2024       map[int]map[string]int
	xpThresholds2014 map[int]map[string]int
	multiplierRanges []encounter.MultiplierRange
}

// NewEncounterRepository creates a new in-memory encounter repository
func NewEncounterRepository() *EncounterRepository {
	return &EncounterRepository{
		xpData2024: map[int]map[string]int{
			1:  {"Low": 50, "Moderate": 75, "High": 100},
			2:  {"Low": 100, "Moderate": 150, "High": 200},
			3:  {"Low": 150, "Moderate": 225, "High": 400},
			4:  {"Low": 250, "Moderate": 375, "High": 500},
			5:  {"Low": 500, "Moderate": 750, "High": 1100},
			6:  {"Low": 600, "Moderate": 1000, "High": 1400},
			7:  {"Low": 750, "Moderate": 1300, "High": 1700},
			8:  {"Low": 1000, "Moderate": 1700, "High": 2100},
			9:  {"Low": 1300, "Moderate": 2000, "High": 2600},
			10: {"Low": 1600, "Moderate": 2300, "High": 3100},
			11: {"Low": 1900, "Moderate": 2900, "High": 4100},
			12: {"Low": 2200, "Moderate": 3700, "High": 4700},
			13: {"Low": 2600, "Moderate": 4200, "High": 5400},
			14: {"Low": 2900, "Moderate": 4900, "High": 6200},
			15: {"Low": 3300, "Moderate": 5400, "High": 7800},
			16: {"Low": 3800, "Moderate": 6100, "High": 9800},
			17: {"Low": 4500, "Moderate": 7200, "High": 11700},
			18: {"Low": 5000, "Moderate": 8700, "High": 14200},
			19: {"Low": 5500, "Moderate": 10700, "High": 17200},
			20: {"Low": 6400, "Moderate": 13200, "High": 22000},
		},
		xpThresholds2014: map[int]map[string]int{
			1:  {"Facile": 25, "Media": 50, "Difficile": 75, "Letale": 100},
			2:  {"Facile": 50, "Media": 100, "Difficile": 150, "Letale": 200},
			3:  {"Facile": 75, "Media": 150, "Difficile": 225, "Letale": 400},
			4:  {"Facile": 125, "Media": 250, "Difficile": 375, "Letale": 500},
			5:  {"Facile": 250, "Media": 500, "Difficile": 750, "Letale": 1100},
			6:  {"Facile": 300, "Media": 600, "Difficile": 900, "Letale": 1400},
			7:  {"Facile": 350, "Media": 750, "Difficile": 1100, "Letale": 1700},
			8:  {"Facile": 450, "Media": 900, "Difficile": 1400, "Letale": 2100},
			9:  {"Facile": 550, "Media": 1100, "Difficile": 1600, "Letale": 2400},
			10: {"Facile": 600, "Media": 1200, "Difficile": 1900, "Letale": 2800},
			11: {"Facile": 800, "Media": 1600, "Difficile": 2400, "Letale": 3600},
			12: {"Facile": 1000, "Media": 2000, "Difficile": 3000, "Letale": 4500},
			13: {"Facile": 1100, "Media": 2200, "Difficile": 3400, "Letale": 5100},
			14: {"Facile": 1250, "Media": 2500, "Difficile": 3800, "Letale": 5700},
			15: {"Facile": 1400, "Media": 2800, "Difficile": 4300, "Letale": 6400},
			16: {"Facile": 1600, "Media": 3200, "Difficile": 4800, "Letale": 7200},
			17: {"Facile": 2000, "Media": 3900, "Difficile": 5900, "Letale": 8800},
			18: {"Facile": 2100, "Media": 4200, "Difficile": 6300, "Letale": 9500},
			19: {"Facile": 2400, "Media": 4900, "Difficile": 7300, "Letale": 10900},
			20: {"Facile": 2800, "Media": 5700, "Difficile": 8500, "Letale": 12700},
		},
		multiplierRanges: []encounter.MultiplierRange{
			{MaxMonsters: 1, Multiplier: 1.0},
			{MaxMonsters: 2, Multiplier: 1.5},
			{MaxMonsters: 3, Multiplier: 2.0},
			{MaxMonsters: 7, Multiplier: 2.5},
			{MaxMonsters: 11, Multiplier: 3.0},
			{MaxMonsters: 15, Multiplier: 4.0},
			{MaxMonsters: 99, Multiplier: 5.0},
		},
	}
}

// GetXPFor2024 returns the XP amount for a given level and difficulty in 2024 rules
func (r *EncounterRepository) GetXPFor2024(level int, difficulty encounter.Difficulty) (int, error) {
	levelData, exists := r.xpData2024[level]
	if !exists {
		return 0, fmt.Errorf("unsupported character level: %d", level)
	}

	xp, exists := levelData[difficulty.String()]
	if !exists {
		return 0, fmt.Errorf("unsupported difficulty for 2024: %s", difficulty.String())
	}

	return xp, nil
}

// GetThresholdFor2014 returns the XP threshold for a given level and difficulty in 2014 rules
func (r *EncounterRepository) GetThresholdFor2014(level int, difficulty encounter.Difficulty) (int, error) {
	levelData, exists := r.xpThresholds2014[level]
	if !exists {
		return 0, fmt.Errorf("unsupported character level: %d", level)
	}

	threshold, exists := levelData[difficulty.String()]
	if !exists {
		return 0, fmt.Errorf("unsupported difficulty for 2014: %s", difficulty.String())
	}

	return threshold, nil
}

// GetMultiplierFor2014 returns the encounter multiplier based on number of monsters
func (r *EncounterRepository) GetMultiplierFor2014(numMonsters int) (float64, error) {
	if numMonsters < 1 {
		return 0, fmt.Errorf("number of monsters must be at least 1")
	}

	for _, multiplierRange := range r.multiplierRanges {
		if numMonsters <= multiplierRange.MaxMonsters {
			return multiplierRange.Multiplier, nil
		}
	}

	// Default to highest multiplier for very large numbers
	return r.multiplierRanges[len(r.multiplierRanges)-1].Multiplier, nil
}

// GetAllDifficultiesFor2024 returns all available difficulties for 2024 ruleset
func (r *EncounterRepository) GetAllDifficultiesFor2024() []encounter.Difficulty {
	return []encounter.Difficulty{
		encounter.DifficultyLow,
		encounter.DifficultyModerate,
		encounter.DifficultyHigh,
	}
}

// GetAllDifficultiesFor2014 returns all available difficulties for 2014 ruleset
func (r *EncounterRepository) GetAllDifficultiesFor2014() []encounter.Difficulty {
	return []encounter.Difficulty{
		encounter.DifficultyEasy,
		encounter.DifficultyMedium,
		encounter.DifficultyHard,
		encounter.DifficultyDeadly,
	}
}

// GetSupportedLevels returns all supported character levels
func (r *EncounterRepository) GetSupportedLevels() []int {
	levels := make([]int, 0, len(r.xpData2024))
	for level := range r.xpData2024 {
		levels = append(levels, level)
	}

	// Sort levels (simple bubble sort for small dataset)
	for i := 0; i < len(levels)-1; i++ {
		for j := 0; j < len(levels)-i-1; j++ {
			if levels[j] > levels[j+1] {
				levels[j], levels[j+1] = levels[j+1], levels[j]
			}
		}
	}

	return levels
}
