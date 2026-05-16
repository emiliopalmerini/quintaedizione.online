package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/monster"
	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/infrastructure/web/templates"
	pkgweb "github.com/emiliopalmerini/quintaedizione.online/pkg/web"
)

// MonsterPickerHandler renders the affordability-aware monster picker
// embedded in the encounter result panel.
type MonsterPickerHandler struct {
	reader monster.Reader
	logger *slog.Logger
}

func NewMonsterPickerHandler(reader monster.Reader, logger *slog.Logger) *MonsterPickerHandler {
	return &MonsterPickerHandler{reader: reader, logger: logger}
}

// Handler serves GET /combattimenti/monsters.
// Query params:
//   - source (required): edition short name, e.g. "5.5e"
//   - budget (optional, int): wallet size; default 0
//   - q (optional): name search
//   - only_afford (optional, "1"): when present, hide monsters with XP > budget
//   - min_cr, max_cr (optional, float): CR range; 0 means no bound
//   - type (optional): exact creature type ("Drago", "Umanoide", ...)
func (h *MonsterPickerHandler) Handler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-Request-Id")

	source := r.URL.Query().Get("source")
	query := r.URL.Query().Get("q")
	budget := pkgweb.ParseIntParam(r, "budget", 0)
	onlyAfford := r.URL.Query().Get("only_afford") == "1"
	minCR := parseFloatParam(r, "min_cr")
	maxCR := parseFloatParam(r, "max_cr")
	creatureType := r.URL.Query().Get("type")

	maxXP := budget
	if !onlyAfford {
		maxXP = 0
	}

	monsters, err := h.reader.Search(r.Context(), monster.SearchQuery{
		Source:     source,
		Query:      query,
		MinCR:      minCR,
		MaxCR:      maxCR,
		Type:       creatureType,
		MaxXP:      maxXP,
		OnlyAfford: onlyAfford,
		Limit:      100,
	})
	if err != nil {
		h.logger.Warn("monster search failed", "request_id", requestID, "error", err)
	}

	// Facets always reflect the full source corpus, not the filtered subset,
	// so users can broaden their search instead of staring at an empty dropdown.
	facets, err := h.reader.Facets(r.Context(), source)
	if err != nil {
		h.logger.Warn("monster facets failed", "request_id", requestID, "error", err)
	}

	data := templates.PickerData{
		Source:       source,
		Query:        query,
		Budget:       budget,
		MaxXP:        maxXP,
		OnlyAfford:   onlyAfford,
		MinCR:        minCR,
		MaxCR:        maxCR,
		Type:         creatureType,
		Types:        facets.Types,
		Monsters:     monsters,
		TotalMatched: len(monsters),
	}

	pkgweb.SetCacheHeaders(w, 300) // 5 minutes
	pkgweb.RenderTempl(w, r, h.logger, templates.MonsterPicker(data))
}

// parseFloatParam parses a float query parameter, returning 0 on missing/invalid.
// 0 is the convention for "no bound" in the CR-range filter.
func parseFloatParam(r *http.Request, key string) float64 {
	v := r.URL.Query().Get(key)
	if v == "" {
		return 0
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0
	}
	return f
}
