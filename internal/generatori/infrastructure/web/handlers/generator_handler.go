package handlers

import (
	"log/slog"
	"net/http"

	"github.com/emiliopalmerini/quintaedizione.online/internal/generatori/application"
	"github.com/emiliopalmerini/quintaedizione.online/internal/generatori/infrastructure/web/templates"
	pkgweb "github.com/emiliopalmerini/quintaedizione.online/pkg/web"
)

type GeneratorHandler struct {
	service *application.Service
	logger  *slog.Logger
}

func NewGeneratorHandler(service *application.Service, logger *slog.Logger) *GeneratorHandler {
	return &GeneratorHandler{
		service: service,
		logger:  logger,
	}
}

func (h *GeneratorHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", h.handleHome)
	mux.HandleFunc("GET /{slug}", h.handleGenerator)
	mux.HandleFunc("POST /{slug}/roll", h.handleRoll)
}

func (h *GeneratorHandler) handleHome(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	groups := h.service.SearchGroups(query)

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := templates.GeneratoriGrid(groups, query).Render(r.Context(), w); err != nil {
			h.logger.Error("Failed to render generatori grid", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	pkgweb.RenderTempl(w, r, h.logger, templates.Home(groups, query))
}

func (h *GeneratorHandler) handleGenerator(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	table, ok := h.service.Get(slug)
	if !ok {
		http.NotFound(w, r)
		return
	}

	prev, next := h.service.Neighbors(slug)
	pkgweb.RenderTempl(w, r, h.logger, templates.Generator(table, prev, next))
}

func (h *GeneratorHandler) handleRoll(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	result, err := h.service.Roll(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	pkgweb.RenderTempl(w, r, h.logger, templates.Result(result))
}
