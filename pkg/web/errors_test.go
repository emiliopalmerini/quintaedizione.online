package web

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestErrorResponder_Respond_MatchesPattern(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	er := &ErrorResponder{
		Logger: logger,
		Mappings: []ErrorMapping{
			{
				Patterns:   []string{"not found"},
				StatusCode: http.StatusNotFound,
				Message:    "Elemento non trovato.",
			},
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	er.Respond(w, r, fmt.Errorf("document not found"), "fallback")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Elemento non trovato.") {
		t.Errorf("expected error message in body, got %q", w.Body.String())
	}
}

func TestErrorResponder_Respond_Fallback(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	er := &ErrorResponder{Logger: logger}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	er.Respond(w, r, fmt.Errorf("something unexpected"), "Errore generico")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Errore generico") {
		t.Errorf("expected fallback message, got %q", w.Body.String())
	}
}

func TestErrorResponder_Respond_HTTPError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	er := &ErrorResponder{Logger: logger}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	er.Respond(w, r, NewHTTPError(http.StatusForbidden, "Accesso negato"), "")

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestErrorResponder_Respond_HTMX(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	er := &ErrorResponder{Logger: logger}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("HX-Request", "true")

	er.Respond(w, r, fmt.Errorf("oops"), "Errore")

	if w.Header().Get("HX-Reswap") != "innerHTML" {
		t.Error("expected HX-Reswap header for HTMX request")
	}
	if !strings.Contains(w.Body.String(), "Errore") {
		t.Errorf("expected error message in HTMX response, got %q", w.Body.String())
	}
}

func TestRenderHTMXError_EscapesHTML(t *testing.T) {
	w := httptest.NewRecorder()
	RenderHTMXError(w, "<script>alert('xss')</script>", http.StatusBadRequest)

	body := w.Body.String()
	if strings.Contains(body, "<script>") {
		t.Error("HTML was not escaped in HTMX error response")
	}
	if !strings.Contains(body, "&lt;script&gt;") {
		t.Error("expected escaped HTML in response")
	}
}
