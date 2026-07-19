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
