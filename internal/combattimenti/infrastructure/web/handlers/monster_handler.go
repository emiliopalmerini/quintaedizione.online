package handlers

import (
	"log/slog"
	"net/http"

	monsterApp "github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/application/monster"
	monsterDomain "github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/monster"
	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/infrastructure/web/templates"
	pkgweb "github.com/emiliopalmerini/quintaedizione.online/pkg/web"
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
	maxXP := pkgweb.ParseIntParam(r, "max_xp", 1_000_000)

	filters := monsterDomain.SearchFilters{
		Query: r.URL.Query().Get("q"),
		MaxXP: maxXP,
		Type:  r.URL.Query().Get("type"),
		Size:  r.URL.Query().Get("size"),
		CRMin: r.URL.Query().Get("cr_min"),
		CRMax: r.URL.Query().Get("cr_max"),
	}

	monsters := h.service.SearchMonstersWithFilters(filters)

	pkgweb.SetCacheHeaders(w, 1800) // 30 minutes
	pkgweb.RenderTempl(w, r, h.logger, templates.MonsterList(monsters, maxXP))
}
