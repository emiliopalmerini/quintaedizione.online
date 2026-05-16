package templates

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-h/templ"
	pkgweb "github.com/emiliopalmerini/quintaedizione.online/pkg/web"
)

func TestSiteLayoutHidesDismissedPatreonBanner(t *testing.T) {
	handler := pkgweb.PatreonBannerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := SiteLayout("", "", "", templ.NopComponent).Render(r.Context(), w); err != nil {
			t.Fatalf("render site layout: %v", err)
		}
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "patreon_banner_dismissed", Value: "1"})
	handler.ServeHTTP(w, r)

	html := w.Body.String()
	if !strings.Contains(html, `id="patreon-banner"`) {
		t.Fatalf("expected patreon banner in layout, got:\n%s", html)
	}
	if !strings.Contains(html, "patreon-banner--hidden") {
		t.Fatalf("expected dismissed patreon banner to render hidden, got:\n%s", html)
	}
}
