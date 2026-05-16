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
