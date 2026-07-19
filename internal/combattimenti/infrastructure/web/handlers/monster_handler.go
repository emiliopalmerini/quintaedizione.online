package handlers

import (
	"context"
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
	limit := parsePickerLimit(r)

	monsters, total, err := searchPickerMonsters(r.Context(), h.reader, monster.SearchQuery{
		Source: source,
		Query:  query,
		MinCR:  minCR,
		MaxCR:  maxCR,
		Type:   creatureType,
	}, limit)
	errorMessage := ""
	if err != nil {
		h.logger.Warn("monster search failed", "request_id", requestID, "error", err)
		errorMessage = "Non è stato possibile caricare i mostri. Riprova."
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
		TotalMatched: total,
		Limit:        limit,
		Error:        errorMessage,
	}

	pkgweb.SetCacheHeaders(w, 300) // 5 minutes
	pkgweb.RenderTempl(w, r, h.logger, templates.MonsterPicker(data))
}

func searchPickerMonsters(ctx context.Context, reader monster.Reader, query monster.SearchQuery, limit int) ([]monster.Monster, int, error) {
	query.Limit = 10000
	monsters, err := reader.Search(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	total := len(monsters)
	if len(monsters) > limit {
		monsters = monsters[:limit]
	}
	return monsters, total, nil
}

func parsePickerLimit(r *http.Request) int {
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 20 {
		return 20
	}
	return min(limit, 200)
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
