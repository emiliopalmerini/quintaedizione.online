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

func TestHandleHomeRendersNativeSearchForm(t *testing.T) {
	handler := newTestGeneratorHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/generatori/?q=patroni", nil)
	rec := httptest.NewRecorder()

	handler.handleHome(rec, req)

	body := rec.Body.String()
	for _, expected := range []string{
		`action="/generatori/"`,
		`method="get"`,
		`role="search"`,
		`type="search"`,
		`type="submit"`,
		`value="patroni"`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("expected native search form to contain %q, got:\n%s", expected, body)
		}
	}
	if strings.Contains(body, `hx-trigger="keyup`) {
		t.Error("generator search should be enhanced on submit, not every keypress")
	}
	if vary := strings.Join(rec.Header().Values("Vary"), ","); !strings.Contains(vary, "HX-Request") || !strings.Contains(vary, "HX-Target") {
		t.Errorf("expected response to vary on HTMX headers, got %q", vary)
	}
}

func TestHandleRollNegotiatesFullPageAndFragment(t *testing.T) {
	handler := newTestGeneratorHandler(t)
	tests := []struct {
		name         string
		hxRequest    string
		hxTarget     string
		wantFullPage bool
	}{
		{name: "native post", wantFullPage: true},
		{name: "boosted post", hxRequest: "true", hxTarget: "body", wantFullPage: true},
		{name: "roll result", hxRequest: "true", hxTarget: "roll-result", wantFullPage: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/generatori/patroni/roll", nil)
			req.SetPathValue("slug", "patroni")
			if tt.hxRequest != "" {
				req.Header.Set("HX-Request", tt.hxRequest)
				req.Header.Set("HX-Target", tt.hxTarget)
			}
			rec := httptest.NewRecorder()

			handler.handleRoll(rec, req)

			body := rec.Body.String()
			hasLayout := strings.Contains(body, "<!doctype html>") && strings.Contains(body, "site-nav")
			if hasLayout != tt.wantFullPage {
				t.Errorf("full page = %t, want %t; body:\n%s", hasLayout, tt.wantFullPage, body)
			}
			if !strings.Contains(body, "Uno") && !strings.Contains(body, "Due") {
				t.Errorf("expected rolled result, got:\n%s", body)
			}
			if vary := strings.Join(rec.Header().Values("Vary"), ","); !strings.Contains(vary, "HX-Request") || !strings.Contains(vary, "HX-Target") {
				t.Errorf("expected response to vary on HTMX headers, got %q", vary)
			}
		})
	}
}
