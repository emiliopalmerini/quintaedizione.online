package templates

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestLandingPageProvidesNativeSearchAndTaskNavigation(t *testing.T) {
	var rendered bytes.Buffer
	if err := LandingPage().Render(context.Background(), &rendered); err != nil {
		t.Fatalf("render landing page: %v", err)
	}

	html := rendered.String()
	for _, expected := range []string{
		`action="/cerca"`,
		`method="get"`,
		`name="q"`,
		`name="scope"`,
		`value="srd"`,
		`value="mappe"`,
		`value="generatori"`,
		`Giocare`,
		`Preparare`,
		`href="/giocare"`,
		`href="/preparare"`,
		`Crea un personaggio`,
		`Prepara un incontro`,
		`Tutto il compendio`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("expected landing page to contain %q", expected)
		}
	}
}

func TestLandingPageLinksToEveryCollection(t *testing.T) {
	var rendered bytes.Buffer
	if err := LandingPage().Render(context.Background(), &rendered); err != nil {
		t.Fatalf("render landing page: %v", err)
	}

	html := rendered.String()
	for _, collection := range []string{
		"classi",
		"specie",
		"backgrounds",
		"talenti",
		"regole",
		"glossario",
		"incantesimi",
		"equipaggiamenti",
		"oggetti_magici",
		"servizi",
		"mostri",
	} {
		if !strings.Contains(html, `/srd/`+collection) {
			t.Errorf("expected landing page to link collection %q", collection)
		}
	}
}
