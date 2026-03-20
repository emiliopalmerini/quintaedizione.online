package parsers

import (
	"strings"
	"testing"
	"testing/fstest"
)

func newTestGlossaryLinker(t *testing.T, glossaryJSON string) *GlossaryLinker {
	t.Helper()
	fsys := fstest.MapFS{
		"glossary.json": &fstest.MapFile{Data: []byte(glossaryJSON)},
	}
	gl, err := NewGlossaryLinker(fsys)
	if err != nil {
		t.Fatalf("NewGlossaryLinker failed: %v", err)
	}
	return gl
}

func TestGlossaryLinker_SimpleText(t *testing.T) {
	gl := newTestGlossaryLinker(t, `[
		{"id": "accecato", "term": "Accecato", "category": "condizione", "definition": "Non è in grado di vedere."}
	]`)

	input := "<p>La creatura è Accecato per 1 minuto.</p>"
	result := gl.LinkGlossaryTerms(input)

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

func TestGlossaryLinker_FirstOccurrenceOnly(t *testing.T) {
	gl := newTestGlossaryLinker(t, `[
		{"id": "accecato", "term": "Accecato", "category": "condizione", "definition": "Non è in grado di vedere."}
	]`)

	input := "<p>Accecato prima. Accecato seconda.</p>"
	result := gl.LinkGlossaryTerms(input)

	count := strings.Count(result, `data-term-id="accecato"`)
	if count != 1 {
		t.Errorf("Expected 1 glossary link (first occurrence only), got %d", count)
	}
}

func TestGlossaryLinker_SkipCodeBlocks(t *testing.T) {
	gl := newTestGlossaryLinker(t, `[
		{"id": "accecato", "term": "Accecato", "category": "condizione", "definition": "Non è in grado di vedere."}
	]`)

	input := "<pre><code>Accecato in code</code></pre><p>Accecato in text</p>"
	result := gl.LinkGlossaryTerms(input)

	count := strings.Count(result, `class="glossary-term"`)
	if count != 1 {
		t.Errorf("Expected 1 glossary link (skipping code block), got %d", count)
	}
}

func TestGlossaryLinker_SkipExistingLinks(t *testing.T) {
	gl := newTestGlossaryLinker(t, `[
		{"id": "accecato", "term": "Accecato", "category": "condizione", "definition": "Non è in grado di vedere."}
	]`)

	input := `<p><a href="/other">Accecato link</a> and Accecato text</p>`
	result := gl.LinkGlossaryTerms(input)

	// Should link the text occurrence (not the one inside <a>)
	count := strings.Count(result, `class="glossary-term"`)
	if count != 1 {
		t.Errorf("Expected 1 glossary link, got %d", count)
	}
}

func TestGlossaryLinker_SkipHeadings(t *testing.T) {
	gl := newTestGlossaryLinker(t, `[
		{"id": "accecato", "term": "Accecato", "category": "condizione", "definition": "Non è in grado di vedere."}
	]`)

	input := "<h2>Accecato</h2><p>Il personaggio è Accecato.</p>"
	result := gl.LinkGlossaryTerms(input)

	count := strings.Count(result, `class="glossary-term"`)
	if count != 1 {
		t.Errorf("Expected 1 glossary link (skipping heading), got %d", count)
	}
}

func TestGlossaryLinker_LongestMatchFirst(t *testing.T) {
	gl := newTestGlossaryLinker(t, `[
		{"id": "privo-di-sensi", "term": "Privo di Sensi", "category": "condizione", "definition": "Non è cosciente."},
		{"id": "prono", "term": "Prono", "category": "condizione", "definition": "Sdraiato a terra."}
	]`)

	input := "<p>La creatura è Privo di Sensi.</p>"
	result := gl.LinkGlossaryTerms(input)

	if !strings.Contains(result, `data-term-id="privo-di-sensi"`) {
		t.Error("Expected link to 'Privo di Sensi'")
	}
	if strings.Contains(result, `data-term-id="prono"`) {
		t.Error("Should not link 'Prono' inside 'Privo di Sensi'")
	}
}

func TestGlossaryLinker_EmptyInput(t *testing.T) {
	gl := newTestGlossaryLinker(t, `[
		{"id": "accecato", "term": "Accecato", "category": "condizione", "definition": "Non è in grado di vedere."}
	]`)

	result := gl.LinkGlossaryTerms("")
	if result != "" {
		t.Errorf("Expected empty result, got '%s'", result)
	}
}

func TestGlossaryLinker_TruncateDefinition(t *testing.T) {
	longDef := strings.Repeat("parola ", 50) // ~350 chars
	gl := newTestGlossaryLinker(t, `[
		{"id": "test", "term": "TestTerm", "category": "", "definition": "`+longDef+`"}
	]`)

	input := "<p>Il TestTerm è importante.</p>"
	result := gl.LinkGlossaryTerms(input)

	// Definition should be truncated
	if !strings.Contains(result, `data-term-def=`) {
		t.Error("Expected data-term-def attribute")
	}
	// The truncated definition should end with …
	if !strings.Contains(result, "…") {
		t.Error("Expected truncated definition to end with …")
	}
}
