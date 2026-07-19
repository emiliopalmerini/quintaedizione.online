package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/application/encounter"
	domainenc "github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/encounter"
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

// HomeHandler renders the encounter builder home page.
//
// It decodes the URL share-link state from the querystring (see
// domain/encounter.DecodeURLState), drops cart entries whose source does not
// match the active ruleset's edition (ADR-024), and, when the URL is not the
// all-defaults state, server-side-prerenders the result panel so a shared
// link lands on a fully populated page without a JS round-trip.
func (h *EncounterHandler) HomeHandler(editions []templates.EditionOption) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Cache header set here mirrors what the previous inline handler used.
		pkgweb.SetCacheHeaders(w, 3600) // 1 hour
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		state := domainenc.DecodeURLState(r.URL.Query())

		sourceShort := sourceShortForRuleset(editions, state.Ruleset)
		state = state.WithSource(sourceShort) // drop foreign-source cart entries

		numMonsters := countCartItems(state.Cart)
		homeData := templates.HomeData{
			Editions:    editions,
			Ruleset:     state.Ruleset,
			Party:       state.Party,
			SourceShort: sourceShort,
			Cart:        toCartSeeds(state.Cart),
		}

		// Always prerender so the central picker column and result rail ship
		// populated on first paint, including the default empty-cart state.
		if rd, ok := h.prerenderResult(r, state, sourceShort, numMonsters); ok {
			homeData.Result = rd
			homeData.Picker = rd.Picker
		} else {
			homeData.Picker = h.buildPicker(r, sourceShort)
		}

		if err := templates.Home(homeData).Render(r.Context(), w); err != nil {
			h.logger.Error("Failed to render combattimenti home", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// prerenderResult mirrors the body of CalculateHandler enough to produce a
// ResultData snapshot for the URL-hydrated state. Failures fall back to the
// placeholder result (handled by the caller).
func (h *EncounterHandler) prerenderResult(r *http.Request, state domainenc.URLState, sourceShort string, numMonsters int) (*templates.ResultData, bool) {
	req := encounter.CalculateXPRequest{
		Ruleset:         state.Ruleset,
		PartyMode:       "different",
		CharacterLevels: state.Party,
		NumMonsters:     numMonsters,
	}

	tiers := h.buildDifficultyTiers(req, "home-prerender")
	if len(tiers) == 0 {
		return nil, false
	}
	maxBudget := tiers[len(tiers)-1].XP

	refs := make([]encounter.CartItemRef, 0, len(state.Cart))
	for _, ref := range state.Cart {
		qty := ref.Qty
		if qty < 1 {
			qty = 1
		}
		refs = append(refs, encounter.CartItemRef{ID: ref.ID, Source: ref.Source, Quantity: qty})
	}

	priced, err := h.pricer.Price(r.Context(), encounter.PriceCartRequest{
		Ruleset: state.Ruleset,
		Refs:    refs,
		Budget:  maxBudget,
	})
	if err != nil {
		h.logger.Warn("home prerender: cart pricing failed", "error", err)
		priced = &encounter.PriceCartResponse{Remaining: maxBudget}
	}

	cartView := templates.CartView{
		Entries:       priced.Entries,
		Subtotal:      priced.Subtotal,
		EffectiveCost: priced.EffectiveCost,
		Remaining:     priced.Remaining,
		Multiplier:    state.Ruleset == "2014",
	}
	inferred := encounter.InferDifficulty(toTierThresholds(tiers), priced.EffectiveCost)
	picker := h.buildPicker(r, sourceShort)

	result := &encounter.CalculateXPResponse{
		XPCalculationResult: domainenc.XPCalculationResult{
			Ruleset:         domainenc.Ruleset(state.Ruleset),
			TotalXP:         maxBudget,
			PartySize:       len(state.Party),
			CharacterLevels: state.Party,
		},
	}

	return &templates.ResultData{
		Result:       result,
		Tiers:        tiers,
		Cart:         cartView,
		Picker:       picker,
		InferredTier: inferred,
		HasCartItems: len(priced.Entries) > 0,
		IsOverspent:  priced.EffectiveCost > maxBudget,
	}, true
}

// sourceShortForRuleset finds the short-name of the edition whose ruleset
// matches. Falls back to the default edition's short-name, then to the empty
// string when the editions list is empty.
func sourceShortForRuleset(editions []templates.EditionOption, ruleset string) string {
	for _, ed := range editions {
		if ed.Ruleset == ruleset {
			return ed.ShortName
		}
	}
	for _, ed := range editions {
		if ed.IsDefault {
			return ed.ShortName
		}
	}
	if len(editions) > 0 {
		return editions[0].ShortName
	}
	return ""
}

// toCartSeeds keeps one compact form input per cart line.
func toCartSeeds(refs []domainenc.CartRef) []templates.CartSeed {
	out := make([]templates.CartSeed, 0, len(refs))
	for _, ref := range refs {
		qty := ref.Qty
		if qty < 1 {
			qty = 1
		}
		out = append(out, templates.CartSeed{ID: ref.ID, Source: ref.Source, Quantity: qty})
	}
	return out
}

func countCartItems(refs []domainenc.CartRef) int {
	n := 0
	for _, ref := range refs {
		if ref.Qty < 1 {
			n++
			continue
		}
		n += ref.Qty
	}
	return n
}

// UpdateHandler applies a native form submission and redirects to the
// canonical shareable GET URL. HTMX may enhance this flow, but is not required.
func (h *EncounterHandler) UpdateHandler(editions []templates.EditionOption) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		ruleset := r.FormValue("ruleset")
		if _, err := domainenc.NewRuleset(ruleset); err != nil {
			http.Error(w, "Invalid ruleset", http.StatusBadRequest)
			return
		}

		levels, err := h.parsePartyForm(r)
		if err != nil {
			http.Error(w, "Invalid party composition", http.StatusBadRequest)
			return
		}

		source := sourceShortForRuleset(editions, ruleset)
		refs := filterRefsBySource(encounter.ParseCartRefs(r.Form["monsters[]"]), source)
		refs = applyCartAction(refs, r.FormValue("cart_action"), source)
		state := domainenc.URLState{Ruleset: ruleset, Party: levels, Cart: toURLCart(refs)}
		http.Redirect(w, r, buildStateURL(state), http.StatusSeeOther)
	}
}

func (h *EncounterHandler) parsePartyForm(r *http.Request) ([]int, error) {
	mode := r.FormValue("party_mode")
	if mode == "same" {
		level, err := strconv.Atoi(r.FormValue("level"))
		if err != nil {
			return nil, err
		}
		count, err := strconv.Atoi(r.FormValue("count"))
		if err != nil {
			return nil, err
		}
		if err := h.service.ValidatePartyComposition(mode, nil, level, count); err != nil {
			return nil, err
		}
		levels := make([]int, count)
		for i := range levels {
			levels[i] = level
		}
		return levels, nil
	}
	if mode != "different" || len(r.Form["character_levels"]) == 0 || len(r.Form["character_levels"]) > domainenc.MaxPartySize {
		return nil, fmt.Errorf("invalid party mode")
	}
	levels := make([]int, len(r.Form["character_levels"]))
	for i, value := range r.Form["character_levels"] {
		level, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return nil, err
		}
		levels[i] = level
	}
	if err := h.service.ValidatePartyComposition(mode, levels, 0, 0); err != nil {
		return nil, err
	}
	return levels, nil
}

func applyCartAction(refs []encounter.CartItemRef, action, activeSource string) []encounter.CartItemRef {
	verb, value, ok := strings.Cut(action, ":")
	if !ok {
		return refs
	}
	id, source, ok := strings.Cut(value, "@")
	if !ok || id == "" || source != activeSource {
		return refs
	}
	for i := range refs {
		if refs[i].ID != id || refs[i].Source != source {
			continue
		}
		switch verb {
		case "add", "increment":
			refs[i].Quantity = min(refs[i].Quantity+1, 999)
		case "decrement":
			refs[i].Quantity--
		case "remove":
			refs[i].Quantity = 0
		}
		if refs[i].Quantity == 0 {
			return append(refs[:i], refs[i+1:]...)
		}
		return refs
	}
	if verb == "add" || verb == "increment" {
		return append(refs, encounter.CartItemRef{ID: id, Source: source, Quantity: 1})
	}
	return refs
}

func toURLCart(refs []encounter.CartItemRef) []domainenc.CartRef {
	cart := make([]domainenc.CartRef, 0, len(refs))
	for _, ref := range refs {
		cart = append(cart, domainenc.CartRef{ID: ref.ID, Source: ref.Source, Qty: ref.Quantity})
	}
	return cart
}

func buildStateURL(state domainenc.URLState) string {
	query := state.EncodeQuery()
	if query == "" {
		return "/combattimenti"
	}
	return "/combattimenti?" + query
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

		if err := h.service.ValidatePartyComposition(partyMode, nil, level, count); err != nil {
			h.logger.Error("Invalid party composition", "request_id", requestID, "error", err)
			http.Error(w, "Invalid party composition", http.StatusBadRequest)
			return
		}

		characterLevels = make([]int, count)
		for i := range characterLevels {
			characterLevels[i] = level
		}
	} else if partyMode == "different" {
		// For different levels mode, get all character_levels values
		levelStrs := r.Form["character_levels"] // Get all values as slice
		if len(levelStrs) == 0 {
			h.logger.Error("No character levels provided", "request_id", requestID)
			http.Error(w, "Character levels are required for different mode", http.StatusBadRequest)
			return
		}
		if len(levelStrs) > domainenc.MaxPartySize {
			h.logger.Error("Too many character levels provided", "request_id", requestID, "count", len(levelStrs))
			http.Error(w, "Party size cannot exceed 100 characters", http.StatusBadRequest)
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
		if err := h.service.ValidatePartyComposition(partyMode, characterLevels, 0, 0); err != nil {
			h.logger.Error("Invalid party composition", "request_id", requestID, "error", err)
			http.Error(w, "Invalid party composition", http.StatusBadRequest)
			return
		}
	} else {
		h.logger.Error("Invalid party mode", "request_id", requestID, "party_mode", partyMode)
		http.Error(w, "Invalid party mode", http.StatusBadRequest)
		return
	}

	// Parse cart refs. Total quantity drives the 2014 multiplier so the
	// pricer and tier thresholds stay consistent with what the user picked.
	cartRefs := encounter.ParseCartRefs(r.Form["monsters[]"])
	numMonsters := 0
	for _, ref := range cartRefs {
		numMonsters += ref.Quantity
	}

	req := encounter.CalculateXPRequest{
		Ruleset:         ruleset,
		PartyMode:       partyMode,
		CharacterLevels: characterLevels,
		NumMonsters:     numMonsters,
	}

	tiers := h.buildDifficultyTiers(req, requestID)
	if len(tiers) == 0 {
		http.Error(w, "Invalid ruleset", http.StatusBadRequest)
		return
	}

	// Use the highest tier as the picker's affordability ceiling and the
	// cart pricer's budget — there is no single target difficulty anymore.
	maxBudget := tiers[len(tiers)-1].XP

	// Price the cart. Ruleset mismatch clears cart entries server-side.
	sourceShort := r.FormValue("source_short")
	filteredRefs := filterRefsBySource(cartRefs, sourceShort)
	pricedCart, err := h.pricer.Price(r.Context(), encounter.PriceCartRequest{
		Ruleset: ruleset,
		Refs:    filteredRefs,
		Budget:  maxBudget,
	})
	if err != nil {
		h.logger.Warn("cart pricing failed", "request_id", requestID, "error", err)
		pricedCart = &encounter.PriceCartResponse{Remaining: maxBudget}
	}

	cartView := templates.CartView{
		Entries:       pricedCart.Entries,
		Subtotal:      pricedCart.Subtotal,
		EffectiveCost: pricedCart.EffectiveCost,
		Remaining:     pricedCart.Remaining,
		Multiplier:    ruleset == "2014",
	}
	inferredTier := encounter.InferDifficulty(toTierThresholds(tiers), pricedCart.EffectiveCost)

	picker := h.buildPicker(r, sourceShort)

	result := &encounter.CalculateXPResponse{
		XPCalculationResult: domainenc.XPCalculationResult{
			Ruleset:         domainenc.Ruleset(ruleset),
			TotalXP:         maxBudget,
			PartySize:       len(characterLevels),
			CharacterLevels: characterLevels,
		},
	}

	data := templates.ResultData{
		Result:       result,
		Tiers:        tiers,
		Cart:         cartView,
		Picker:       picker,
		InferredTier: inferredTier,
		HasCartItems: len(pricedCart.Entries) > 0,
		IsOverspent:  pricedCart.EffectiveCost > maxBudget,
	}

	// Round-trip the active state into the browser URL so refresh/share
	// preserves everything. HTMX's HX-Push-Url uses an absolute path; build
	// it from the form values we just consumed (post-filtering for ruleset
	// mismatch, which would otherwise leave stale cart entries in the URL).
	pushURL := buildShareURL(ruleset, characterLevels, filteredRefs)
	w.Header().Set("HX-Push-Url", pushURL)

	pkgweb.RenderTempl(w, r, h.logger, templates.Result(data))
	_ = err
}

// toTierThresholds adapts the view-layer tier list to the inference helper's
// minimal shape so the application package stays free of view-model types.
func toTierThresholds(tiers []templates.DifficultyTier) []encounter.TierThreshold {
	out := make([]encounter.TierThreshold, 0, len(tiers))
	for _, t := range tiers {
		out = append(out, encounter.TierThreshold{Label: t.Label, Value: t.Value, XP: t.XP})
	}
	return out
}

// buildShareURL turns the calculate POST's effective state into the absolute
// URL path the browser should display afterwards. Cart refs already exclude
// foreign-source entries (see filterRefsBySource), so the URL is consistent
// with what the user actually got priced.
func buildShareURL(ruleset string, levels []int, refs []encounter.CartItemRef) string {
	cart := make([]domainenc.CartRef, 0, len(refs))
	for _, ref := range refs {
		qty := ref.Quantity
		if qty < 1 {
			qty = 1
		}
		cart = append(cart, domainenc.CartRef{ID: ref.ID, Source: ref.Source, Qty: qty})
	}
	state := domainenc.URLState{
		Ruleset: ruleset,
		Party:   levels,
		Cart:    cart,
	}
	q := state.EncodeQuery()
	// The only mount point for this handler is /combattimenti; HX-Push-Url
	// must be the absolute browser URL, not a fragment.
	base := "/combattimenti"
	if q == "" {
		return base
	}
	return base + "?" + q
}

// buildPicker renders the monster-picker panel for the given source.
func (h *EncounterHandler) buildPicker(r *http.Request, source string) templates.PickerData {
	query := r.URL.Query().Get("q")
	minCR := parseFloatParam(r, "min_cr")
	maxCR := parseFloatParam(r, "max_cr")
	creatureType := r.URL.Query().Get("type")
	monsters, err := h.reader.Search(r.Context(), monster.SearchQuery{
		Source: source,
		Query:  query,
		MinCR:  minCR,
		MaxCR:  maxCR,
		Type:   creatureType,
		Limit:  20,
	})
	if err != nil {
		h.logger.Warn("picker search failed", "error", err)
	}

	// Facets reflect the source's full corpus so dropdowns stay populated
	// even when the current filter combo yields an empty list.
	facets, err := h.reader.Facets(r.Context(), source)
	if err != nil {
		h.logger.Warn("picker facets failed", "error", err)
	}

	return templates.PickerData{
		Source:       source,
		Query:        query,
		MinCR:        minCR,
		MaxCR:        maxCR,
		Type:         creatureType,
		Types:        facets.Types,
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

	mode, err := domainenc.NewPartyMode(partyMode)
	if err != nil {
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
		PartyMode:    mode.String(),
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

// GetDifficultiesHandler returns available difficulties for a ruleset
func (h *EncounterHandler) GetDifficultiesHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-Request-Id")

	ruleset := r.URL.Query().Get("ruleset")
	if ruleset == "" {
		http.Error(w, "Ruleset parameter is required", http.StatusBadRequest)
		return
	}

	difficulties := h.queryHandler.GetDifficultyOptions(ruleset)
	if len(difficulties) == 0 {
		http.Error(w, "Invalid ruleset", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(difficulties); err != nil {
		h.logger.Error("Failed to encode difficulties response", "request_id", requestID, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
