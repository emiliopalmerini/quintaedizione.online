package display

import (
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/dto"
)

func valuesByType(elements []dto.DisplayElementDTO) map[string]string {
	out := make(map[string]string, len(elements))
	for _, e := range elements {
		out[e.Type] = e.Value
	}
	return out
}

func TestMostriDisplayStrategyUsesStructuredFields(t *testing.T) {
	doc := &domain.Document{
		Fields: map[string]any{
			"ac":          "15",
			"hp":          "45",
			"grado_sfida": "2",
		},
	}

	elements := (&MostriDisplayStrategy{}).GetElements(doc)
	got := valuesByType(elements)

	if got["ac"] != "CA 15" {
		t.Fatalf("ac display = %q, want %q", got["ac"], "CA 15")
	}
	if got["hp"] != "PF 45" {
		t.Fatalf("hp display = %q, want %q", got["hp"], "PF 45")
	}
	if got["challenge_rating"] != "GS 2" {
		t.Fatalf("challenge rating display = %q, want %q", got["challenge_rating"], "GS 2")
	}
}

func TestBackgroundsDisplayStrategyMatchesCurrentMarkdownLabels(t *testing.T) {
	doc := &domain.Document{
		RawContent: domain.MarkdownContent("**Competenze nelle Abilita:** Furtivita\n\n**Talento:** Allerta"),
	}

	elements := (&BackgroundsDisplayStrategy{}).GetElements(doc)
	got := valuesByType(elements)

	if got["skills"] != "Furtivita" {
		t.Fatalf("skills display = %q, want %q", got["skills"], "Furtivita")
	}
	if got["feat"] != "Allerta" {
		t.Fatalf("feat display = %q, want %q", got["feat"], "Allerta")
	}
}

func TestTalentiDisplayStrategyUsesStructuredCategory(t *testing.T) {
	doc := &domain.Document{
		Fields: map[string]any{"categoria": "Origine"},
	}

	elements := (&TalentiDisplayStrategy{}).GetElements(doc)
	got := valuesByType(elements)

	if got["category"] != "Origine" {
		t.Fatalf("category display = %q, want %q", got["category"], "Origine")
	}
}
