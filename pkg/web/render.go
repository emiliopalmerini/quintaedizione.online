package web

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
)

// RenderTempl sets the Content-Type header and renders a templ component.
// On error it logs the failure and sends a 500 response.
func RenderTempl(w http.ResponseWriter, r *http.Request, logger *slog.Logger, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(r.Context(), w); err != nil {
		logger.Error("Failed to render template", "path", r.URL.Path, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
