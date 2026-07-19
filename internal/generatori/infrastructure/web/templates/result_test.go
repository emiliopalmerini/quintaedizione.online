package templates

import (
	"bytes"
	"strings"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/generatori/domain"
)

func TestResultDoesNotEmitInlineHandlers(t *testing.T) {
	var buf bytes.Buffer
	result := domain.RollResult{
		Entries: []domain.RollEntry{{Value: "Risultato", Link: "/srd/mostri/5.5e/risultato"}},
	}

	if err := Result(result).Render(t.Context(), &buf); err != nil {
		t.Fatalf("render result: %v", err)
	}
	body := buf.String()
	if strings.Contains(body, "onclick=") {
		t.Fatalf("result contains an inline handler: %s", body)
	}
}

func TestGeneratorRendersNativeRollFormAndLiveResult(t *testing.T) {
	var buf bytes.Buffer
	table := domain.Table{ID: "patroni", Name: "Patroni", Die: "1D2", Items: []domain.Item{{Text: "Uno"}, {Text: "Due"}}}

	if err := Generator(table, nil, nil, nil).Render(t.Context(), &buf); err != nil {
		t.Fatalf("render generator: %v", err)
	}
	body := buf.String()
	for _, expected := range []string{
		`action="/generatori/patroni/roll"`,
		`method="post"`,
		`hx-post="/generatori/patroni/roll"`,
		`type="submit"`,
		`id="roll-result"`,
		`role="status"`,
		`aria-live="polite"`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("expected generator page to contain %q, got:\n%s", expected, body)
		}
	}
}

func TestMultiColumnTableIsCompleteAndAccessible(t *testing.T) {
	var buf bytes.Buffer
	table := domain.Table{
		ID: "multi", Name: "Tabella multipla",
		Columns: []domain.Column{
			{Name: "Colore", Items: []domain.Item{{Text: "Rosso"}}},
			{Name: "Forma", Items: []domain.Item{{Text: "Cerchio"}, {Text: "Quadrato"}}},
		},
	}

	if err := multiColumnTable(table).Render(t.Context(), &buf); err != nil {
		t.Fatalf("render table: %v", err)
	}
	body := buf.String()
	for _, expected := range []string{
		`class="generator-table-scroll"`,
		`role="region"`,
		`tabindex="0"`,
		`<caption class="sr-only">Tabella multipla</caption>`,
		`<th scope="col">Tiro</th>`,
		`<th scope="row">2</th>`,
		`Quadrato`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("expected complete accessible table to contain %q, got:\n%s", expected, body)
		}
	}
}
