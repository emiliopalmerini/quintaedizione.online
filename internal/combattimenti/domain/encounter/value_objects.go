package encounter

import (
	"errors"
	"strings"
)

// Ruleset represents the D&D ruleset version
type Ruleset string

const (
	Ruleset2024 Ruleset = "2024"
	Ruleset2014 Ruleset = "2014"
)

// NewRuleset creates and validates a new Ruleset
func NewRuleset(value string) (Ruleset, error) {
	switch strings.ToLower(value) {
	case "2024":
		return Ruleset2024, nil
	case "2014":
		return Ruleset2014, nil
	default:
		return "", errors.New("invalid ruleset: must be '2024' or '2014'")
	}
}

// String returns the string representation of the ruleset
func (r Ruleset) String() string {
	return string(r)
}

// IsValid checks if the ruleset is valid
func (r Ruleset) IsValid() bool {
	return r == Ruleset2024 || r == Ruleset2014
}

// Difficulty represents encounter difficulty
type Difficulty string

const (
	// 2024 difficulties
	DifficultyLow      Difficulty = "Low"
	DifficultyModerate Difficulty = "Moderate"
	DifficultyHigh     Difficulty = "High"

	// 2014 difficulties
	DifficultyEasy   Difficulty = "Facile"
	DifficultyMedium Difficulty = "Media"
	DifficultyHard   Difficulty = "Difficile"
	DifficultyDeadly Difficulty = "Letale"
)

// NewDifficulty creates and validates a new Difficulty for the given ruleset
func NewDifficulty(value string, ruleset Ruleset) (Difficulty, error) {
	difficulty := Difficulty(value)

	switch ruleset {
	case Ruleset2024:
		if !difficulty.IsValidFor2024() {
			return "", errors.New("invalid difficulty for 2024 ruleset: must be 'Low', 'Moderate', or 'High'")
		}
	case Ruleset2014:
		if !difficulty.IsValidFor2014() {
			return "", errors.New("invalid difficulty for 2014 ruleset: must be 'Facile', 'Media', 'Difficile', or 'Letale'")
		}
	default:
		return "", errors.New("invalid ruleset")
	}

	return difficulty, nil
}

// String returns the string representation of the difficulty
func (d Difficulty) String() string {
	return string(d)
}

// IsValidFor2024 checks if the difficulty is valid for 2024 ruleset
func (d Difficulty) IsValidFor2024() bool {
	switch d {
	case DifficultyLow, DifficultyModerate, DifficultyHigh:
		return true
	default:
		return false
	}
}

// IsValidFor2014 checks if the difficulty is valid for 2014 ruleset
func (d Difficulty) IsValidFor2014() bool {
	switch d {
	case DifficultyEasy, DifficultyMedium, DifficultyHard, DifficultyDeadly:
		return true
	default:
		return false
	}
}

// Level represents a character level
type Level int

// NewLevel creates and validates a new Level
func NewLevel(value int) (Level, error) {
	if value < 1 || value > 20 {
		return 0, errors.New("level must be between 1 and 20")
	}
	return Level(value), nil
}

// Int returns the integer value of the level
func (l Level) Int() int {
	return int(l)
}

// IsValid checks if the level is within valid range
func (l Level) IsValid() bool {
	return l >= 1 && l <= 20
}

// PartyMode represents how party composition is specified
type PartyMode string

const (
	PartyModeSame      PartyMode = "same"
	PartyModeDifferent PartyMode = "different"
)

// NewPartyMode creates and validates a new PartyMode
func NewPartyMode(value string) (PartyMode, error) {
	mode := PartyMode(strings.ToLower(value))
	switch mode {
	case PartyModeSame, PartyModeDifferent:
		return mode, nil
	default:
		return "", errors.New("invalid party mode: must be 'same' or 'different'")
	}
}

// String returns the string representation of the party mode
func (p PartyMode) String() string {
	return string(p)
}

// IsValid checks if the party mode is valid
func (p PartyMode) IsValid() bool {
	return p == PartyModeSame || p == PartyModeDifferent
}
