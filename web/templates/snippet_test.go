package templates

import "testing"

func TestExtractSnippet_MatchInMiddle(t *testing.T) {
	body := "Questo incantesimo crea una sfera di fuoco che esplode in un punto a scelta dell'incantatore entro gittata. Ogni creatura nell'area deve effettuare un tiro salvezza su Destrezza."
	got := ExtractSnippet(body, "fuoco", 100)

	if got == "" {
		t.Fatal("expected a non-empty snippet")
	}
	if len(got) > 120 {
		t.Errorf("snippet too long: %d chars", len(got))
	}
	if !containsCaseInsensitive(got, "fuoco") {
		t.Errorf("snippet should contain the query term, got: %s", got)
	}
}

func TestExtractSnippet_MatchAtStart(t *testing.T) {
	body := "Fuoco e fiamme avvolgono il bersaglio per tutta la durata."
	got := ExtractSnippet(body, "fuoco", 100)

	if got == "" {
		t.Fatal("expected a non-empty snippet")
	}
	if !containsCaseInsensitive(got, "fuoco") {
		t.Errorf("snippet should contain the query, got: %s", got)
	}
}

func TestExtractSnippet_NoMatch(t *testing.T) {
	body := "Questo incantesimo non contiene la parola cercata."
	got := ExtractSnippet(body, "drago", 100)

	if got != "" {
		t.Errorf("expected empty snippet for no match, got: %s", got)
	}
}

func TestExtractSnippet_EmptyBody(t *testing.T) {
	got := ExtractSnippet("", "fuoco", 100)
	if got != "" {
		t.Errorf("expected empty snippet for empty body, got: %s", got)
	}
}

func TestExtractSnippet_EmptyQuery(t *testing.T) {
	got := ExtractSnippet("some body text", "", 100)
	if got != "" {
		t.Errorf("expected empty snippet for empty query, got: %s", got)
	}
}

func TestExtractSnippet_HTMLStripped(t *testing.T) {
	body := "<p>Una <strong>palla di fuoco</strong> esplode nell'area.</p>"
	got := ExtractSnippet(body, "fuoco", 100)

	if got == "" {
		t.Fatal("expected a non-empty snippet")
	}
	if containsHTML(got) {
		t.Errorf("snippet should not contain HTML tags, got: %s", got)
	}
	if !containsCaseInsensitive(got, "fuoco") {
		t.Errorf("snippet should contain the query, got: %s", got)
	}
}

func TestExtractSnippet_MarkdownStripped(t *testing.T) {
	body := "**Palla di Fuoco.** _Evocazione di 3° livello._ Una sfera di fuoco luminosa."
	got := ExtractSnippet(body, "fuoco", 100)

	if got == "" {
		t.Fatal("expected a non-empty snippet")
	}
	// Should not contain markdown formatting characters
	if containsSubstring(got, "**") || containsSubstring(got, "__") {
		t.Errorf("snippet should not contain markdown bold, got: %s", got)
	}
}

func TestExtractSnippet_TruncationWithEllipsis(t *testing.T) {
	body := "Prefisso lungo che precede il termine. La parola fuoco appare qui nel mezzo del testo. E poi c'è ancora molto altro testo che segue dopo il termine cercato per testare il troncamento."
	got := ExtractSnippet(body, "fuoco", 60)

	if got == "" {
		t.Fatal("expected non-empty snippet")
	}
	if len(got) > 80 { // allow some margin for ellipsis and word boundaries
		t.Errorf("snippet should be roughly maxLen, got %d chars: %s", len(got), got)
	}
}

func TestExtractSnippet_CaseInsensitive(t *testing.T) {
	body := "Una PALLA DI FUOCO esplode nell'area."
	got := ExtractSnippet(body, "fuoco", 100)

	if got == "" {
		t.Fatal("expected a non-empty snippet")
	}
}

// helpers

func containsCaseInsensitive(s, sub string) bool {
	return len(s) > 0 && len(sub) > 0 &&
		len(s) >= len(sub) &&
		(containsSubstring(toLowerCase(s), toLowerCase(sub)))
}

func toLowerCase(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func containsHTML(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '<' {
			for j := i + 1; j < len(s); j++ {
				if s[j] == '>' {
					return true
				}
			}
		}
	}
	return false
}
