package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterRoutesRedirectsLegacyAreas(t *testing.T) {
	tests := []struct {
		path     string
		location string
	}{
		{path: "/area/magia-mostri", location: "/srd/area/magia"},
		{path: "/area/riferimento", location: "/srd/area/regole"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			mux := http.NewServeMux()
			(&Handlers{}).RegisterRoutes(mux)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if w.Code != http.StatusMovedPermanently {
				t.Fatalf("expected status %d, got %d", http.StatusMovedPermanently, w.Code)
			}
			if location := w.Header().Get("Location"); location != tt.location {
				t.Errorf("expected location %q, got %q", tt.location, location)
			}
		})
	}
}
