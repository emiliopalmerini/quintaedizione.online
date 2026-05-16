package encounter

import (
	"log/slog"
	"os"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/encounter"
	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/infrastructure/persistence/memory"
)

func TestService_CalculateXP2024(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repo := memory.NewEncounterRepository()
	service := NewService(logger, repo)

	tests := []struct {
		name        string
		request     CalculateXPRequest
		expectedXP  int
		expectError bool
	}{
		{
			name: "2024 single character moderate difficulty",
			request: CalculateXPRequest{
				Ruleset:         "2024",
				PartyMode:       "same",
				Difficulty:      "Moderate",
				CharacterLevels: []int{5},
			},
			expectedXP:  750,
			expectError: false,
		},
		{
			name: "2024 party of 4 level 3 characters high difficulty",
			request: CalculateXPRequest{
				Ruleset:         "2024",
				PartyMode:       "same",
				Difficulty:      "High",
				CharacterLevels: []int{3, 3, 3, 3},
			},
			expectedXP:  1600, // 4 * 400
			expectError: false,
		},
		{
			name: "2024 mixed level party low difficulty",
			request: CalculateXPRequest{
				Ruleset:         "2024",
				PartyMode:       "different",
				Difficulty:      "Low",
				CharacterLevels: []int{1, 5, 10, 20},
			},
			expectedXP:  8550, // 50 + 500 + 1600 + 6400
			expectError: false,
		},
		{
			name: "invalid ruleset",
			request: CalculateXPRequest{
				Ruleset:         "invalid",
				PartyMode:       "same",
				Difficulty:      "Moderate",
				CharacterLevels: []int{5},
			},
			expectError: true,
		},
		{
			name: "invalid difficulty for 2024",
			request: CalculateXPRequest{
				Ruleset:         "2024",
				PartyMode:       "same",
				Difficulty:      "Facile",
				CharacterLevels: []int{5},
			},
			expectError: true,
		},
		{
			name: "invalid character level",
			request: CalculateXPRequest{
				Ruleset:         "2024",
				PartyMode:       "same",
				Difficulty:      "Moderate",
				CharacterLevels: []int{0},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CalculateXP(tt.request)

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

			if result.TotalXP != tt.expectedXP {
				t.Errorf("expected XP %d, got %d", tt.expectedXP, result.TotalXP)
			}

			if result.Ruleset != encounter.Ruleset2024 {
				t.Errorf("expected ruleset %s, got %s", encounter.Ruleset2024, result.Ruleset)
			}

			if result.PartySize != len(tt.request.CharacterLevels) {
				t.Errorf("expected party size %d, got %d", len(tt.request.CharacterLevels), result.PartySize)
			}
		})
	}
}

func TestService_CalculateXP2014(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repo := memory.NewEncounterRepository()
	service := NewService(logger, repo)

	tests := []struct {
		name        string
		request     CalculateXPRequest
		expectedXP  int
		expectError bool
	}{
		{
			name: "2014 single character media difficulty 1 monster",
			request: CalculateXPRequest{
				Ruleset:         "2014",
				PartyMode:       "same",
				Difficulty:      "Media",
				CharacterLevels: []int{5},
				NumMonsters:     1,
			},
			expectedXP:  500,
			expectError: false,
		},
		{
			name: "2014 party of 4 level 5 characters media difficulty 2 monsters",
			request: CalculateXPRequest{
				Ruleset:         "2014",
				PartyMode:       "same",
				Difficulty:      "Media",
				CharacterLevels: []int{5, 5, 5, 5},
				NumMonsters:     2,
			},
			expectedXP:  2000, // 4 * 500, cart cost carries the monster-count multiplier
			expectError: false,
		},
		{
			name: "2014 mixed level party difficile difficulty 5 monsters",
			request: CalculateXPRequest{
				Ruleset:         "2014",
				PartyMode:       "different",
				Difficulty:      "Difficile",
				CharacterLevels: []int{3, 5, 7},
				NumMonsters:     5,
			},
			expectedXP:  2075, // 225 + 750 + 1100
			expectError: false,
		},
		{
			name: "invalid difficulty for 2014",
			request: CalculateXPRequest{
				Ruleset:         "2014",
				PartyMode:       "same",
				Difficulty:      "Low",
				CharacterLevels: []int{5},
				NumMonsters:     1,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CalculateXP(tt.request)

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

			if result.TotalXP != tt.expectedXP {
				t.Errorf("expected XP %d, got %d", tt.expectedXP, result.TotalXP)
			}

			if result.Ruleset != encounter.Ruleset2014 {
				t.Errorf("expected ruleset %s, got %s", encounter.Ruleset2014, result.Ruleset)
			}

			// For 2014, we should have a calculated difficulty
			if result.CalculatedDifficulty2014 == "" {
				t.Errorf("expected calculated difficulty for 2014 ruleset, got empty string")
			}
		})
	}
}

func TestService_CalculateXP2014BudgetDoesNotApplyMonsterMultiplier(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repo := memory.NewEncounterRepository()
	service := NewService(logger, repo)

	result, err := service.CalculateXP(CalculateXPRequest{
		Ruleset:         "2014",
		PartyMode:       "same",
		Difficulty:      "Media",
		CharacterLevels: []int{5, 5, 5, 5},
		NumMonsters:     2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	const want = 2000 // 4 level-5 medium thresholds; cart cost carries the multiplier.
	if result.TotalXP != want {
		t.Fatalf("TotalXP = %d, want %d", result.TotalXP, want)
	}
}

func TestService_GetAvailableDifficulties(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repo := memory.NewEncounterRepository()
	service := NewService(logger, repo)

	tests := []struct {
		name        string
		ruleset     string
		expected    []string
		expectError bool
	}{
		{
			name:        "2024 difficulties",
			ruleset:     "2024",
			expected:    []string{"Low", "Moderate", "High"},
			expectError: false,
		},
		{
			name:        "2014 difficulties",
			ruleset:     "2014",
			expected:    []string{"Facile", "Media", "Difficile", "Letale"},
			expectError: false,
		},
		{
			name:        "invalid ruleset",
			ruleset:     "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetAvailableDifficulties(tt.ruleset)

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

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d difficulties, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("expected difficulty %s at index %d, got %s", expected, i, result[i])
				}
			}
		})
	}
}

func TestService_ValidatePartyComposition(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repo := memory.NewEncounterRepository()
	service := NewService(logger, repo)

	tests := []struct {
		name        string
		partyMode   string
		levels      []int
		sameLevel   int
		sameCount   int
		expectError bool
	}{
		{
			name:        "valid same mode",
			partyMode:   "same",
			levels:      nil,
			sameLevel:   5,
			sameCount:   4,
			expectError: false,
		},
		{
			name:        "valid different mode",
			partyMode:   "different",
			levels:      []int{1, 5, 10, 20},
			sameLevel:   0,
			sameCount:   0,
			expectError: false,
		},
		{
			name:        "invalid party mode",
			partyMode:   "invalid",
			levels:      nil,
			sameLevel:   5,
			sameCount:   4,
			expectError: true,
		},
		{
			name:        "same mode with invalid level",
			partyMode:   "same",
			levels:      nil,
			sameLevel:   0,
			sameCount:   4,
			expectError: true,
		},
		{
			name:        "same mode with invalid count",
			partyMode:   "same",
			levels:      nil,
			sameLevel:   5,
			sameCount:   0,
			expectError: true,
		},
		{
			name:        "different mode with empty levels",
			partyMode:   "different",
			levels:      []int{},
			sameLevel:   0,
			sameCount:   0,
			expectError: true,
		},
		{
			name:        "different mode with invalid level",
			partyMode:   "different",
			levels:      []int{1, 0, 10},
			sameLevel:   0,
			sameCount:   0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePartyComposition(tt.partyMode, tt.levels, tt.sameLevel, tt.sameCount)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
