package templates

import (
	"bytes"
	"strings"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/models"
)

func TestSearchPageRendersSnippet(t *testing.T) {
	data := models.SearchPageData{
		Query: "fuoco",
		Total: 1,
		Results: []models.CollectionSearchResult{
			{
				CollectionName:  "incantesimi",
				CollectionLabel: "Incantesimi",
				Total:           1,
				Documents: []models.Document{
					{
						ID:      "5.5e/palla-di-fuoco",
						Title:   "Palla di Fuoco",
						Snippet: "Una sfera di fuoco esplode nell'area scelta.",
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := SearchPage(data).Render(t.Context(), &buf); err != nil {
		t.Fatalf("render search page: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "search-page-result-snippet") {
		t.Fatalf("expected rendered search result snippet, got:\n%s", html)
	}
	if !strings.Contains(html, "<mark>fuoco</mark>") && !strings.Contains(html, "<mark>Fuoco</mark>") {
		t.Fatalf("expected snippet query highlight, got:\n%s", html)
	}
}

func TestSearchPageRendersCollectionIndex(t *testing.T) {
	data := models.SearchPageData{
		Query: "drago",
		Total: 45,
		Results: []models.CollectionSearchResult{
			{
				CollectionName:  "mostri",
				CollectionLabel: "Mostri",
				Total:           43,
			},
			{
				CollectionName:  "oggetti_magici",
				CollectionLabel: "Oggetti magici",
				Total:           2,
			},
		},
	}

	var buf bytes.Buffer
	if err := SearchPage(data).Render(t.Context(), &buf); err != nil {
		t.Fatalf("render search page: %v", err)
	}

	html := buf.String()
	for _, expected := range []string{
		`class="search-page-collection-index"`,
		`href="#search-collection-mostri"`,
		`id="search-collection-mostri"`,
		`href="#search-collection-oggetti_magici"`,
		`id="search-collection-oggetti_magici"`,
		`Mostri`,
		`43`,
		`Oggetti magici`,
		`2`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("expected rendered collection index to contain %q, got:\n%s", expected, html)
		}
	}
}

func TestSearchPageOmitsCollectionIndexWithoutResults(t *testing.T) {
	var buf bytes.Buffer
	if err := SearchPage(models.SearchPageData{Query: "inesistente"}).Render(t.Context(), &buf); err != nil {
		t.Fatalf("render search page: %v", err)
	}

	if strings.Contains(buf.String(), "search-page-collection-index") {
		t.Fatalf("expected no collection index without results, got:\n%s", buf.String())
	}
}

func TestSearchBrowseDropdownRendersSnippet(t *testing.T) {
	data := models.SearchBrowseData{
		Collections: []models.Collection{{Name: "incantesimi", Label: "Incantesimi", Count: 1}},
		Documents: []models.Document{
			{
				ID:      "5.5e/palla-di-fuoco",
				Title:   "Palla di Fuoco",
				Snippet: "Una sfera di fuoco esplode nell'area scelta.",
			},
		},
		CollectionName:   "incantesimi",
		CollectionLabel:  "Incantesimi",
		ActiveCollection: "incantesimi",
		Query:            "fuoco",
	}

	var buf bytes.Buffer
	if err := SearchBrowseDropdown(data).Render(t.Context(), &buf); err != nil {
		t.Fatalf("render search browse dropdown: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "search-browse-item-snippet") {
		t.Fatalf("expected rendered dropdown snippet, got:\n%s", html)
	}
}
