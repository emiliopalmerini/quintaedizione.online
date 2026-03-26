package web

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type contextKey string

const themeContextKey contextKey = "theme"

// ThemeMiddleware reads the "theme" cookie and stores its value in the
// request context so that templates can render the correct class on <html>
// without waiting for client-side JavaScript.
func ThemeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie("theme"); err == nil && (c.Value == "dark" || c.Value == "light") {
			ctx := context.WithValue(r.Context(), themeContextKey, c.Value)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

// ThemeFromContext returns the theme stored in ctx ("dark" or "light"),
// or an empty string if none is set.
func ThemeFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(themeContextKey).(string); ok {
		return v
	}
	return ""
}

// SecurityMiddleware sets standard security headers on every response.
func SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; script-src 'self' 'sha256-pLQrlHs1p6y0CfQot82jEWlEtkspKO8xyNLNHUouY68='; img-src 'self' data: https:; font-src 'self' https://fonts.gstatic.com; connect-src 'self'")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware handles CORS headers with origin allowlisting.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		allowedOrigins := map[string]bool{
			"https://quintaedizione.online":     true,
			"https://www.quintaedizione.online": true,
		}

		if os.Getenv("ENVIRONMENT") != "production" {
			allowedOrigins["http://localhost:3000"] = true
			allowedOrigins["http://localhost:8000"] = true
			allowedOrigins["http://127.0.0.1:3000"] = true
			allowedOrigins["http://127.0.0.1:8000"] = true
		}

		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RateLimiter tracks per-IP rate limiters with automatic cleanup.
type RateLimiter struct {
	limiters map[string]*rateLimiterEntry
	mu       sync.RWMutex
	maxAge   time.Duration
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a per-IP rate limiter with hourly cleanup.
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*rateLimiterEntry),
		maxAge:   time.Hour,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		for ip, entry := range rl.limiters {
			if time.Since(entry.lastSeen) > rl.maxAge {
				delete(rl.limiters, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// GetLimiter returns (or creates) the rate limiter for the given IP.
func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if entry, exists := rl.limiters[ip]; exists {
		entry.lastSeen = time.Now()
		return entry.limiter
	}

	limiter := rate.NewLimiter(rate.Limit(100), 10)
	rl.limiters[ip] = &rateLimiterEntry{
		limiter:  limiter,
		lastSeen: time.Now(),
	}
	return limiter
}

// RateLimitMiddleware enforces per-IP rate limiting.
func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := r.RemoteAddr

			limiter := rl.GetLimiter(clientIP)

			if !limiter.Allow() {
				w.Header().Set("Retry-After", "1")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Troppe richieste. Per favore riprova tra un secondo.",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ErrorRecoveryMiddleware recovers from panics and returns a 500 response.
func ErrorRecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("Panic recovered", "error", err, "path", r.URL.Path)
					http.Error(w, "Si è verificato un errore interno del server", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
