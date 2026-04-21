package handlers

import (
	"log/slog"
	"net/http"

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
func (h *MonsterPickerHandler) Handler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-Request-Id")

	source := r.URL.Query().Get("source")
	query := r.URL.Query().Get("q")
	budget := pkgweb.ParseIntParam(r, "budget", 0)
	onlyAfford := r.URL.Query().Get("only_afford") == "1"

	maxXP := budget
	if !onlyAfford {
		maxXP = 0
	}

	monsters, err := h.reader.Search(r.Context(), monster.SearchQuery{
		Source:     source,
		Query:      query,
		MaxXP:      maxXP,
		OnlyAfford: onlyAfford,
		Limit:      100,
	})
	if err != nil {
		h.logger.Warn("monster search failed", "request_id", requestID, "error", err)
	}

	data := templates.PickerData{
		Source:       source,
		Query:        query,
		Budget:       budget,
		MaxXP:        maxXP,
		OnlyAfford:   onlyAfford,
		Monsters:     monsters,
		TotalMatched: len(monsters),
	}

	pkgweb.SetCacheHeaders(w, 300) // 5 minutes
	pkgweb.RenderTempl(w, r, h.logger, templates.MonsterPicker(data))
}
