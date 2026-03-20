package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/application/encounter"
	monsterApp "github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/application/monster"
	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/infrastructure/web/templates"
	pkgweb "github.com/emiliopalmerini/quintaedizione.online/pkg/web"
)

// EncounterHandler handles HTTP requests for encounter-related operations
type EncounterHandler struct {
	service        *encounter.Service
	queryHandler   *encounter.QueryHandler
	monsterService *monsterApp.Service
	logger         *slog.Logger
}

// NewEncounterHandler creates a new encounter HTTP handler
func NewEncounterHandler(service *encounter.Service, queryHandler *encounter.QueryHandler, monsterService *monsterApp.Service, logger *slog.Logger) *EncounterHandler {
	return &EncounterHandler{
		service:        service,
		queryHandler:   queryHandler,
		monsterService: monsterService,
		logger:         logger,
	}
}

// Form data structures for HTTP requests
type PartyInputSameForm struct {
	Level int `json:"level" validate:"required,min=1,max=20"`
	Count int `json:"count" validate:"required,min=1,max=100"`
}

type PartyInputDifferentForm struct {
	CharacterLevels []int `json:"character_levels" validate:"required,min=1,dive,min=1,max=20"`
}

type Calculate2024Form struct {
	Ruleset         string `json:"ruleset" validate:"required,eq=2024"`
	PartyMode       string `json:"party_mode" validate:"required,oneof=same different"`
	Difficulty2024  string `json:"difficulty_2024" validate:"required"`
	CharacterLevels []int  `json:"character_levels" validate:"required,min=1"`
}

type Calculate2014Form struct {
	Ruleset         string `json:"ruleset" validate:"required,eq=2014"`
	PartyMode       string `json:"party_mode" validate:"required,oneof=same different"`
	Difficulty2014  string `json:"difficulty_2014" validate:"required"`
	NumMonsters     int    `json:"num_monsters_2014" validate:"min=0"`
	CharacterLevels []int  `json:"character_levels" validate:"required,min=1"`
}

// CalculateHandler handles XP calculation requests
func (h *EncounterHandler) CalculateHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-Request-Id")

	if err := r.ParseForm(); err != nil {
		h.logger.Error("Failed to parse form", "request_id", requestID, "error", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Extract form data
	ruleset := r.FormValue("ruleset")
	partyMode := r.FormValue("party_mode")

	// Parse character levels based on party mode
	var characterLevels []int
	var err error

	if partyMode == "same" {
		level, err := strconv.Atoi(r.FormValue("level"))
		if err != nil {
			h.logger.Error("Invalid level", "request_id", requestID, "error", err)
			http.Error(w, "Invalid character level", http.StatusBadRequest)
			return
		}

		count, err := strconv.Atoi(r.FormValue("count"))
		if err != nil {
			h.logger.Error("Invalid count", "request_id", requestID, "error", err)
			http.Error(w, "Invalid character count", http.StatusBadRequest)
			return
		}

		characterLevels = make([]int, count)
		for i := range characterLevels {
			characterLevels[i] = level
		}
	} else {
		// For different levels mode, get all character_levels values
		levelStrs := r.Form["character_levels"] // Get all values as slice
		if len(levelStrs) == 0 {
			h.logger.Error("No character levels provided", "request_id", requestID)
			http.Error(w, "Character levels are required for different mode", http.StatusBadRequest)
			return
		}

		characterLevels = make([]int, len(levelStrs))
		for i, levelStr := range levelStrs {
			level, err := strconv.Atoi(strings.TrimSpace(levelStr))
			if err != nil {
				h.logger.Error("Invalid character level", "request_id", requestID, "level", levelStr, "error", err)
				http.Error(w, fmt.Sprintf("Invalid character level '%s'", levelStr), http.StatusBadRequest)
				return
			}
			characterLevels[i] = level
		}
	}

	// Create service request
	req := encounter.CalculateXPRequest{
		Ruleset:         ruleset,
		PartyMode:       partyMode,
		CharacterLevels: characterLevels,
	}

	// Add ruleset-specific fields
	switch ruleset {
	case "2024":
		req.Difficulty = r.FormValue("difficulty_2024")
	case "2014":
		req.Difficulty = r.FormValue("difficulty_2014")
		if numMonstersStr := r.FormValue("num_monsters_2014"); numMonstersStr != "" {
			req.NumMonsters, err = strconv.Atoi(numMonstersStr)
			if err != nil {
				h.logger.Error("Invalid number of monsters", "request_id", requestID, "error", err)
				http.Error(w, "Invalid number of monsters", http.StatusBadRequest)
				return
			}
		}
	default:
		http.Error(w, "Invalid ruleset", http.StatusBadRequest)
		return
	}

	// Calculate XP
	result, err := h.service.CalculateXP(req)
	if err != nil {
		h.logger.Error("XP calculation failed", "request_id", requestID, "error", err)
		http.Error(w, fmt.Sprintf("Calculation error: %v", err), http.StatusBadRequest)
		return
	}

	// Calculate all difficulty tiers for visual comparison
	tiers := h.buildDifficultyTiers(req, requestID)

	// Return HTML response for HTMX
	// Include monster browser only on the first render (when header is absent)
	includeMonsterBrowser := r.Header.Get("X-Monster-Browser-Loaded") == ""

	data := templates.ResultData{
		Result:                result,
		Tiers:                 tiers,
		IncludeMonsterBrowser: includeMonsterBrowser,
		Facets: templates.MonsterFacets{
			Types: h.monsterService.AvailableTypes(),
			Sizes: h.monsterService.AvailableSizes(),
			CRs:   h.monsterService.AvailableCRs(),
		},
	}
	pkgweb.RenderTempl(w, r, h.logger, templates.Result(data))
}

// PartyInputHandler handles party input form requests
func (h *EncounterHandler) PartyInputHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-Request-Id")

	partyMode := r.URL.Query().Get("party_mode")
	if partyMode == "" {
		partyMode = "same" // default
	}

	// Validate party mode
	if err := h.service.ValidatePartyComposition(partyMode, nil, 1, 1); err != nil {
		h.logger.Error("Invalid party mode", "request_id", requestID, "party_mode", partyMode, "error", err)
		http.Error(w, "Invalid party mode", http.StatusBadRequest)
		return
	}

	// Get level options for the form
	levelOptions := h.queryHandler.GetLevelOptions()

	// Create response data
	responseData := struct {
		PartyMode    string                  `json:"party_mode"`
		LevelOptions []encounter.LevelOption `json:"level_options"`
	}{
		PartyMode:    partyMode,
		LevelOptions: levelOptions,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responseData); err != nil {
		h.logger.Error("Failed to encode party input response", "request_id", requestID, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// difficultyLabel maps difficulty values to Italian display labels.
var difficultyLabel = map[string]string{
	"Low":       "Bassa",
	"Moderate":  "Moderata",
	"High":      "Alta",
	"Facile":    "Facile",
	"Media":     "Media",
	"Difficile": "Difficile",
	"Letale":    "Letale",
}

// buildDifficultyTiers calculates XP for all difficulty levels of the current ruleset.
func (h *EncounterHandler) buildDifficultyTiers(req encounter.CalculateXPRequest, requestID string) []templates.DifficultyTier {
	var difficulties []string
	switch req.Ruleset {
	case "2024":
		difficulties = []string{"Low", "Moderate", "High"}
	case "2014":
		difficulties = []string{"Facile", "Media", "Difficile", "Letale"}
	default:
		return nil
	}

	tiers := make([]templates.DifficultyTier, 0, len(difficulties))
	for _, diff := range difficulties {
		tierReq := req
		tierReq.Difficulty = diff
		tierResult, err := h.service.CalculateXP(tierReq)
		if err != nil {
			h.logger.Warn("Failed to calculate tier XP", "request_id", requestID, "difficulty", diff, "error", err)
			continue
		}
		label := difficultyLabel[diff]
		if label == "" {
			label = diff
		}
		tiers = append(tiers, templates.DifficultyTier{
			Label:    label,
			Value:    diff,
			XP:       tierResult.TotalXP,
			Selected: diff == req.Difficulty,
		})
	}
	return tiers
}

// Helper function to parse character levels from comma-separated string
func (h *EncounterHandler) parseCharacterLevels(levelsStr string) ([]int, error) {
	if levelsStr == "" {
		return nil, fmt.Errorf("character levels cannot be empty")
	}

	levelStrs := strings.Split(levelsStr, ",")
	levels := make([]int, len(levelStrs))

	for i, levelStr := range levelStrs {
		level, err := strconv.Atoi(strings.TrimSpace(levelStr))
		if err != nil {
			return nil, fmt.Errorf("invalid level '%s': %w", levelStr, err)
		}
		levels[i] = level
	}

	return levels, nil
}

// GetDifficultiesHandler returns available difficulties for a ruleset
func (h *EncounterHandler) GetDifficultiesHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-Request-Id")

	ruleset := r.URL.Query().Get("ruleset")
	if ruleset == "" {
		http.Error(w, "Ruleset parameter is required", http.StatusBadRequest)
		return
	}

	difficulties := h.queryHandler.GetDifficultyOptions(ruleset)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(difficulties); err != nil {
		h.logger.Error("Failed to encode difficulties response", "request_id", requestID, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// sanitizeMonsterName sanitizes the monster name for URL safety
func sanitizeMonsterName(name string) string {
	// Trim whitespace
	name = strings.TrimSpace(name)

	// Limit to 100 characters
	if len(name) > 100 {
		name = name[:100]
	}

	// Remove control characters
	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1 // Remove control characters
		}
		return r
	}, name)

	return strings.TrimSpace(name)
}
