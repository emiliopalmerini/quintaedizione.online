package web

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/models"
	pkgweb "github.com/emiliopalmerini/quintaedizione.online/pkg/web"
)

// errorMappings is an ordered list of error patterns to match.
var errorMappings = []pkgweb.ErrorMapping{
	{
		Patterns:   []string{"not found", "document not found"},
		StatusCode: http.StatusNotFound,
		Message:    "La pagina o l'elemento richiesto non è stato trovato.",
	},
	{
		Patterns:   []string{"invalid collection", "invalid"},
		StatusCode: http.StatusBadRequest,
		Message:    "Collezione non valida o non supportata.",
	},
	{
		Patterns:   []string{"forbidden"},
		StatusCode: http.StatusForbidden,
		Message:    "Accesso non autorizzato.",
	},
	{
		Patterns:   []string{"unauthorized"},
		StatusCode: http.StatusUnauthorized,
		Message:    "Accesso non autorizzato.",
	},
	{
		Patterns:   []string{"timeout", "context deadline exceeded"},
		StatusCode: http.StatusGatewayTimeout,
		Message:    "Il server ha impiegato troppo tempo a rispondere. Riprova più tardi.",
	},
	{
		Patterns:   []string{"connection", "network"},
		StatusCode: http.StatusServiceUnavailable,
		Message:    "Problema di connessione al database. Riprova più tardi.",
	},
}

// ErrorResponse classifies err using SRD domain errors and shared patterns,
// then renders an appropriate error page.
func (h *baseHandler) ErrorResponse(w http.ResponseWriter, r *http.Request, err error, fallbackMessage string) {
	var httpErr *pkgweb.HTTPError
	if errors.As(err, &httpErr) {
		h.renderErrorPage(w, r, httpErr.Message, httpErr.Code)
		return
	}

	statusCode := h.getErrorStatusCode(err)
	message := h.getErrorMessage(err, fallbackMessage)

	log.Printf("Request error [%s %s]: %v", r.Method, r.URL.Path, err)

	h.renderErrorPage(w, r, message, statusCode)
}

func (h *baseHandler) getErrorStatusCode(err error) int {
	if domain.IsDocumentNotFound(err) {
		return http.StatusNotFound
	}
	if domain.IsInvalidDocumentID(err) {
		return http.StatusBadRequest
	}

	errStr := err.Error()
	for _, mapping := range errorMappings {
		if containsAny(errStr, mapping.Patterns...) {
			return mapping.StatusCode
		}
	}
	return http.StatusInternalServerError
}

func (h *baseHandler) getErrorMessage(err error, fallback string) string {
	if domain.IsDocumentNotFound(err) {
		return "La pagina o l'elemento richiesto non è stato trovato."
	}
	if domain.IsInvalidDocumentID(err) {
		return "Collezione non valida o non supportata."
	}

	errStr := err.Error()
	for _, mapping := range errorMappings {
		if containsAny(errStr, mapping.Patterns...) {
			return mapping.Message
		}
	}
	if fallback != "" {
		return fallback
	}
	return "Si è verificato un errore inaspettato. Riprova più tardi."
}

func (h *baseHandler) renderErrorPage(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	if r.Header.Get("HX-Request") == "true" {
		pkgweb.RenderHTMXError(w, message, statusCode)
		return
	}

	content, err := h.templateEngine.RenderError(r.Context(), models.ErrorPageData{
		PageData: models.PageData{
			Title: "Errore",
		},
		ErrorTitle:   errorPageTitle(statusCode),
		ErrorMessage: message,
		ErrorCode:    statusCode,
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

func errorPageTitle(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "Richiesta non valida"
	case http.StatusUnauthorized:
		return "Accesso non autorizzato"
	case http.StatusForbidden:
		return "Accesso negato"
	case http.StatusNotFound:
		return "Pagina non trovata"
	case http.StatusGatewayTimeout:
		return "Timeout della richiesta"
	case http.StatusServiceUnavailable:
		return "Servizio non disponibile"
	default:
		return "Errore"
	}
}

func containsAny(str string, substrings ...string) bool {
	lowerStr := strings.ToLower(str)
	for _, substr := range substrings {
		if strings.Contains(lowerStr, strings.ToLower(substr)) {
			return true
		}
	}
	return false
}
