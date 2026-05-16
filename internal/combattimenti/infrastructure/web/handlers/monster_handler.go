package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/monster"
	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/infrastructure/web/templates"
	pkgweb "github.com/emiliopalmerini/quintaedizione.online/pkg/web"
)

// MonsterPickerHandler renders the monster picker used in the encounter
// builder. Filtering is by name search, CR range, and creature type.
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
//   - q (optional): name search
//   - min_cr, max_cr (optional, float): CR range; 0 means no bound
//   - type (optional): exact creature type ("Drago", "Umanoide", ...)
func (h *MonsterPickerHandler) Handler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-Request-Id")

	source := r.URL.Query().Get("source")
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
		Limit:  100,
	})
	if err != nil {
		h.logger.Warn("monster search failed", "request_id", requestID, "error", err)
	}

	facets, err := h.reader.Facets(r.Context(), source)
	if err != nil {
		h.logger.Warn("monster facets failed", "request_id", requestID, "error", err)
	}

	data := templates.PickerData{
		Source:       source,
		Query:        query,
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
