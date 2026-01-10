package web

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/collections"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func (h *Handlers) ErrorRecoveryMiddleware() gin.HandlerFunc {
	base := h.baseHandlerForMiddleware()
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {

				// Only log detailed stack trace in development mode
				isProduction := os.Getenv("ENVIRONMENT") == "production"
				if isProduction {
					// In production, log minimal information for security
					log.Printf("PANIC recovered: %v", err)
				} else {
					// In development, log full stack trace for debugging
					stack := debug.Stack()
					log.Printf("PANIC recovered: %v\n%s", err, stack)
				}

				errMsg := fmt.Sprintf("Si è verificato un errore interno del server")
				base.ErrorResponse(c, fmt.Errorf("internal server error"), errMsg)

				c.Abort()
			}
		}()

		c.Next()
	}
}

func RequestLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		start := c.Request.Context()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		if len(c.Errors) > 0 {
			log.Printf("Request errors for %s %s: %v", c.Request.Method, path, c.Errors)
		}

		if start != nil {

		}
	}
}

func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content-Security-Policy prevents inline scripts and restricts resource loading
		c.Header("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'")

		// Strict-Transport-Security enforces HTTPS connections
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		c.Next()
	}
}

func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		if collection := c.Param("collection"); collection != "" {
			if !isValidCollection(collection) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":             "Collezione non valida",
					"valid_collections": getValidCollections(),
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
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
func RateLimitMiddleware(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()

		// Get the rate limiter for this IP
		limiter := rl.GetLimiter(clientIP)

		// Check if request is allowed
		if !limiter.Allow() {
			c.Header("Retry-After", "1")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Troppe richieste. Per favore riprova tra un secondo.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
