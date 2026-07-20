package templates

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/models"
)

func TestCollectionPageUsesNativeCanonicalFilterForm(t *testing.T) {
	data := models.CollectionPageData{
		PageData: models.PageData{Collection: "incantesimi", Title: "Incantesimi"},
		Filters: []models.FilterOption{
			{
				Name: "livello", Label: "Livello",
				Values:        []models.FilterValueOption{{Value: "0", Count: 12}, {Value: "1", Count: 8}},
				CurrentValues: []string{"0"},
			},
		},
	}

	var rendered bytes.Buffer
	if err := CollectionPage(data).Render(context.Background(), &rendered); err != nil {
		t.Fatalf("render collection page: %v", err)
	}
	html := rendered.String()

	for _, expected := range []string{
		`action="/srd/incantesimi"`,
		`method="get"`,
		`hx-get="/srd/incantesimi"`,
		`type="submit"`,
		`name="q"`,
		`type="checkbox" name="livello" value="0" checked`,
		`<noscript>`,
		`<h1`,
		`aria-current="page">Incantesimi`,
		`class="collection-search-sentinel"`,
		`class="collection-search-placeholder"`,
		`class="collection-search-clearance"`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("expected collection page to contain %q, got:\n%s", expected, html)
		}
	}
	if strings.Contains(html, "/srd/rows/") {
		t.Errorf("collection page must not contain fragment URLs, got:\n%s", html)
	}
	if strings.Contains(html, `type="hidden" name="livello"`) {
		t.Error("JavaScript filter state must not duplicate native named controls")
	}
	if count := strings.Count(html, `id="search-form"`); count != 1 {
		t.Errorf("collection page must contain exactly one search form, got %d", count)
	}
}
