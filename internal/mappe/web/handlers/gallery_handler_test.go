package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	maps "github.com/emiliopalmerini/quintaedizione.online/internal/mappe/domain"
)

type stubMapRepo struct{}

func (stubMapRepo) FindAll() []maps.Mappa {
	return []maps.Mappa{{Slug: "abbazia", Nome: "Abbazia", Immagine: "abbazia.webp"}}
}

type recordingMapRepo struct {
	filters maps.SearchFilters
	results []maps.Mappa
	total   int
}

func (r *recordingMapRepo) FindAll() []maps.Mappa                { return r.results }
func (r *recordingMapRepo) FindBySlug(string) (maps.Mappa, bool) { return maps.Mappa{}, false }
func (r *recordingMapRepo) Search(filters maps.SearchFilters) ([]maps.Mappa, int) {
	r.filters = filters
	return r.results, r.total
}
func (r *recordingMapRepo) Tags() []string { return []string{"dungeon", "rovine"} }

func (stubMapRepo) FindBySlug(slug string) (maps.Mappa, bool) {
	return maps.Mappa{}, false
}

func (stubMapRepo) Search(filters maps.SearchFilters) ([]maps.Mappa, int) {
	return []maps.Mappa{{Slug: "abbazia", Nome: "Abbazia", Immagine: "abbazia.webp"}}, 1
}

func (stubMapRepo) Tags() []string {
	return []string{"dungeon"}
}

func TestHandleGallery_BoostedNavigationRendersFullPage(t *testing.T) {
	handler := NewGalleryHandler(stubMapRepo{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := httptest.NewRequest(http.MethodGet, "/mappe/", nil)
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Target", "body")
	rec := httptest.NewRecorder()

	handler.HandleGallery(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "site-nav") {
		t.Fatalf("boosted navigation should render full layout, got fragment:\n%s", body)
	}
}

func TestHandleGallery_FilterRequestRendersGridFragment(t *testing.T) {
	handler := NewGalleryHandler(stubMapRepo{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := httptest.NewRequest(http.MethodGet, "/mappe?q=abbazia", nil)
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Target", "mappe-grid")
	rec := httptest.NewRecorder()

	handler.HandleGallery(rec, req)

	body := rec.Body.String()
	if strings.Contains(body, "site-nav") {
		t.Fatalf("grid filter request should render only the grid fragment")
	}
	if !strings.Contains(body, `id="mappe-grid"`) {
		t.Fatalf("grid filter request should include #mappe-grid, got:\n%s", body)
	}
}

func TestHandleGalleryParsesRepeatedTagsAndVariesByHTMXHeaders(t *testing.T) {
	repo := &recordingMapRepo{results: []maps.Mappa{{Slug: "torre", Nome: "Torre", Immagine: "torre.webp"}}, total: 1}
	handler := NewGalleryHandler(repo, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := httptest.NewRequest(http.MethodGet, "/mappe?q=torre&tag=dungeon&tag=rovine&offset=-2", nil)
	rec := httptest.NewRecorder()

	handler.HandleGallery(rec, req)

	if repo.filters.Query != "torre" {
		t.Errorf("expected query torre, got %q", repo.filters.Query)
	}
	if len(repo.filters.Tags) != 2 || repo.filters.Tags[0] != "dungeon" || repo.filters.Tags[1] != "rovine" {
		t.Errorf("expected repeated tags, got %v", repo.filters.Tags)
	}
	if repo.filters.Offset != 0 || repo.filters.Limit != defaultPageSize {
		t.Errorf("expected normalized pagination, got offset=%d limit=%d", repo.filters.Offset, repo.filters.Limit)
	}
	if vary := strings.Join(rec.Header().Values("Vary"), ","); !strings.Contains(vary, "HX-Request") || !strings.Contains(vary, "HX-Target") {
		t.Errorf("expected response to vary on HTMX headers, got %q", vary)
	}
}

func TestHandleGallerySelectsFragmentByTarget(t *testing.T) {
	repo := &recordingMapRepo{results: []maps.Mappa{{Slug: "torre", Nome: "Torre", Immagine: "torre.webp"}}, total: 100}
	handler := NewGalleryHandler(repo, slog.New(slog.NewTextHandler(io.Discard, nil)))
	tests := []struct {
		name       string
		url        string
		target     string
		wantGridID bool
	}{
		{name: "grid at nonzero offset", url: "/mappe?offset=40", target: "mappe-grid", wantGridID: true},
		{name: "cards at zero offset", url: "/mappe", target: "mappe-load-more", wantGridID: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.Header.Set("HX-Request", "true")
			req.Header.Set("HX-Target", tt.target)
			rec := httptest.NewRecorder()

			handler.HandleGallery(rec, req)

			hasGridID := strings.Contains(rec.Body.String(), `id="mappe-grid"`)
			if hasGridID != tt.wantGridID {
				t.Errorf("grid wrapper = %t, want %t; body:\n%s", hasGridID, tt.wantGridID, rec.Body.String())
			}
		})
	}
}
