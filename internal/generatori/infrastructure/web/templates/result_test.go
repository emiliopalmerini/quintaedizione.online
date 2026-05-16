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
