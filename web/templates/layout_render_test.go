package templates

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-h/templ"
	pkgweb "github.com/emiliopalmerini/quintaedizione.online/pkg/web"
)

func TestSiteLayoutIsStableAcrossPreferenceCookies(t *testing.T) {
	handler := pkgweb.ThemeMiddleware(pkgweb.PatreonBannerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := SiteLayout("", "", "", templ.NopComponent).Render(r.Context(), w); err != nil {
			t.Fatalf("render site layout: %v", err)
		}
	})))

	plain := httptest.NewRecorder()
	handler.ServeHTTP(plain, httptest.NewRequest("GET", "/", nil))
	personalized := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "theme", Value: "dark"})
	r.AddCookie(&http.Cookie{Name: "patreon_banner_dismissed", Value: "1"})
	handler.ServeHTTP(personalized, r)

	if plain.Body.String() != personalized.Body.String() {
		t.Fatal("cacheable layout must not vary with preference cookies")
	}
}

func TestSiteLayoutUsesTaskOrientedNavigation(t *testing.T) {
	var rendered strings.Builder
	if err := SiteLayout("", "", "", templ.NopComponent).Render(t.Context(), &rendered); err != nil {
		t.Fatalf("render site layout: %v", err)
	}

	html := rendered.String()
	for _, expected := range []string{
		`href="/giocare"`,
		`href="/preparare"`,
		`href="/srd"`,
		`>Giocare</a>`,
		`>Preparare</a>`,
		`>Compendio</a>`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("expected task-oriented navigation to contain %q", expected)
		}
	}
	if strings.Contains(html, `hx-boost="true"`) || strings.Contains(html, `hx-history-elt`) {
		t.Error("global shell must not mix body-level boosted swaps with content-only history")
	}
	if !strings.Contains(html, `<noscript><style>`) || !strings.Contains(html, `.site-nav-links`) {
		t.Error("layout must keep primary navigation available without JavaScript")
	}
}

func TestSiteLayoutMarksCurrentTaskHub(t *testing.T) {
	var rendered strings.Builder
	if err := SiteLayout("", "", "giocare", templ.NopComponent).Render(t.Context(), &rendered); err != nil {
		t.Fatalf("render site layout: %v", err)
	}
	if !strings.Contains(rendered.String(), `href="/giocare" class="site-nav-link active" aria-current="page"`) {
		t.Error("Giocare link must expose its current-page state")
	}
}
