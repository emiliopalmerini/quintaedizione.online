package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	monsterApp "github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/application/monster"
	monsterDomain "github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/monster"
	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/infrastructure/web/templates"
)

// MonsterHandler handles HTTP requests for monster browsing.
type MonsterHandler struct {
	service *monsterApp.Service
	logger  *slog.Logger
}

// NewMonsterHandler creates a new monster HTTP handler.
func NewMonsterHandler(service *monsterApp.Service, logger *slog.Logger) *MonsterHandler {
	return &MonsterHandler{
		service: service,
		logger:  logger,
	}
}

// SearchHandler handles monster search requests via HTMX.
// GET /api/monsters?max_xp=N&q=search&type=T&size=S&cr_min=X&cr_max=Y
func (h *MonsterHandler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-Request-Id")

	maxXP := 1_000_000
	if v := r.URL.Query().Get("max_xp"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			h.logger.Error("Invalid max_xp", "request_id", requestID, "error", err)
			http.Error(w, "Invalid max_xp parameter", http.StatusBadRequest)
			return
		}
		maxXP = parsed
	}

	filters := monsterDomain.SearchFilters{
		Query: r.URL.Query().Get("q"),
		MaxXP: maxXP,
		Type:  r.URL.Query().Get("type"),
		Size:  r.URL.Query().Get("size"),
		CRMin: r.URL.Query().Get("cr_min"),
		CRMax: r.URL.Query().Get("cr_max"),
	}

	monsters := h.service.SearchMonstersWithFilters(filters)

	w.Header().Set("Content-Type", "text/html")
	if err := templates.MonsterList(monsters, maxXP).Render(r.Context(), w); err != nil {
		h.logger.Error("Failed to render monster list", "request_id", requestID, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
