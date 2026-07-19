package templates

import (
	"bytes"
	"strings"
	"testing"

	maps "github.com/emiliopalmerini/quintaedizione.online/internal/mappe/domain"
)

func TestGalleryPageUsesNativeCanonicalFilters(t *testing.T) {
	data := maps.GalleryData{
		Tags: []string{"dungeon", "rovine"}, ActiveTags: []string{"dungeon"}, Query: "torre",
	}
	var rendered bytes.Buffer
	if err := GalleryPage(data).Render(t.Context(), &rendered); err != nil {
		t.Fatalf("render gallery: %v", err)
	}
	body := rendered.String()

	for _, expected := range []string{
		`action="/mappe"`,
		`method="get"`,
		`role="search"`,
		`type="search"`,
		`name="q"`,
		`type="checkbox" name="tag" value="dungeon" checked`,
		`type="checkbox" name="tag" value="rovine"`,
		`type="submit"`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("expected gallery to contain %q, got:\n%s", expected, body)
		}
	}
	if strings.Contains(body, `/mappe/gallery`) || strings.Contains(body, `type="hidden" name="tag"`) {
		t.Error("gallery filters must use canonical URLs and native named controls")
	}
}

func TestGalleryLoadMoreAndImagesHaveProgressiveSemantics(t *testing.T) {
	data := maps.GalleryData{
		Mappe: []maps.Mappa{
			{Slug: "prima", Nome: "Prima", Immagine: "prima.webp"},
			{Slug: "seconda", Nome: "Seconda", Immagine: "seconda.webp"},
		},
		Query: "torre", ActiveTags: []string{"dungeon", "rovine"},
		Total: 100, Offset: 0, Limit: 40, HasMore: true,
	}
	var rendered bytes.Buffer
	if err := GalleryGrid(data).Render(t.Context(), &rendered); err != nil {
		t.Fatalf("render gallery grid: %v", err)
	}
	body := rendered.String()

	for _, expected := range []string{
		`<a href="/mappe?`,
		`hx-get="/mappe?`,
		`hx-target="#mappe-load-more"`,
		`alt=""`,
		`decoding="async"`,
		`fetchpriority="high"`,
		`loading="lazy"`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("expected gallery grid to contain %q, got:\n%s", expected, body)
		}
	}
	if strings.Contains(body, `/mappe/gallery`) {
		t.Error("load-more must use the canonical gallery URL")
	}
}
