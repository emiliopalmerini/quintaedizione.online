package encounter

import (
	"fmt"
	"log/slog"

	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/encounter"
)

// Service provides use cases for encounter management
type Service struct {
	logger     *slog.Logger
	repository encounter.Repository
}

// NewService creates a new encounter application service
func NewService(logger *slog.Logger, repository encounter.Repository) *Service {
	return &Service{
		logger:     logger,
		repository: repository,
	}
}

// CalculateXPRequest represents a request to calculate encounter XP
type CalculateXPRequest struct {
	Ruleset         string
	PartyMode       string
	Difficulty      string
	CharacterLevels []int
	NumMonsters     int // Only used for 2014 ruleset
}

// CalculateXPResponse represents the response from XP calculation
type CalculateXPResponse struct {
	encounter.XPCalculationResult
	CalculatedDifficulty2014 string `json:"calculated_difficulty_2014,omitempty"`
}

// CalculateXP calculates encounter XP based on the request parameters
func (s *Service) CalculateXP(req CalculateXPRequest) (*CalculateXPResponse, error) {
	s.logger.Debug("Calculating XP",
		"ruleset", req.Ruleset,
		"difficulty", req.Difficulty,
		"character_levels", req.CharacterLevels,
		"num_monsters", req.NumMonsters,
	)

	// Validate and create value objects
	ruleset, err := encounter.NewRuleset(req.Ruleset)
	if err != nil {
		return nil, fmt.Errorf("invalid ruleset: %w", err)
	}

	difficulty, err := encounter.NewDifficulty(req.Difficulty, ruleset)
	if err != nil {
		return nil, fmt.Errorf("invalid difficulty: %w", err)
	}

	// Create party
	party, err := encounter.NewParty(req.CharacterLevels)
	if err != nil {
		return nil, fmt.Errorf("invalid party: %w", err)
	}

	// Create encounter
	enc := encounter.NewEncounter("temp-id", party, ruleset, difficulty)
	if ruleset == encounter.Ruleset2014 {
		enc.NumMonsters = req.NumMonsters
	}

	// Calculate XP
	if err := enc.CalculateXP(s.repository); err != nil {
		return nil, fmt.Errorf("failed to calculate XP: %w", err)
	}

	// Convert to response
	result := enc.ToResult()
	response := &CalculateXPResponse{
		XPCalculationResult: result,
	}

	// For 2014 ruleset, calculate difficulty based on XP
	if ruleset == encounter.Ruleset2014 {
		calculatedDifficulty, err := s.calculateDifficulty2014(party, enc.TotalXP)
		if err != nil {
			s.logger.Warn("Failed to calculate 2014 difficulty", "error", err)
		} else {
			response.CalculatedDifficulty2014 = calculatedDifficulty.String()
		}
	}

	s.logger.Info("XP calculation completed",
		"total_xp", enc.TotalXP,
		"party_size", party.Size(),
		"ruleset", ruleset,
	)

	return response, nil
}

// calculateDifficulty2014 determines the actual encounter difficulty for 2014 rules
func (s *Service) calculateDifficulty2014(party encounter.Party, totalXP int) (encounter.Difficulty, error) {
	// Calculate thresholds for each difficulty level
	difficulties := s.repository.GetAllDifficultiesFor2014()

	var bestMatch encounter.Difficulty
	minDifference := int(^uint(0) >> 1) // Max int

	for _, diff := range difficulties {
		totalThreshold := 0
		for _, char := range party.Characters {
			threshold, err := s.repository.GetThresholdFor2014(char.Level, diff)
			if err != nil {
				continue
			}
			totalThreshold += threshold
		}

		difference := abs(totalXP - totalThreshold)
		if difference < minDifference {
			minDifference = difference
			bestMatch = diff
		}
	}

	return bestMatch, nil
}

// GetAvailableDifficulties returns available difficulties for the given ruleset
func (s *Service) GetAvailableDifficulties(ruleset string) ([]string, error) {
	rulesetValue, err := encounter.NewRuleset(ruleset)
	if err != nil {
		return nil, fmt.Errorf("invalid ruleset: %w", err)
	}

	var difficulties []encounter.Difficulty
	switch rulesetValue {
	case encounter.Ruleset2024:
		difficulties = s.repository.GetAllDifficultiesFor2024()
	case encounter.Ruleset2014:
		difficulties = s.repository.GetAllDifficultiesFor2014()
	}

	result := make([]string, len(difficulties))
	for i, d := range difficulties {
		result[i] = d.String()
	}

	return result, nil
}

// GetSupportedLevels returns all supported character levels
func (s *Service) GetSupportedLevels() []int {
	return s.repository.GetSupportedLevels()
}

// ValidatePartyComposition validates party composition parameters
func (s *Service) ValidatePartyComposition(partyMode string, levels []int, sameLevel, sameCount int) error {
	mode, err := encounter.NewPartyMode(partyMode)
	if err != nil {
		return fmt.Errorf("invalid party mode: %w", err)
	}

	switch mode {
	case encounter.PartyModeSame:
		if sameLevel < 1 || sameLevel > 20 {
			return fmt.Errorf("character level must be between 1 and 20")
		}
		if sameCount < 1 || sameCount > 100 {
			return fmt.Errorf("party size must be between 1 and 100")
		}
	case encounter.PartyModeDifferent:
		if len(levels) == 0 {
			return fmt.Errorf("at least one character level must be specified")
		}
		if len(levels) > 100 {
			return fmt.Errorf("party size cannot exceed 100 characters")
		}
		for i, level := range levels {
			if level < 1 || level > 20 {
				return fmt.Errorf("character %d level must be between 1 and 20", i+1)
			}
		}
	}

	return nil
}

// Helper function
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
