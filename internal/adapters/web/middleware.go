package web

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) ErrorRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {

				stack := debug.Stack()
				log.Printf("PANIC recovered: %v\n%s", err, stack)

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
