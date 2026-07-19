package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestScopedSearchHandlerRedirectsToSelectedSearch(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		location string
	}{
		{name: "compendio", url: "/cerca?q=palla+di+fuoco&scope=srd", location: "/srd/search?q=palla+di+fuoco"},
		{name: "mappe", url: "/cerca?q=castello&scope=mappe", location: "/mappe?q=castello"},
		{name: "generatori", url: "/cerca?q=png&scope=generatori", location: "/generatori?q=png"},
		{name: "unknown scope", url: "/cerca?q=drago&scope=unknown", location: "/srd/search?q=drago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tt.url, nil)

			ScopedSearchHandler(w, r)

			if w.Code != http.StatusSeeOther {
				t.Fatalf("expected status %d, got %d", http.StatusSeeOther, w.Code)
			}
			if location := w.Header().Get("Location"); location != tt.location {
				t.Errorf("expected location %q, got %q", tt.location, location)
			}
		})
	}
}
