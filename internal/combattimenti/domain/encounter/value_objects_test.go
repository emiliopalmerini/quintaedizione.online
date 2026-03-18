package encounter

import (
	"testing"
)

func TestNewRuleset(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Ruleset
		expectError bool
	}{
		{name: "2024 ruleset", input: "2024", expected: Ruleset2024, expectError: false},
		{name: "2014 ruleset", input: "2014", expected: Ruleset2014, expectError: false},
		{name: "lowercase 2024", input: "2024", expected: Ruleset2024, expectError: false},
		{name: "invalid ruleset", input: "invalid", expected: "", expectError: true},
		{name: "empty string", input: "", expected: "", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewRuleset(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRulesetIsValid(t *testing.T) {
	tests := []struct {
		ruleset  Ruleset
		expected bool
	}{
		{Ruleset2024, true},
		{Ruleset2014, true},
		{Ruleset("invalid"), false},
		{Ruleset(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.ruleset), func(t *testing.T) {
			if tt.ruleset.IsValid() != tt.expected {
				t.Errorf("expected IsValid() to return %v for %s", tt.expected, tt.ruleset)
			}
		})
	}
}

func TestNewDifficulty(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		ruleset     Ruleset
		expected    Difficulty
		expectError bool
	}{
		// 2024 difficulties
		{name: "2024 Low", value: "Low", ruleset: Ruleset2024, expected: DifficultyLow, expectError: false},
		{name: "2024 Moderate", value: "Moderate", ruleset: Ruleset2024, expected: DifficultyModerate, expectError: false},
		{name: "2024 High", value: "High", ruleset: Ruleset2024, expected: DifficultyHigh, expectError: false},
		{name: "2024 invalid", value: "Easy", ruleset: Ruleset2024, expected: "", expectError: true},

		// 2014 difficulties
		{name: "2014 Facile", value: "Facile", ruleset: Ruleset2014, expected: DifficultyEasy, expectError: false},
		{name: "2014 Media", value: "Media", ruleset: Ruleset2014, expected: DifficultyMedium, expectError: false},
		{name: "2014 Difficile", value: "Difficile", ruleset: Ruleset2014, expected: DifficultyHard, expectError: false},
		{name: "2014 Letale", value: "Letale", ruleset: Ruleset2014, expected: DifficultyDeadly, expectError: false},
		{name: "2014 invalid", value: "Low", ruleset: Ruleset2014, expected: "", expectError: true},

		// Invalid ruleset
		{name: "invalid ruleset", value: "Low", ruleset: Ruleset("invalid"), expected: "", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewDifficulty(tt.value, tt.ruleset)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDifficultyIsValidFor2024(t *testing.T) {
	tests := []struct {
		difficulty Difficulty
		expected   bool
	}{
		{DifficultyLow, true},
		{DifficultyModerate, true},
		{DifficultyHigh, true},
		{DifficultyEasy, false},
		{DifficultyMedium, false},
		{DifficultyHard, false},
		{DifficultyDeadly, false},
		{Difficulty("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.difficulty), func(t *testing.T) {
			if tt.difficulty.IsValidFor2024() != tt.expected {
				t.Errorf("expected IsValidFor2024() to return %v for %s", tt.expected, tt.difficulty)
			}
		})
	}
}

func TestDifficultyIsValidFor2014(t *testing.T) {
	tests := []struct {
		difficulty Difficulty
		expected   bool
	}{
		{DifficultyEasy, true},
		{DifficultyMedium, true},
		{DifficultyHard, true},
		{DifficultyDeadly, true},
		{DifficultyLow, false},
		{DifficultyModerate, false},
		{DifficultyHigh, false},
		{Difficulty("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.difficulty), func(t *testing.T) {
			if tt.difficulty.IsValidFor2014() != tt.expected {
				t.Errorf("expected IsValidFor2014() to return %v for %s", tt.expected, tt.difficulty)
			}
		})
	}
}

func TestNewLevel(t *testing.T) {
	tests := []struct {
		name        string
		value       int
		expected    Level
		expectError bool
	}{
		{name: "level 1", value: 1, expected: Level(1), expectError: false},
		{name: "level 10", value: 10, expected: Level(10), expectError: false},
		{name: "level 20", value: 20, expected: Level(20), expectError: false},
		{name: "level 0", value: 0, expected: Level(0), expectError: true},
		{name: "level -1", value: -1, expected: Level(0), expectError: true},
		{name: "level 21", value: 21, expected: Level(0), expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewLevel(tt.value)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}

			if result.Int() != tt.value {
				t.Errorf("expected Int() to return %d, got %d", tt.value, result.Int())
			}
		})
	}
}

func TestLevelIsValid(t *testing.T) {
	tests := []struct {
		level    Level
		expected bool
	}{
		{Level(1), true},
		{Level(10), true},
		{Level(20), true},
		{Level(0), false},
		{Level(-1), false},
		{Level(21), false},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.level)), func(t *testing.T) {
			if tt.level.IsValid() != tt.expected {
				t.Errorf("expected IsValid() to return %v for level %d", tt.expected, tt.level)
			}
		})
	}
}

func TestNewPartyMode(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expected    PartyMode
		expectError bool
	}{
		{name: "same mode", value: "same", expected: PartyModeSame, expectError: false},
		{name: "different mode", value: "different", expected: PartyModeDifferent, expectError: false},
		{name: "uppercase same", value: "SAME", expected: PartyModeSame, expectError: false},
		{name: "mixed case different", value: "Different", expected: PartyModeDifferent, expectError: false},
		{name: "invalid mode", value: "invalid", expected: "", expectError: true},
		{name: "empty string", value: "", expected: "", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewPartyMode(tt.value)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestPartyModeIsValid(t *testing.T) {
	tests := []struct {
		mode     PartyMode
		expected bool
	}{
		{PartyModeSame, true},
		{PartyModeDifferent, true},
		{PartyMode("invalid"), false},
		{PartyMode(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			if tt.mode.IsValid() != tt.expected {
				t.Errorf("expected IsValid() to return %v for mode %s", tt.expected, tt.mode)
			}
		})
	}
}
