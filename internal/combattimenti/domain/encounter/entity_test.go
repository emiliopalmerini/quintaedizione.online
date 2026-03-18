package encounter

import (
	"testing"
)

func TestNewParty(t *testing.T) {
	tests := []struct {
		name        string
		levels      []int
		expectError bool
	}{
		{
			name:        "valid party with single character",
			levels:      []int{5},
			expectError: false,
		},
		{
			name:        "valid party with multiple characters",
			levels:      []int{1, 5, 10, 20},
			expectError: false,
		},
		{
			name:        "empty party should fail",
			levels:      []int{},
			expectError: true,
		},
		{
			name:        "party with invalid level should fail",
			levels:      []int{0, 5},
			expectError: true,
		},
		{
			name:        "party with level too high should fail",
			levels:      []int{5, 21},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			party, err := NewParty(tt.levels)

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

			if party.Size() != len(tt.levels) {
				t.Errorf("expected party size %d, got %d", len(tt.levels), party.Size())
			}

			levels := party.Levels()
			for i, expectedLevel := range tt.levels {
				if levels[i] != expectedLevel {
					t.Errorf("expected level %d at index %d, got %d", expectedLevel, i, levels[i])
				}
			}
		})
	}
}

func TestNewCharacter(t *testing.T) {
	tests := []struct {
		name        string
		level       int
		expectError bool
	}{
		{name: "level 1", level: 1, expectError: false},
		{name: "level 10", level: 10, expectError: false},
		{name: "level 20", level: 20, expectError: false},
		{name: "level 0", level: 0, expectError: true},
		{name: "level -1", level: -1, expectError: true},
		{name: "level 21", level: 21, expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			char, err := NewCharacter(tt.level)

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

			if char.Level != tt.level {
				t.Errorf("expected level %d, got %d", tt.level, char.Level)
			}
		})
	}
}

func TestPartyAverageLevel(t *testing.T) {
	tests := []struct {
		name        string
		levels      []int
		expectedAvg float64
	}{
		{
			name:        "single character",
			levels:      []int{10},
			expectedAvg: 10.0,
		},
		{
			name:        "multiple same level characters",
			levels:      []int{5, 5, 5, 5},
			expectedAvg: 5.0,
		},
		{
			name:        "mixed level characters",
			levels:      []int{1, 5, 10, 20},
			expectedAvg: 9.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			party, err := NewParty(tt.levels)
			if err != nil {
				t.Errorf("unexpected error creating party: %v", err)
				return
			}

			avg := party.AverageLevel()
			if avg != tt.expectedAvg {
				t.Errorf("expected average level %f, got %f", tt.expectedAvg, avg)
			}
		})
	}
}

func TestNewEncounter(t *testing.T) {
	party, err := NewParty([]int{5, 5, 5, 5})
	if err != nil {
		t.Fatalf("unexpected error creating party: %v", err)
	}

	ruleset := Ruleset2024
	difficulty := DifficultyModerate

	encounter := NewEncounter("test-id", party, ruleset, difficulty)

	if encounter.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got %s", encounter.ID)
	}

	if encounter.Party.Size() != 4 {
		t.Errorf("expected party size 4, got %d", encounter.Party.Size())
	}

	if encounter.Ruleset != ruleset {
		t.Errorf("expected ruleset %s, got %s", ruleset, encounter.Ruleset)
	}

	if encounter.Difficulty != difficulty {
		t.Errorf("expected difficulty %s, got %s", difficulty, encounter.Difficulty)
	}
}

func TestEncounterToResult(t *testing.T) {
	party, err := NewParty([]int{1, 5, 10})
	if err != nil {
		t.Fatalf("unexpected error creating party: %v", err)
	}

	encounter := NewEncounter("test-id", party, Ruleset2024, DifficultyHigh)
	encounter.TotalXP = 1500

	result := encounter.ToResult()

	if result.Ruleset != Ruleset2024 {
		t.Errorf("expected ruleset %s, got %s", Ruleset2024, result.Ruleset)
	}

	if result.TotalXP != 1500 {
		t.Errorf("expected total XP 1500, got %d", result.TotalXP)
	}

	if result.PartySize != 3 {
		t.Errorf("expected party size 3, got %d", result.PartySize)
	}

	expectedLevels := []int{1, 5, 10}
	for i, level := range result.CharacterLevels {
		if level != expectedLevels[i] {
			t.Errorf("expected level %d at index %d, got %d", expectedLevels[i], i, level)
		}
	}
}
