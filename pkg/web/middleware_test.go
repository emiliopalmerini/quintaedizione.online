package web

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCompressionMiddlewareNegotiatesGzip(t *testing.T) {
	handler := CompressionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "HX-Request")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("contenuto compresso"))
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/giocare", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("Content-Encoding = %q", rec.Header().Get("Content-Encoding"))
	}
	if vary := strings.Join(rec.Header().Values("Vary"), ","); !strings.Contains(vary, "Accept-Encoding") || !strings.Contains(vary, "HX-Request") {
		t.Errorf("Vary = %q", vary)
	}
	reader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "contenuto compresso" {
		t.Errorf("body = %q", body)
	}
}

func TestCompressionMiddlewareSkipsImages(t *testing.T) {
	handler := CompressionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("image"))
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/map.webp", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	handler.ServeHTTP(rec, req)
	if rec.Header().Get("Content-Encoding") != "" || rec.Body.String() != "image" {
		t.Error("pre-compressed images must pass through unchanged")
	}
}

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

func TestPatreonBannerMiddlewareStoresDismissedCookie(t *testing.T) {
	handler := PatreonBannerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !PatreonBannerDismissedFromContext(r.Context()) {
			t.Fatal("expected patreon banner dismissal in request context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.AddCookie(&http.Cookie{Name: "patreon_banner_dismissed", Value: "1"})
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestPatreonBannerMiddlewareDefaultsToVisible(t *testing.T) {
	handler := PatreonBannerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if PatreonBannerDismissedFromContext(r.Context()) {
			t.Fatal("expected patreon banner dismissal to be absent")
		}
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
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
