package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/maps"
	"github.com/emiliopalmerini/quintaedizione.online/internal/mappe/web/templates"
)

const defaultPageSize = 40

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
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	filters := maps.SearchFilters{
		Query:     r.URL.Query().Get("q"),
		Categoria: r.URL.Query().Get("categoria"),
		Tag:       r.URL.Query().Get("tag"),
		Offset:    offset,
		Limit:     defaultPageSize,
	}

	results, total := h.repo.Search(filters)

	data := maps.GalleryData{
		Mappe:     results,
		Categorie: h.repo.Categorie(),
		Tags:      h.repo.Tags(),
		Query:     filters.Query,
		Categoria: filters.Categoria,
		Tag:       filters.Tag,
		Total:     total,
		Offset:    offset,
		Limit:     defaultPageSize,
		HasMore:   offset+len(results) < total,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if r.Header.Get("HX-Request") == "true" {
		var tmpl func() error
		if offset > 0 {
			tmpl = func() error { return templates.GalleryCards(data).Render(r.Context(), w) }
		} else {
			tmpl = func() error { return templates.GalleryGrid(data).Render(r.Context(), w) }
		}
		if err := tmpl(); err != nil {
			h.logger.Error("Failed to render gallery", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if err := templates.GalleryPage(data).Render(r.Context(), w); err != nil {
		h.logger.Error("Failed to render gallery page", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
