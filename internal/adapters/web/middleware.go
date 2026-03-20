package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/collections"
	"golang.org/x/time/rate"
)

func ErrorRecoveryMiddleware(base *baseHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					isProduction := os.Getenv("ENVIRONMENT") == "production"
					if isProduction {
						log.Printf("PANIC recovered: %v", err)
					} else {
						stack := debug.Stack()
						log.Printf("PANIC recovered: %v\n%s", err, stack)
					}

					errMsg := "Si è verificato un errore interno del server"
					base.ErrorResponse(w, r, fmt.Errorf("internal server error"), errMsg)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func RequestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		raw := r.URL.RawQuery

		next.ServeHTTP(w, r)

		if raw != "" {
			path = path + "?" + raw
		}

		_ = path
	})
}

func SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; script-src 'self' 'unsafe-eval' 'sha256-iF3Ah6Tg3ke9rlMZ13UTaPKhQsXKcaTrBio4PaJBVCA='; img-src 'self' data: https:; font-src 'self' https://fonts.gstatic.com; connect-src 'self'")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}

func ValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validation is now handled per-route in chi via route-specific middleware
		next.ServeHTTP(w, r)
	})
}

func isValidCollection(collection string) bool {
	return collections.IsValid(collection)
}

func getValidCollections() []string {
	return collections.GetAllCollectionNames()
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	limiters map[string]*rateLimiterEntry
	mu       sync.RWMutex
	maxAge   time.Duration
}

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

// RateLimitMiddleware enforces per-IP rate limiting
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

// CollectionValidationMiddleware validates the {collection} URL parameter.
func CollectionValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		collection := r.PathValue("collection")
		if collection != "" && !isValidCollection(collection) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"error":             "Collezione non valida",
				"valid_collections": getValidCollections(),
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware handles CORS headers.
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
