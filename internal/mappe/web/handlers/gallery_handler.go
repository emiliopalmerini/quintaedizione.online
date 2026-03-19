package handlers

import (
	"log/slog"
	"net/http"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/maps"
	"github.com/emiliopalmerini/quintaedizione.online/internal/mappe/web/templates"
)

// GalleryHandler handles HTTP requests for the map gallery.
type GalleryHandler struct {
	repo   maps.Repository
	logger *slog.Logger
}

// NewGalleryHandler creates a new gallery HTTP handler.
func NewGalleryHandler(repo maps.Repository, logger *slog.Logger) *GalleryHandler {
	return &GalleryHandler{
		repo:   repo,
		logger: logger,
	}
}

// HandleDetail renders a single map detail page.
// GET /{slug}
func (h *GalleryHandler) HandleDetail(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	m, ok := h.repo.FindBySlug(slug)
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.DetailPage(m).Render(r.Context(), w); err != nil {
		h.logger.Error("Failed to render map detail", "slug", slug, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// HandleGallery renders the full gallery page.
// GET /
func (h *GalleryHandler) HandleGallery(w http.ResponseWriter, r *http.Request) {
	filters := maps.SearchFilters{
		Query:     r.URL.Query().Get("q"),
		Categoria: r.URL.Query().Get("categoria"),
		Tag:       r.URL.Query().Get("tag"),
	}

	results := h.repo.Search(filters)

	data := maps.GalleryData{
		Mappe:     results,
		Categorie: h.repo.Categorie(),
		Tags:      h.repo.Tags(),
		Query:     filters.Query,
		Categoria: filters.Categoria,
		Tag:       filters.Tag,
		Total:     len(results),
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := templates.GalleryGrid(data).Render(r.Context(), w); err != nil {
			h.logger.Error("Failed to render gallery grid", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.GalleryPage(data).Render(r.Context(), w); err != nil {
		h.logger.Error("Failed to render gallery page", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
