package web

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// ErrorRecoveryMiddleware provides panic recovery and structured error handling
func (h *Handlers) ErrorRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				stack := debug.Stack()
				log.Printf("PANIC recovered: %v\n%s", err, stack)
				
				// Create error response
				errMsg := fmt.Sprintf("Si è verificato un errore interno del server")
				h.ErrorResponse(c, fmt.Errorf("internal server error"), errMsg)
				
				// Abort the request
				c.Abort()
			}
		}()
		
		c.Next()
	}
}

// RequestLoggingMiddleware logs requests with additional context
func RequestLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := c.Request.Context()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		
		// Process request
		c.Next()
		
		// Log request details
		if raw != "" {
			path = path + "?" + raw
		}
		
		// Log errors if any
		if len(c.Errors) > 0 {
			log.Printf("Request errors for %s %s: %v", c.Request.Method, path, c.Errors)
		}
		
		// Log long-running requests
		if start != nil {
			// Note: We could add timing here if needed
		}
	}
}

// SecurityMiddleware adds basic security headers
func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Basic security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		c.Next()
	}
}

// ValidationMiddleware validates common request parameters
func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate collection parameter if present
		if collection := c.Param("collection"); collection != "" {
			if !isValidCollection(collection) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Collezione non valida",
					"valid_collections": getValidCollections(),
				})
				c.Abort()
				return
			}
		}
		
		c.Next()
	}
}

// isValidCollection checks if a collection name is valid
func isValidCollection(collection string) bool {
	validCollections := getValidCollections()
	for _, valid := range validCollections {
		if collection == valid {
			return true
		}
	}
	return false
}

// getValidCollections returns the list of valid collection names
func getValidCollections() []string {
	return []string{
		"incantesimi", "mostri", "classi", "backgrounds", "equipaggiamenti",
		"oggetti_magici", "armi", "armature", "talenti", "servizi",
		"strumenti", "animali", "regole", "cavalcature_veicoli",
	}
}