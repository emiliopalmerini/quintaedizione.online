package handlers

import (
	"log/slog"
	"net/http"

	maps "github.com/emiliopalmerini/quintaedizione.online/internal/mappe/domain"
	"github.com/emiliopalmerini/quintaedizione.online/internal/mappe/web/templates"
	pkgweb "github.com/emiliopalmerini/quintaedizione.online/pkg/web"
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

	pkgweb.SetCacheHeaders(w, 14400) // 4 hours
	pkgweb.RenderTempl(w, r, h.logger, templates.DetailPage(m))
}

// HandleGallery renders the full gallery page.
// GET /
func (h *GalleryHandler) HandleGallery(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Vary", "HX-Request")
	w.Header().Add("Vary", "HX-Target")

	offset := pkgweb.ParseIntParam(r, "offset", 0)
	if offset < 0 {
		offset = 0
	}

	activeTags := parseActiveTags(r)

	filters := maps.SearchFilters{
		Query:  r.URL.Query().Get("q"),
		Tags:   activeTags,
		Offset: offset,
		Limit:  defaultPageSize,
	}

	results, total := h.repo.Search(filters)

	data := maps.GalleryData{
		Mappe:      results,
		Tags:       h.repo.Tags(),
		Query:      filters.Query,
		ActiveTags: activeTags,
		Total:      total,
		Offset:     offset,
		Limit:      defaultPageSize,
		HasMore:    offset+len(results) < total,
	}

	pkgweb.SetCacheHeaders(w, 1800) // 30 minutes
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if target := galleryFragmentTarget(r); target != "" {
		component := func() error { return templates.GalleryGrid(data).Render(r.Context(), w) }
		if target == "mappe-load-more" {
			component = func() error { return templates.GalleryCards(data).Render(r.Context(), w) }
		}
		if err := component(); err != nil {
			h.logger.Error("Failed to render gallery", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	pkgweb.RenderTempl(w, r, h.logger, templates.GalleryPage(data))
}

func galleryFragmentTarget(r *http.Request) string {
	if r.Header.Get("HX-Request") != "true" {
		return ""
	}

	switch r.Header.Get("HX-Target") {
	case "mappe-grid", "mappe-load-more":
		return r.Header.Get("HX-Target")
	default:
		return ""
	}
}

func parseActiveTags(r *http.Request) []string {
	var tags []string
	seen := make(map[string]bool)
	for _, value := range r.URL.Query()["tag"] {
		for _, tag := range pkgweb.ParseTags(value) {
			if !seen[tag] {
				tags = append(tags, tag)
				seen[tag] = true
			}
		}
	}
	return tags
}
