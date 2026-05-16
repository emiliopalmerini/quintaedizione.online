package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/emiliopalmerini/quintaedizione.online/internal/generatori/application"
)

func newTestGeneratorHandler(t *testing.T) *GeneratorHandler {
	t.Helper()

	data := fstest.MapFS{
		"test.json": &fstest.MapFile{Data: []byte(`{
			"id":"patroni",
			"name":"Patroni",
			"description":"Genera un patrono",
			"die":"1D2",
			"order":1,
			"group":"core-adventure",
			"source":{"author":"Test","url":"https://example.com"},
			"items":["Uno","Due"]
		}`)},
	}

	service, err := application.NewService(data)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	return NewGeneratorHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestHandleHome_BoostedNavigationRendersFullPage(t *testing.T) {
	handler := newTestGeneratorHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/generatori/", nil)
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Target", "body")
	rec := httptest.NewRecorder()

	handler.handleHome(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "site-nav") {
		t.Fatalf("boosted navigation should render full layout, got fragment:\n%s", body)
	}
}

func TestHandleHome_SearchRequestRendersGridFragment(t *testing.T) {
	handler := newTestGeneratorHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/generatori/?q=patroni", nil)
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Target", "generatori-grid")
	rec := httptest.NewRecorder()

	handler.handleHome(rec, req)

	body := rec.Body.String()
	if strings.Contains(body, "site-nav") {
		t.Fatalf("search request should render only the generator grid fragment")
	}
	if !strings.Contains(body, `id="generatori-grid"`) {
		t.Fatalf("search request should include #generatori-grid, got:\n%s", body)
	}
}
