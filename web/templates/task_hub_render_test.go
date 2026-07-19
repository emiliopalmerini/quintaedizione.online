package templates

import (
	"bytes"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestTaskHubsExposeCanonicalDestinations(t *testing.T) {
	tests := []struct {
		name      string
		component templ.Component
		expected  []string
	}{
		{"giocare", GiocarePage(), []string{"Giocare", "/srd/area/personaggi", "/srd/incantesimi", "/srd/regole", "/srd/equipaggiamenti"}},
		{"preparare", PrepararePage(), []string{"Preparare", "/combattimenti", "/srd/mostri", "/mappe", "/generatori"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rendered bytes.Buffer
			if err := tt.component.Render(t.Context(), &rendered); err != nil {
				t.Fatalf("render hub: %v", err)
			}
			for _, expected := range tt.expected {
				if !strings.Contains(rendered.String(), expected) {
					t.Errorf("expected hub to contain %q", expected)
				}
			}
		})
	}
}
