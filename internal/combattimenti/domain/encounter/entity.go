package encounter

import (
	"errors"
	"fmt"
)

// Encounter represents a D&D encounter with XP calculation
type Encounter struct {
	ID          string
	Party       Party
	Ruleset     Ruleset
	Difficulty  Difficulty
	TotalXP     int
	NumMonsters int
}

// Party represents a group of characters
type Party struct {
	Characters []Character
}

// Character represents a single party member
type Character struct {
	Level int
}

// XPCalculationResult represents the result of XP calculation
type XPCalculationResult struct {
	Ruleset                  Ruleset
	TotalXP                  int
	CalculatedDifficulty2014 Difficulty
	PartySize                int
	CharacterLevels          []int
}

// NewParty creates a new party with the given character levels
func NewParty(levels []int) (Party, error) {
	if len(levels) == 0 {
		return Party{}, errors.New("party must have at least one character")
	}

	var characters []Character
	for _, level := range levels {
		char, err := NewCharacter(level)
		if err != nil {
			return Party{}, fmt.Errorf("invalid character level %d: %w", level, err)
		}
		characters = append(characters, char)
	}

	return Party{Characters: characters}, nil
}

// NewCharacter creates a new character with the given level
func NewCharacter(level int) (Character, error) {
	if level < 1 || level > 20 {
		return Character{}, errors.New("character level must be between 1 and 20")
	}
	return Character{Level: level}, nil
}

// Size returns the number of characters in the party
func (p Party) Size() int {
	return len(p.Characters)
}

// Levels returns the levels of all characters in the party
func (p Party) Levels() []int {
	levels := make([]int, len(p.Characters))
	for i, char := range p.Characters {
		levels[i] = char.Level
	}
	return levels
}

// AverageLevel calculates the average level of the party
func (p Party) AverageLevel() float64 {
	if len(p.Characters) == 0 {
		return 0
	}

	total := 0
	for _, char := range p.Characters {
		total += char.Level
	}
	return float64(total) / float64(len(p.Characters))
}

// NewEncounter creates a new encounter
func NewEncounter(id string, party Party, ruleset Ruleset, difficulty Difficulty) *Encounter {
	return &Encounter{
		ID:         id,
		Party:      party,
		Ruleset:    ruleset,
		Difficulty: difficulty,
	}
}

// CalculateXP calculates the total XP for this encounter
func (e *Encounter) CalculateXP(repo Repository) error {
	switch e.Ruleset {
	case Ruleset2024:
		return e.calculateXP2024(repo)
	case Ruleset2014:
		return e.calculateXP2014(repo)
	default:
		return fmt.Errorf("unsupported ruleset: %s", e.Ruleset)
	}
}

func (e *Encounter) calculateXP2024(repo Repository) error {
	totalXP := 0
	for _, char := range e.Party.Characters {
		xp, err := repo.GetXPFor2024(char.Level, e.Difficulty)
		if err != nil {
			return fmt.Errorf("failed to get XP for level %d: %w", char.Level, err)
		}
		totalXP += xp
	}
	e.TotalXP = totalXP
	return nil
}

func (e *Encounter) calculateXP2014(repo Repository) error {
	// For 2014 rules, we need to get thresholds and apply multipliers
	totalThreshold := 0
	for _, char := range e.Party.Characters {
		threshold, err := repo.GetThresholdFor2014(char.Level, e.Difficulty)
		if err != nil {
			return fmt.Errorf("failed to get threshold for level %d: %w", char.Level, err)
		}
		totalThreshold += threshold
	}

	// Apply multiplier based on number of monsters
	multiplier, err := repo.GetMultiplierFor2014(e.NumMonsters)
	if err != nil {
		return fmt.Errorf("failed to get multiplier for %d monsters: %w", e.NumMonsters, err)
	}

	e.TotalXP = int(float64(totalThreshold) * multiplier)
	return nil
}

// ToResult converts the encounter to an XPCalculationResult
func (e *Encounter) ToResult() XPCalculationResult {
	return XPCalculationResult{
		Ruleset:         e.Ruleset,
		TotalXP:         e.TotalXP,
		PartySize:       e.Party.Size(),
		CharacterLevels: e.Party.Levels(),
	}
}
