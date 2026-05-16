package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/application/encounter"
	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/monster"
	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/infrastructure/web/templates"
	pkgweb "github.com/emiliopalmerini/quintaedizione.online/pkg/web"
)

// EncounterHandler handles HTTP requests for encounter-related operations
type EncounterHandler struct {
	service      *encounter.Service
	queryHandler *encounter.QueryHandler
	pricer       *encounter.CartPricer
	reader       monster.Reader
	logger       *slog.Logger
}

// NewEncounterHandler creates a new encounter HTTP handler
func NewEncounterHandler(
	service *encounter.Service,
	queryHandler *encounter.QueryHandler,
	pricer *encounter.CartPricer,
	reader monster.Reader,
	logger *slog.Logger,
) *EncounterHandler {
	return &EncounterHandler{
		service:      service,
		queryHandler: queryHandler,
		pricer:       pricer,
		reader:       reader,
		logger:       logger,
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

	// Parse cart refs (may be empty). Cart drives num_monsters_2014 when
	// non-empty so the budget stays consistent with what the user picked.
	// num_monsters_2014 must be the total monster count (sum of quantities),
	// not the unique chip count, so the 2014 multiplier lookup matches the
	// pricer's effective-cost calculation.
	cartRefs := encounter.ParseCartRefs(r.Form["monsters[]"])
	if ruleset == "2014" && len(cartRefs) > 0 {
		total := 0
		for _, ref := range cartRefs {
			total += ref.Quantity
		}
		req.NumMonsters = total
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

	// Price the cart. Ruleset mismatch clears cart entries server-side.
	sourceShort := r.FormValue("source_short")
	filteredRefs := filterRefsBySource(cartRefs, sourceShort)
	pricedCart, err := h.pricer.Price(r.Context(), encounter.PriceCartRequest{
		Ruleset: ruleset,
		Refs:    filteredRefs,
		Budget:  result.TotalXP,
	})
	if err != nil {
		h.logger.Warn("cart pricing failed", "request_id", requestID, "error", err)
		pricedCart = &encounter.PriceCartResponse{Remaining: result.TotalXP}
	}

	cartView := templates.CartView{
		Entries:       pricedCart.Entries,
		Subtotal:      pricedCart.Subtotal,
		EffectiveCost: pricedCart.EffectiveCost,
		Remaining:     pricedCart.Remaining,
		Multiplier:    ruleset == "2014",
	}

	picker := h.buildPicker(r, sourceShort, result.TotalXP)

	data := templates.ResultData{
		Result: result,
		Tiers:  tiers,
		Cart:   cartView,
		Picker: picker,
	}
	pkgweb.RenderTempl(w, r, h.logger, templates.Result(data))
}

// buildPicker renders the monster-picker panel with the current budget as the
// default affordability ceiling.
func (h *EncounterHandler) buildPicker(r *http.Request, source string, budget int) templates.PickerData {
	onlyAfford := true
	query := ""
	maxXP := budget

	monsters, err := h.reader.Search(r.Context(), monster.SearchQuery{
		Source:     source,
		Query:      query,
		MaxXP:      maxXP,
		OnlyAfford: onlyAfford,
		Limit:      100,
	})
	if err != nil {
		h.logger.Warn("picker search failed", "error", err)
	}

	return templates.PickerData{
		Source:       source,
		Query:        query,
		Budget:       budget,
		MaxXP:        maxXP,
		OnlyAfford:   onlyAfford,
		Monsters:     monsters,
		TotalMatched: len(monsters),
	}
}

// filterRefsBySource drops cart entries whose source does not match the
// currently active edition. Per ADR-024, switching ruleset/source clears the
// cart because monsters from different editions cannot share a budget.
func filterRefsBySource(refs []encounter.CartItemRef, source string) []encounter.CartItemRef {
	if source == "" {
		return refs
	}
	out := make([]encounter.CartItemRef, 0, len(refs))
	for _, ref := range refs {
		if ref.Source == source {
			out = append(out, ref)
		}
	}
	return out
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
