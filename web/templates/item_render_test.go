package templates

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/models"
)

func TestItemPageUsesNativeDocumentNavigationSemantics(t *testing.T) {
	data := models.ItemPageData{
		PageData:        models.PageData{Collection: "incantesimi", DocTitle: "Luce", Title: "Luce"},
		CollectionLabel: "Incantesimi",
		Versions: []models.VersionTab{
			{Label: "5e", URL: "/srd/incantesimi/5e/luce"},
			{Label: "5.5e", URL: "/srd/incantesimi/5.5e/luce", IsCurrent: true},
		},
	}
	var rendered bytes.Buffer
	if err := ItemPage(data).Render(context.Background(), &rendered); err != nil {
		t.Fatalf("render item page: %v", err)
	}
	html := rendered.String()

	for _, expected := range []string{
		`aria-label="Navigazione elemento"`,
		`<h1 id="item-title" class="item-title" tabindex="-1">Luce</h1>`,
		`aria-label="Versione delle regole"`,
		`aria-current="page"`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("expected item page to contain %q, got:\n%s", expected, html)
		}
	}
	if strings.Contains(html, `role="tab`) || strings.Contains(html, `aria-selected=`) {
		t.Error("edition links are page navigation, not an ARIA tab widget")
	}
}

func TestMonsterSectionsFollowHeadingHierarchy(t *testing.T) {
	monster := models.MonsterStatBlock{Sections: []models.FeatureSection{{Heading: "Azioni"}}}
	var rendered bytes.Buffer
	if err := MonsterStatBlock(monster).Render(context.Background(), &rendered); err != nil {
		t.Fatalf("render monster stat block: %v", err)
	}

	if !strings.Contains(rendered.String(), `<h2 class="stat-block-section-heading">Azioni</h2>`) {
		t.Errorf("expected monster section heading to use h2, got %s", rendered.String())
	}
}
