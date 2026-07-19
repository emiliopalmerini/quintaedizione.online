package datastore

import (
	"strings"
	"testing"
)

func TestAddDescriptionsEnrichesGlossaryLinksAndLegacySpans(t *testing.T) {
	index := map[string]string{"5.5e/luce": "Illumina una zona."}

	for _, input := range []string{
		`<p><a href="/srd/incantesimi/5.5e/luce" class="glossary-term" data-term-link="/srd/incantesimi/5.5e/luce">Luce</a></p>`,
		`<p><span class="glossary-term" data-term-link="/srd/incantesimi/5.5e/luce">Luce</span></p>`,
	} {
		got := addDescriptionsToSpans(input, index)
		if !strings.Contains(got, `data-term-def="Illumina una zona."`) {
			t.Errorf("expected glossary element to be enriched, got %s", got)
		}
	}
}

func TestBuildPreviewUsesFirstMonsterTrait(t *testing.T) {
	doc := map[string]any{
		"_stat_block": "monster",
		"raw_content": "---\n\n**CA** 19\n\n| FOR | DES |\n|:---:|:---:|",
		"traits": []map[string]any{
			{
				"name":             "Resistenza Leggendaria",
				"description_html": "<p>Se il drago fallisce un <strong>tiro salvezza</strong>, può invece superarlo.</p>",
			},
		},
	}

	got := BuildPreview(doc)
	want := "Resistenza Leggendaria. Se il drago fallisce un tiro salvezza, può invece superarlo."
	if got != want {
		t.Fatalf("expected monster trait preview %q, got %q", want, got)
	}
}

func TestBuildPreviewUsesMonsterActionWhenTraitsAreEmpty(t *testing.T) {
	doc := map[string]any{
		"_stat_block": "monster",
		"raw_content": "---\n\n**CA** 19\n\n| FOR | DES |\n|:---:|:---:|",
		"traits":      []map[string]any{},
		"actions": []map[string]any{
			{
				"name":             "Morso",
				"description_html": "<p>Attacco con arma da mischia: +7 al tiro per colpire.</p>",
			},
		},
	}

	got := BuildPreview(doc)
	want := "Morso. Attacco con arma da mischia: +7 al tiro per colpire."
	if got != want {
		t.Fatalf("expected monster action preview %q, got %q", want, got)
	}
}
