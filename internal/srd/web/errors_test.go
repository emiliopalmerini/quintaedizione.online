package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/pkg/templates"
)

func TestErrorResponseRendersTemplErrorPage(t *testing.T) {
	handler := &baseHandler{templateEngine: templates.NewTemplEngine()}
	req := httptest.NewRequest(http.MethodGet, "/srd/incantesimi/missing", nil)
	rec := httptest.NewRecorder()

	handler.ErrorResponse(rec, req, errors.New("document not found"), "Elemento non trovato")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(strings.ToLower(body), "<!doctype html>") {
		t.Fatalf("expected templ HTML error page, got: %s", body)
	}
	if !strings.Contains(body, "404") {
		t.Fatalf("expected rendered status code in body, got: %s", body)
	}
}

func TestErrorResponseClassifiesForbiddenAsForbidden(t *testing.T) {
	handler := &baseHandler{templateEngine: templates.NewTemplEngine()}
	req := httptest.NewRequest(http.MethodGet, "/srd/private", nil)
	rec := httptest.NewRecorder()

	handler.ErrorResponse(rec, req, errors.New("forbidden"), "")

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if !strings.Contains(rec.Body.String(), "403") {
		t.Fatalf("expected rendered status code in body, got: %s", rec.Body.String())
	}
}
