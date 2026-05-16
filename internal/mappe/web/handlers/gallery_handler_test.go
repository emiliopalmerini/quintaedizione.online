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
	req := httptest.NewRequest(http.MethodGet, "/mappe/gallery?q=abbazia", nil)
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
