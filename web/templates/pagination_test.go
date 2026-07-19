package templates

import (
	"bytes"
	"context"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/models"
)

func TestPaginationPages(t *testing.T) {
	tests := []struct {
		name      string
		current   int
		total     int
		wantPages []int
		wantEllip bool // at least one ellipsis expected
	}{
		{
			name:      "small total shows all pages",
			current:   1,
			total:     5,
			wantPages: []int{1, 2, 3, 4, 5},
		},
		{
			name:      "beginning of large range",
			current:   1,
			total:     33,
			wantPages: []int{1, 2, 3, -1, 33}, // -1 = ellipsis
			wantEllip: true,
		},
		{
			name:      "middle of large range",
			current:   15,
			total:     33,
			wantPages: []int{1, -1, 13, 14, 15, 16, 17, -1, 33},
			wantEllip: true,
		},
		{
			name:      "end of large range",
			current:   33,
			total:     33,
			wantPages: []int{1, -1, 31, 32, 33},
			wantEllip: true,
		},
		{
			name:      "near beginning",
			current:   3,
			total:     33,
			wantPages: []int{1, 2, 3, 4, 5, -1, 33},
			wantEllip: true,
		},
		{
			name:      "near end",
			current:   31,
			total:     33,
			wantPages: []int{1, -1, 29, 30, 31, 32, 33},
			wantEllip: true,
		},
		{
			name:      "single page",
			current:   1,
			total:     1,
			wantPages: []int{1},
		},
		{
			name:      "two pages",
			current:   1,
			total:     2,
			wantPages: []int{1, 2},
		},
		{
			name:      "seven pages no ellipsis needed",
			current:   4,
			total:     7,
			wantPages: []int{1, 2, 3, 4, 5, 6, 7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := paginationPages(tt.current, tt.total)
			if !reflect.DeepEqual(got, tt.wantPages) {
				t.Errorf("paginationPages(%d, %d) = %v, want %v", tt.current, tt.total, got, tt.wantPages)
			}
		})
	}
}

func TestPaginationURLUsesCanonicalCollectionAndReplacesPage(t *testing.T) {
	data := models.CollectionPageData{
		PageData: models.PageData{
			Collection:  "incantesimi",
			QueryString: "q=fuoco+%26+fiamme&page=9&livello=0&livello=1",
		},
	}

	got := paginationURL(data, 2)
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse pagination URL: %v", err)
	}
	if parsed.Path != "/srd/incantesimi" {
		t.Errorf("expected canonical collection path, got %q", parsed.Path)
	}
	if pages := parsed.Query()["page"]; !reflect.DeepEqual(pages, []string{"2"}) {
		t.Errorf("expected one replaced page value, got %v", pages)
	}
	if levels := parsed.Query()["livello"]; !reflect.DeepEqual(levels, []string{"0", "1"}) {
		t.Errorf("expected repeated filters to be preserved, got %v", levels)
	}
	if query := parsed.Query().Get("q"); query != "fuoco & fiamme" {
		t.Errorf("expected encoded search query to round trip, got %q", query)
	}
}

func TestPaginationRendersNativeEnhancedLinks(t *testing.T) {
	data := models.CollectionPageData{
		PageData: models.PageData{Collection: "incantesimi", QueryString: "q=fuoco"},
		Page:     2, TotalPages: 3, HasPrev: true, HasNext: true,
		StartItem: 21, EndItem: 40, Total: 60,
	}
	var rendered bytes.Buffer
	if err := pagination(data).Render(context.Background(), &rendered); err != nil {
		t.Fatalf("render pagination: %v", err)
	}

	html := rendered.String()
	for _, expected := range []string{
		`href="/srd/incantesimi?`,
		`hx-get="/srd/incantesimi?`,
		`hx-target="#rows"`,
		`hx-push-url="true"`,
		`aria-current="page"`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("expected pagination to contain %q, got:\n%s", expected, html)
		}
	}
	if strings.Contains(html, "/srd/rows/") {
		t.Errorf("pagination must not contain fragment URLs, got:\n%s", html)
	}
}
