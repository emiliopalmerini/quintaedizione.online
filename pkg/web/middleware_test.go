package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecurityMiddleware_Headers(t *testing.T) {
	handler := SecurityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
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

func TestSecurityMiddleware_CSP(t *testing.T) {
	handler := SecurityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	handler.ServeHTTP(w, r)

	csp := w.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Fatal("Content-Security-Policy header not set")
	}

	for _, want := range []string{"'sha256-", "https://fonts.googleapis.com", "https://fonts.gstatic.com"} {
		if !strings.Contains(csp, want) {
			t.Errorf("CSP missing %q\nGot: %s", want, csp)
		}
	}
}

func TestCORSMiddleware_Preflight(t *testing.T) {
	handler := CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called for OPTIONS")
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("OPTIONS", "/test", nil)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 for OPTIONS, got %d", w.Code)
	}
}

func TestRateLimitMiddleware_AllowsNormal(t *testing.T) {
	rl := NewRateLimiter()
	handler := RateLimitMiddleware(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestClientIPStripsRemoteAddrPort(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	r.RemoteAddr = "203.0.113.10:49152"

	if got := clientIP(r); got != "203.0.113.10" {
		t.Fatalf("clientIP = %q, want %q", got, "203.0.113.10")
	}
}

func TestClientIPFallsBackToRemoteAddrWhenUnparseable(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	r.RemoteAddr = "not-a-host-port"

	if got := clientIP(r); got != "not-a-host-port" {
		t.Fatalf("clientIP = %q, want raw RemoteAddr", got)
	}
}
