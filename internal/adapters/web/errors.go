package web

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HTTPError represents a structured HTTP error
type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

// Error implements the error interface
func (e HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Message)
}

// ErrorResponse handles different types of errors and renders appropriate responses
func (h *Handlers) ErrorResponse(c *gin.Context, err error, fallbackMessage string) {
	var httpErr *HTTPError

	// Check if it's already a structured HTTP error
	if errors.As(err, &httpErr) {
		h.renderErrorPage(c, httpErr.Message, httpErr.Code)
		return
	}

	// Map common errors to HTTP status codes
	statusCode := h.getErrorStatusCode(err)
	message := h.getErrorMessage(err, fallbackMessage)

	// Log error for debugging
	log.Printf("Request error [%s %s]: %v", c.Request.Method, c.Request.URL.Path, err)

	h.renderErrorPage(c, message, statusCode)
}

// getErrorStatusCode maps common errors to HTTP status codes
func (h *Handlers) getErrorStatusCode(err error) int {
	errStr := err.Error()

	switch {
	case contains(errStr, "not found", "document not found"):
		return http.StatusNotFound
	case contains(errStr, "invalid collection", "invalid"):
		return http.StatusBadRequest
	case contains(errStr, "unauthorized", "forbidden"):
		return http.StatusUnauthorized
	case contains(errStr, "timeout", "context deadline exceeded"):
		return http.StatusGatewayTimeout
	case contains(errStr, "connection", "network"):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// getErrorMessage provides user-friendly error messages
func (h *Handlers) getErrorMessage(err error, fallback string) string {
	errStr := err.Error()

	switch {
	case contains(errStr, "not found", "document not found"):
		return "La pagina o l'elemento richiesto non è stato trovato."
	case contains(errStr, "invalid collection"):
		return "Collezione non valida o non supportata."
	case contains(errStr, "timeout"):
		return "Il server ha impiegato troppo tempo a rispondere. Riprova più tardi."
	case contains(errStr, "connection", "network"):
		return "Problema di connessione al database. Riprova più tardi."
	default:
		if fallback != "" {
			return fallback
		}
		return "Si è verificato un errore inaspettato. Riprova più tardi."
	}
}

// renderErrorPage renders a structured error page
func (h *Handlers) renderErrorPage(c *gin.Context, message string, statusCode int) {
	// Check if this is an HTMX request
	if c.GetHeader("HX-Request") == "true" {
		h.renderHTMXError(c, message, statusCode)
		return
	}

	// Render full error page
	data := gin.H{
		"title":       "Errore",
		"error":       message,
		"status_code": statusCode,
		"show_home":   true,
	}

	content, err := h.templateEngine.Render("error.html", data)
	if err != nil {
		// Fallback to simple error response if template rendering fails
		c.String(statusCode, "Errore: %s", message)
		return
	}

	c.Data(statusCode, "text/html; charset=utf-8", []byte(content))
}

// renderHTMXError renders error content for HTMX requests
func (h *Handlers) renderHTMXError(c *gin.Context, message string, statusCode int) {
	// Return simple error content for HTMX partial updates
	errorHTML := fmt.Sprintf(`
		<div class="error-message" style="padding: 1rem; background: var(--error); color: white; border-radius: 4px; margin: 1rem 0;">
			<strong>Errore:</strong> %s
			<button onclick="this.parentElement.remove()" style="float: right; background: none; border: none; color: white; cursor: pointer;">×</button>
		</div>
	`, message)

	c.Header("HX-Reswap", "innerHTML")
	c.Data(statusCode, "text/html; charset=utf-8", []byte(errorHTML))
}

// NewHTTPError creates a new HTTP error
func NewHTTPError(code int, message string) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: message,
	}
}

// NewHTTPErrorWithDetail creates a new HTTP error with detail
func NewHTTPErrorWithDetail(code int, message, detail string) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: message,
		Detail:  detail,
	}
}

// contains checks if a string contains any of the given substrings (case insensitive)
func contains(str string, substrings ...string) bool {
	str = fmt.Sprintf("%s", str) // Convert to lowercase
	for _, substr := range substrings {
		if len(str) >= len(substr) {
			for i := 0; i <= len(str)-len(substr); i++ {
				match := true
				for j := 0; j < len(substr); j++ {
					if toLower(str[i+j]) != toLower(substr[j]) {
						match = false
						break
					}
				}
				if match {
					return true
				}
			}
		}
	}
	return false
}

// toLower converts a byte to lowercase
func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}
