package web

import (
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"strings"
)

// HTTPError represents a typed HTTP error with status code and message.
type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Message)
}

// NewHTTPError creates an HTTPError with the given code and message.
func NewHTTPError(code int, message string) *HTTPError {
	return &HTTPError{Code: code, Message: message}
}

// NewHTTPErrorWithDetail creates an HTTPError with an additional detail field.
func NewHTTPErrorWithDetail(code int, message, detail string) *HTTPError {
	return &HTTPError{Code: code, Message: message, Detail: detail}
}

// ErrorMapping defines a pattern-based error classification rule.
type ErrorMapping struct {
	Patterns   []string
	StatusCode int
	Message    string
}

// ErrorResponder provides structured error responses for HTTP handlers.
// It classifies errors using the provided mappings and renders either
// a full error page or an HTMX inline error fragment.
type ErrorResponder struct {
	Logger   *slog.Logger
	Mappings []ErrorMapping
}

// Respond classifies err, logs it, and writes an appropriate error response.
// For HTMX requests it returns an inline HTML fragment; otherwise a plain text error.
func (er *ErrorResponder) Respond(w http.ResponseWriter, r *http.Request, err error, fallbackMessage string) {
	statusCode, message := er.classify(err, fallbackMessage)

	er.Logger.Error("Request error",
		"method", r.Method,
		"path", r.URL.Path,
		"status", statusCode,
		"error", err,
	)

	if r.Header.Get("HX-Request") == "true" {
		RenderHTMXError(w, message, statusCode)
		return
	}

	http.Error(w, message, statusCode)
}

// classify determines the HTTP status code and user-facing message for an error.
func (er *ErrorResponder) classify(err error, fallback string) (int, string) {
	var httpErr *HTTPError
	if asHTTPError(err, &httpErr) {
		return httpErr.Code, httpErr.Message
	}

	errStr := strings.ToLower(err.Error())
	for _, m := range er.Mappings {
		for _, p := range m.Patterns {
			if strings.Contains(errStr, strings.ToLower(p)) {
				return m.StatusCode, m.Message
			}
		}
	}

	if fallback != "" {
		return http.StatusInternalServerError, fallback
	}
	return http.StatusInternalServerError, "Si è verificato un errore inaspettato. Riprova più tardi."
}

// asHTTPError unwraps err into an *HTTPError if possible.
func asHTTPError(err error, target **HTTPError) bool {
	type unwrapper interface {
		Unwrap() error
	}
	// Direct type assertion first
	if he, ok := err.(*HTTPError); ok {
		*target = he
		return true
	}
	// Walk the chain
	for {
		u, ok := err.(unwrapper)
		if !ok {
			return false
		}
		err = u.Unwrap()
		if err == nil {
			return false
		}
		if he, ok := err.(*HTTPError); ok {
			*target = he
			return true
		}
	}
}

// RenderHTMXError writes an inline HTML error fragment suitable for HTMX swap.
func RenderHTMXError(w http.ResponseWriter, message string, statusCode int) {
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
