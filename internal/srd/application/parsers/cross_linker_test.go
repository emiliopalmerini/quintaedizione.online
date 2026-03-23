package parsers

import (
	"strings"
	"testing"
	"testing/fstest"
)

func newTestCrossLinker(t *testing.T, glossaryJSON string) *CrossLinker {
	t.Helper()
	fsys := fstest.MapFS{
		"glossary.json": &fstest.MapFile{Data: []byte(glossaryJSON)},
	}
	cl, err := NewCrossLinker(fsys)
	if err != nil {
		t.Fatalf("NewCrossLinker failed: %v", err)
	}
	return cl
}

func TestCrossLinker_SimpleText(t *testing.T) {
	cl := newTestCrossLinker(t, `[
		{"id": "accecato", "term": "Accecato", "category": "condizione", "definition": [{"type":"text","text":"Non è in grado di vedere."}]}
	]`)

	input := "<p>La creatura è Accecato per 1 minuto.</p>"
	result := cl.LinkTerms(input)

	if !strings.Contains(result, `class="glossary-term"`) {
		t.Error("Expected glossary-term class in output")
	}
	if !strings.Contains(result, `data-term-id="accecato"`) {
		t.Error("Expected data-term-id attribute")
	}
	if !strings.Contains(result, `data-term-def=`) {
		t.Error("Expected data-term-def attribute")
	}
	if !strings.Contains(result, `data-term-cat="condizione"`) {
		t.Error("Expected data-term-cat attribute")
	}
	if !strings.Contains(result, `tabindex="0"`) {
		t.Error("Expected tabindex attribute")
	}
}

func TestCrossLinker_FirstOccurrenceOnly(t *testing.T) {
	cl := newTestCrossLinker(t, `[
		{"id": "accecato", "term": "Accecato", "category": "condizione", "definition": [{"type":"text","text":"Non è in grado di vedere."}]}
	]`)

	input := "<p>Accecato prima. Accecato seconda.</p>"
	result := cl.LinkTerms(input)

	count := strings.Count(result, `data-term-id="accecato"`)
	if count != 1 {
		t.Errorf("Expected 1 glossary link (first occurrence only), got %d", count)
	}
}

func TestCrossLinker_SkipCodeBlocks(t *testing.T) {
	cl := newTestCrossLinker(t, `[
		{"id": "accecato", "term": "Accecato", "category": "condizione", "definition": [{"type":"text","text":"Non è in grado di vedere."}]}
	]`)

	input := "<pre><code>Accecato in code</code></pre><p>Accecato in text</p>"
	result := cl.LinkTerms(input)

	count := strings.Count(result, `class="glossary-term"`)
	if count != 1 {
		t.Errorf("Expected 1 glossary link (skipping code block), got %d", count)
	}
}

func TestCrossLinker_SkipExistingLinks(t *testing.T) {
	cl := newTestCrossLinker(t, `[
		{"id": "accecato", "term": "Accecato", "category": "condizione", "definition": [{"type":"text","text":"Non è in grado di vedere."}]}
	]`)

	input := `<p><a href="/other">Accecato link</a> and Accecato text</p>`
	result := cl.LinkTerms(input)

	count := strings.Count(result, `class="glossary-term"`)
	if count != 1 {
		t.Errorf("Expected 1 glossary link, got %d", count)
	}
}

func TestCrossLinker_SkipHeadings(t *testing.T) {
	cl := newTestCrossLinker(t, `[
		{"id": "accecato", "term": "Accecato", "category": "condizione", "definition": [{"type":"text","text":"Non è in grado di vedere."}]}
	]`)

	input := "<h2>Accecato</h2><p>Il personaggio è Accecato.</p>"
	result := cl.LinkTerms(input)

	count := strings.Count(result, `class="glossary-term"`)
	if count != 1 {
		t.Errorf("Expected 1 glossary link (skipping heading), got %d", count)
	}
}

func TestCrossLinker_LongestMatchFirst(t *testing.T) {
	cl := newTestCrossLinker(t, `[
		{"id": "privo-di-sensi", "term": "Privo di Sensi", "category": "condizione", "definition": [{"type":"text","text":"Non è cosciente."}]},
		{"id": "prono", "term": "Prono", "category": "condizione", "definition": [{"type":"text","text":"Sdraiato a terra."}]}
	]`)

	input := "<p>La creatura è Privo di Sensi.</p>"
	result := cl.LinkTerms(input)

	if !strings.Contains(result, `data-term-id="privo-di-sensi"`) {
		t.Error("Expected link to 'Privo di Sensi'")
	}
	if strings.Contains(result, `data-term-id="prono"`) {
		t.Error("Should not link 'Prono' inside 'Privo di Sensi'")
	}
}

func TestCrossLinker_EmptyInput(t *testing.T) {
	cl := newTestCrossLinker(t, `[
		{"id": "accecato", "term": "Accecato", "category": "condizione", "definition": [{"type":"text","text":"Non è in grado di vedere."}]}
	]`)

	result := cl.LinkTerms("")
	if result != "" {
		t.Errorf("Expected empty result, got '%s'", result)
	}
}

func TestCrossLinker_TruncateDefinition(t *testing.T) {
	longDef := strings.Repeat("parola ", 50) // ~350 chars
	cl := newTestCrossLinker(t, `[
		{"id": "test", "term": "TestTerm", "category": "", "definition": [{"type":"text","text":"`+longDef+`"}]}
	]`)

	input := "<p>Il TestTerm è importante.</p>"
	result := cl.LinkTerms(input)

	if !strings.Contains(result, `data-term-def=`) {
		t.Error("Expected data-term-def attribute")
	}
	if !strings.Contains(result, "…") {
		t.Error("Expected truncated definition to end with …")
	}
}

func TestCrossLinker_ConvertsInternalLinksToSpans(t *testing.T) {
	cl := newTestCrossLinker(t, `[
		{"id": "accecato", "term": "Accecato", "category": "condizione", "definition": [{"type":"text","text":"Non è in grado di vedere."}]}
	]`)

	// Data-driven <a> link to /srd/ should be converted to glossary-term span
	input := `<p>Se è <a href="/srd/glossario/5.5e/accecato">accecato</a>, la magia fallisce.</p>`
	result := cl.LinkTerms(input)

	// Should be a span, not an <a> link
	if strings.Contains(result, "<a ") {
		t.Error("Internal /srd/ link should be converted to span, not kept as <a>")
	}
	if !strings.Contains(result, `class="glossary-term"`) {
		t.Error("Expected glossary-term span")
	}
	if !strings.Contains(result, `data-term-link="/srd/glossario/5.5e/accecato"`) {
		t.Error("Expected data-term-link attribute with original href")
	}
	if !strings.Contains(result, `data-term-id="accecato"`) {
		t.Error("Expected data-term-id extracted from URL")
	}
	if !strings.Contains(result, `data-term-cat="condizione"`) {
		t.Error("Expected data-term-cat from collection")
	}
	// Should also get data-term-def merged from glossary
	if !strings.Contains(result, `data-term-def=`) {
		t.Error("Expected data-term-def merged from glossary")
	}
}

func TestCrossLinker_InternalLinkWithoutGlossaryEntry(t *testing.T) {
	cl := newTestCrossLinker(t, `[]`)

	// Spell link that has no glossary entry
	input := `<p>Lancia <a href="/srd/incantesimi/5.5e/palla-di-fuoco">palla di fuoco</a>.</p>`
	result := cl.LinkTerms(input)

	// Should still be converted to span (no data-term-def since no glossary entry)
	if strings.Contains(result, "<a ") {
		t.Error("Internal /srd/ link should be converted to span")
	}
	if !strings.Contains(result, `data-term-link="/srd/incantesimi/5.5e/palla-di-fuoco"`) {
		t.Error("Expected data-term-link")
	}
	if !strings.Contains(result, `data-term-id="palla-di-fuoco"`) {
		t.Error("Expected data-term-id from URL")
	}
	if !strings.Contains(result, `data-term-cat="incantesimo"`) {
		t.Error("Expected data-term-cat from collection")
	}
}

func TestCrossLinker_ExternalLinksUntouched(t *testing.T) {
	cl := newTestCrossLinker(t, `[]`)

	input := `<p>Vedi <a href="https://example.com">esempio</a>.</p>`
	result := cl.LinkTerms(input)

	if !strings.Contains(result, `<a href="https://example.com">`) {
		t.Error("External links should be left untouched")
	}
}

func TestCrossLinker_RuleLinkCategory(t *testing.T) {
	cl := newTestCrossLinker(t, `[]`)

	input := `<p>Tiro su <a href="/srd/regole/5.5e/le-sei-caratteristiche">Costituzione</a>.</p>`
	result := cl.LinkTerms(input)

	if !strings.Contains(result, `data-term-cat="regola"`) {
		t.Error("Expected data-term-cat='regola' for rules link")
	}
	if !strings.Contains(result, `data-term-id="le-sei-caratteristiche"`) {
		t.Error("Expected data-term-id from URL")
	}
}

func TestCrossLinker_HomogeneousOutput(t *testing.T) {
	cl := newTestCrossLinker(t, `[
		{"id": "accecato", "term": "Accecato", "category": "condizione", "definition": [{"type":"text","text":"Non è in grado di vedere."}]},
		{"id": "tiro-salvezza", "term": "tiro salvezza", "category": "meccanica", "definition": [{"type":"text","text":"Un tiro per resistere."}]}
	]`)

	// Mix of data-driven link + glossary term → both should be <span class="glossary-term">
	input := `<p>Se è <a href="/srd/glossario/5.5e/accecato">accecato</a>, effettua un tiro salvezza.</p>`
	result := cl.LinkTerms(input)

	// No <a> links should remain (internal ones converted)
	if strings.Contains(result, "<a ") {
		t.Error("All internal links should be converted to spans")
	}

	// Both should be glossary-term spans
	count := strings.Count(result, `class="glossary-term"`)
	if count != 2 {
		t.Errorf("Expected 2 glossary-term spans (homogeneous), got %d\n%s", count, result)
	}
}

func TestCrossLinker_LinkConversionPreventsDoubleGlossary(t *testing.T) {
	cl := newTestCrossLinker(t, `[
		{"id": "accecato", "term": "Accecato", "category": "condizione", "definition": [{"type":"text","text":"Non è in grado di vedere."}]}
	]`)

	// "accecato" appears as a link AND would match glossary regex
	input := `<p>Se è <a href="/srd/glossario/5.5e/accecato">accecato</a>, e poi accecato di nuovo.</p>`
	result := cl.LinkTerms(input)

	// Should have exactly 1 glossary-term span (link converted, second occurrence skipped)
	count := strings.Count(result, `data-term-id="accecato"`)
	if count != 1 {
		t.Errorf("Expected 1 occurrence (link converted, no duplicate), got %d\n%s", count, result)
	}
}

func TestCrossLinker_SkipExistingSpans(t *testing.T) {
	cl := newTestCrossLinker(t, `[
		{"id": "forza", "term": "Forza", "category": "caratteristica", "definition": [{"type":"text","text":"La caratteristica Forza."}]}
	]`)

	// Simulate HTML with an already-enriched span (from a previous processing pass)
	input := `<p><span class="glossary-term" data-term-id="forza">Forza</span> e Forza di nuovo.</p>`
	result := cl.LinkTerms(input)

	// Should not nest spans
	count := strings.Count(result, `class="glossary-term"`)
	if count != 1 {
		t.Errorf("Expected 1 glossary-term span (no nesting), got %d", count)
	}
}
