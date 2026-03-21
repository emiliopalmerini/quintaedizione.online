package handlers

import (
	"log/slog"
	"net/http"

	"github.com/emiliopalmerini/quintaedizione.online/internal/generatori/application"
	"github.com/emiliopalmerini/quintaedizione.online/internal/generatori/infrastructure/web/templates"
)

// GeneratorHandler handles HTTP requests for random generators.
type GeneratorHandler struct {
	service *application.Service
	logger  *slog.Logger
}

// NewGeneratorHandler creates a new generator HTTP handler.
func NewGeneratorHandler(service *application.Service, logger *slog.Logger) *GeneratorHandler {
	return &GeneratorHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers all generator routes on the given mux.
func (h *GeneratorHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", h.handleHome)
	mux.HandleFunc("GET /{slug}", h.handleGenerator)
	mux.HandleFunc("POST /{slug}/roll", h.handleRoll)
}

func (h *GeneratorHandler) handleHome(w http.ResponseWriter, r *http.Request) {
	groups := h.service.ListGroups()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.Home(groups).Render(r.Context(), w); err != nil {
		h.logger.Error("Failed to render generatori home", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *GeneratorHandler) handleGenerator(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	table, ok := h.service.Get(slug)
	if !ok {
		http.NotFound(w, r)
		return
	}

	prev, next := h.service.Neighbors(slug)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.Generator(table, prev, next).Render(r.Context(), w); err != nil {
		h.logger.Error("Failed to render generator", "slug", slug, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *GeneratorHandler) handleRoll(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	result, err := h.service.Roll(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.Result(result).Render(r.Context(), w); err != nil {
		h.logger.Error("Failed to render roll result", "slug", slug, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
