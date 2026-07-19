package encounter

import (
	"fmt"
	"log/slog"

	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/encounter"
)

// QueryHandler handles read-only queries for the UI
type QueryHandler struct {
	logger     *slog.Logger
	repository encounter.Repository
}

// NewQueryHandler creates a new query handler
func NewQueryHandler(logger *slog.Logger, repository encounter.Repository) *QueryHandler {
	return &QueryHandler{
		logger:     logger,
		repository: repository,
	}
}

// RulesetOption represents a ruleset option for UI
type RulesetOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// DifficultyOption represents a difficulty option for UI
type DifficultyOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// LevelOption represents a level option for UI
type LevelOption struct {
	Value int    `json:"value"`
	Label string `json:"label"`
}

// GetRulesetOptions returns available ruleset options
func (q *QueryHandler) GetRulesetOptions() []RulesetOption {
	return []RulesetOption{
		{Value: "2024", Label: "D&D 2024 (One D&D)"},
		{Value: "2014", Label: "D&D 2014 (5th Edition)"},
	}
}

// GetDifficultyOptions returns available difficulty options for a ruleset
func (q *QueryHandler) GetDifficultyOptions(ruleset string) []DifficultyOption {
	rulesetValue, err := encounter.NewRuleset(ruleset)
	if err != nil {
		q.logger.Warn("Invalid ruleset for difficulty options", "ruleset", ruleset, "error", err)
		return []DifficultyOption{}
	}

	switch rulesetValue {
	case encounter.Ruleset2024:
		return []DifficultyOption{
			{Value: "Low", Label: "Low"},
			{Value: "Moderate", Label: "Moderate"},
			{Value: "High", Label: "High"},
		}
	case encounter.Ruleset2014:
		return []DifficultyOption{
			{Value: "Facile", Label: "Facile"},
			{Value: "Media", Label: "Media"},
			{Value: "Difficile", Label: "Difficile"},
			{Value: "Letale", Label: "Letale"},
		}
	default:
		return []DifficultyOption{}
	}
}

// GetLevelOptions returns available character level options
func (q *QueryHandler) GetLevelOptions() []LevelOption {
	levels := q.repository.GetSupportedLevels()
	options := make([]LevelOption, len(levels))

	for i, level := range levels {
		options[i] = LevelOption{
			Value: level,
			Label: levelToLabel(level),
		}
	}

	return options
}

// GetPartyModeOptions returns available party mode options
func (q *QueryHandler) GetPartyModeOptions() []struct {
	Value string `json:"value"`
	Label string `json:"label"`
} {
	return []struct {
		Value string `json:"value"`
		Label string `json:"label"`
	}{
		{Value: "same", Label: "All characters same level"},
		{Value: "different", Label: "Characters different levels"},
	}
}

// GetXPThreshold returns XP threshold for a specific level and difficulty (2014 rules)
func (q *QueryHandler) GetXPThreshold(level int, difficulty string, ruleset string) (int, error) {
	rulesetValue, err := encounter.NewRuleset(ruleset)
	if err != nil {
		return 0, err
	}

	difficultyValue, err := encounter.NewDifficulty(difficulty, rulesetValue)
	if err != nil {
		return 0, err
	}

	switch rulesetValue {
	case encounter.Ruleset2024:
		return q.repository.GetXPFor2024(level, difficultyValue)
	case encounter.Ruleset2014:
		return q.repository.GetThresholdFor2014(level, difficultyValue)
	default:
		return 0, err
	}
}

// GetMultiplierRanges returns encounter multiplier ranges (2014 rules)
func (q *QueryHandler) GetMultiplierRanges() []encounter.MultiplierRange {
	// Return hardcoded ranges for now, could be moved to repository
	return []encounter.MultiplierRange{
		{MaxMonsters: 1, Multiplier: 1.0},
		{MaxMonsters: 2, Multiplier: 1.5},
		{MaxMonsters: 6, Multiplier: 2.0},
		{MaxMonsters: 10, Multiplier: 2.5},
		{MaxMonsters: 14, Multiplier: 3.0},
		{MaxMonsters: int(^uint(0) >> 1), Multiplier: 4.0},
	}
}

// Helper function to convert level to display label
func levelToLabel(level int) string {
	switch level {
	case 1:
		return "1st Level"
	case 2:
		return "2nd Level"
	case 3:
		return "3rd Level"
	default:
		return fmt.Sprintf("%dth Level", level)
	}
}
