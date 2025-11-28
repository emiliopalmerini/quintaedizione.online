package web

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Message)
}

func (h *Handlers) ErrorResponse(c *gin.Context, err error, fallbackMessage string) {
	var httpErr *HTTPError

	if errors.As(err, &httpErr) {
		h.renderErrorPage(c, httpErr.Message, httpErr.Code)
		return
	}

	statusCode := h.getErrorStatusCode(err)
	message := h.getErrorMessage(err, fallbackMessage)

	log.Printf("Request error [%s %s]: %v", c.Request.Method, c.Request.URL.Path, err)

	h.renderErrorPage(c, message, statusCode)
}

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

func (h *Handlers) renderErrorPage(c *gin.Context, message string, statusCode int) {

	if c.GetHeader("HX-Request") == "true" {
		h.renderHTMXError(c, message, statusCode)
		return
	}

	data := gin.H{
		"title":       "Errore",
		"error":       message,
		"status_code": statusCode,
		"show_home":   true,
	}

	content, err := h.templateEngine.Render("error.html", data)
	if err != nil {

		c.String(statusCode, "Errore: %s", message)
		return
	}

	c.Data(statusCode, "text/html; charset=utf-8", []byte(content))
}

func (h *Handlers) renderHTMXError(c *gin.Context, message string, statusCode int) {

	errorHTML := fmt.Sprintf(`
		<div class="error-message" style="padding: 1rem; background: var(--error); color: white; border-radius: 4px; margin: 1rem 0;">
			<strong>Errore:</strong> %s
			<button onclick="this.parentElement.remove()" style="float: right; background: none; border: none; color: white; cursor: pointer;">×</button>
		</div>
	`, message)

	c.Header("HX-Reswap", "innerHTML")
	c.Data(statusCode, "text/html; charset=utf-8", []byte(errorHTML))
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
	str = fmt.Sprintf("%s", str)
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

func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}
