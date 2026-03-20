package web

import (
	"errors"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain"
)

type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Message)
}

func (h *baseHandler) ErrorResponse(w http.ResponseWriter, r *http.Request, err error, fallbackMessage string) {
	var httpErr *HTTPError

	if errors.As(err, &httpErr) {
		h.renderErrorPage(w, r, httpErr.Message, httpErr.Code)
		return
	}

	statusCode := h.getErrorStatusCode(err)
	message := h.getErrorMessage(err, fallbackMessage)

	log.Printf("Request error [%s %s]: %v", r.Method, r.URL.Path, err)

	h.renderErrorPage(w, r, message, statusCode)
}

// errorMapping defines error patterns and their corresponding status codes and messages.
type errorMapping struct {
	patterns   []string
	statusCode int
	message    string
}

// errorMappings is an ordered list of error patterns to match.
// Order matters: more specific patterns should come first.
var errorMappings = []errorMapping{
	{
		patterns:   []string{"not found", "document not found"},
		statusCode: http.StatusNotFound,
		message:    "La pagina o l'elemento richiesto non è stato trovato.",
	},
	{
		patterns:   []string{"invalid collection", "invalid"},
		statusCode: http.StatusBadRequest,
		message:    "Collezione non valida o non supportata.",
	},
	{
		patterns:   []string{"unauthorized", "forbidden"},
		statusCode: http.StatusUnauthorized,
		message:    "Accesso non autorizzato.",
	},
	{
		patterns:   []string{"timeout", "context deadline exceeded"},
		statusCode: http.StatusGatewayTimeout,
		message:    "Il server ha impiegato troppo tempo a rispondere. Riprova più tardi.",
	},
	{
		patterns:   []string{"connection", "network"},
		statusCode: http.StatusServiceUnavailable,
		message:    "Problema di connessione al database. Riprova più tardi.",
	},
}

func (h *baseHandler) getErrorStatusCode(err error) int {
	// Check typed domain errors first
	if domain.IsDocumentNotFound(err) {
		return http.StatusNotFound
	}
	if domain.IsInvalidDocumentID(err) {
		return http.StatusBadRequest
	}

	// Fall back to string matching for errors without typed wrappers
	errStr := err.Error()
	for _, mapping := range errorMappings {
		if contains(errStr, mapping.patterns...) {
			return mapping.statusCode
		}
	}
	return http.StatusInternalServerError
}

func (h *baseHandler) getErrorMessage(err error, fallback string) string {
	// Check typed domain errors first
	if domain.IsDocumentNotFound(err) {
		return "La pagina o l'elemento richiesto non è stato trovato."
	}
	if domain.IsInvalidDocumentID(err) {
		return "Collezione non valida o non supportata."
	}

	// Fall back to string matching for errors without typed wrappers
	errStr := err.Error()
	for _, mapping := range errorMappings {
		if contains(errStr, mapping.patterns...) {
			return mapping.message
		}
	}
	if fallback != "" {
		return fallback
	}
	return "Si è verificato un errore inaspettato. Riprova più tardi."
}

func (h *baseHandler) renderErrorPage(w http.ResponseWriter, r *http.Request, message string, statusCode int) {

	if r.Header.Get("HX-Request") == "true" {
		h.renderHTMXError(w, message, statusCode)
		return
	}

	content, err := h.templateEngine.Render("error.html", map[string]any{
		"title":       "Errore",
		"error":       message,
		"status_code": statusCode,
		"show_home":   true,
	})
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(statusCode)
		fmt.Fprintf(w, "Errore: %s", message)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write([]byte(content))
}

func (h *baseHandler) renderHTMXError(w http.ResponseWriter, message string, statusCode int) {

	// Escape HTML special characters to prevent XSS in error messages
	escapedMessage := html.EscapeString(message)

	errorHTML := fmt.Sprintf(`
		<div class="error-message" style="padding: 1rem; background: var(--error); color: white; border-radius: 4px; margin: 1rem 0;">
			<strong>Errore:</strong> %s
			<button onclick="this.parentElement.remove()" style="float: right; background: none; border: none; color: white; cursor: pointer;">×</button>
		</div>
	`, escapedMessage)

	w.Header().Set("HX-Reswap", "innerHTML")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write([]byte(errorHTML))
}

func NewHTTPError(code int, message string) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: message,
	}
}

func NewHTTPErrorWithDetail(code int, message, detail string) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: message,
		Detail:  detail,
	}
}

func contains(str string, substrings ...string) bool {
	lowerStr := strings.ToLower(str)
	for _, substr := range substrings {
		if strings.Contains(lowerStr, strings.ToLower(substr)) {
			return true
		}
	}
	return false
}
