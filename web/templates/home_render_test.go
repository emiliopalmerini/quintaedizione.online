package templates

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/models"
)

func TestHomePageProvidesNativeCompendiumSearch(t *testing.T) {
	var rendered bytes.Buffer
	if err := HomePage(models.HomePageData{}).Render(context.Background(), &rendered); err != nil {
		t.Fatalf("render home page: %v", err)
	}
	html := rendered.String()

	for _, expected := range []string{
		`<h1>Compendio</h1>`,
		`action="/srd/search"`,
		`method="get"`,
		`name="q"`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("expected Compendio home to contain %q", expected)
		}
	}
}
