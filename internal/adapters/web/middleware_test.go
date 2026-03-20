package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecurityMiddleware_CSP(t *testing.T) {
	handler := SecurityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	handler.ServeHTTP(w, r)

	csp := w.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Fatal("Content-Security-Policy header not set")
	}

	tests := []struct {
		name     string
		contains string
		reason   string
	}{
		{
			name:     "inline theme script hash",
			contains: "'sha256-",
			reason:   "inline theme-detection script needs a hash allowance",
		},
		{
			name:     "Google Fonts stylesheets",
			contains: "https://fonts.googleapis.com",
			reason:   "Inter font loaded from Google Fonts",
		},
		{
			name:     "Google Fonts files",
			contains: "https://fonts.gstatic.com",
			reason:   "font files served from fonts.gstatic.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(csp, tt.contains) {
				t.Errorf("CSP missing %s (%s)\nGot: %s", tt.contains, tt.reason, csp)
			}
		})
	}
}

func TestSecurityMiddleware_Headers(t *testing.T) {
	handler := SecurityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	handler.ServeHTTP(w, r)

	headers := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}

	for name, want := range headers {
		got := w.Header().Get(name)
		if got != want {
			t.Errorf("Header %s = %q, want %q", name, got, want)
		}
	}
}
