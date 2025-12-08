package web

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func (h *Handlers) ErrorRecoveryMiddleware() gin.HandlerFunc {
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
				h.ErrorResponse(c, fmt.Errorf("internal server error"), errMsg)

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
	validCollections := getValidCollections()
	for _, valid := range validCollections {
		if collection == valid {
			return true
		}
	}
	return false
}

func getValidCollections() []string {
	return []string{
		"incantesimi", "mostri", "classi", "backgrounds", "equipaggiamenti",
		"oggetti_magici", "armi", "armature", "talenti", "servizi",
		"strumenti", "animali", "regole", "cavalcature_veicoli",
	}
}

// RateLimiter stores per-IP rate limiters
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

// NewRateLimiter creates a new rate limiter with per-IP tracking
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

// GetLimiter returns a rate limiter for the given IP, creating one if needed
func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if limiter, exists := rl.limiters[ip]; exists {
		return limiter
	}

	// 100 requests per second per IP, burst of 10
	limiter := rate.NewLimiter(rate.Limit(100), 10)
	rl.limiters[ip] = limiter
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
				"error": "Troppi richieste. Per favore riprova tra un secondo.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
