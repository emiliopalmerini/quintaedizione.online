package filters

import (
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/collections"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/filters"
)

func mustPred(t *testing.T, defs []filters.FilterValue) filters.DocumentPredicate {
	t.Helper()
	fs := filters.NewFilterSet(collections.Incantesimi)
	for _, d := range defs {
		fs.AddFilter(d)
	}
	p, err := NewPredicateBuilder().BuildPredicate(fs)
	if err != nil {
		t.Fatalf("BuildPredicate: %v", err)
	}
	return p
}

func TestPredicate_ScalarField_CaseInsensitive(t *testing.T) {
	p := mustPred(t, []filters.FilterValue{{
		Definition: filters.FilterDefinition{Name: "scuola", FieldPath: "scuola"},
		Value:      "Necromanzia",
	}})
	if !p(map[string]any{"scuola": "necromanzia"}) {
		t.Error("expected case-insensitive match against lowercase data")
	}
	if p(map[string]any{"scuola": "Evocazione"}) {
		t.Error("unexpected match against unrelated value")
	}
}

func TestPredicate_SliceField_AnyElementMatches(t *testing.T) {
	p := mustPred(t, []filters.FilterValue{{
		Definition: filters.FilterDefinition{Name: "classe", FieldPath: "classe"},
		Value:      "Mago",
	}})
	doc := map[string]any{"classe": []any{"bardo", "mago", "chierico"}}
	if !p(doc) {
		t.Error("expected match against slice element")
	}
	if p(map[string]any{"classe": []any{"druido"}}) {
		t.Error("unexpected match against slice without target")
	}
}

func TestPredicate_CommaSeparatedValues_OR(t *testing.T) {
	p := mustPred(t, []filters.FilterValue{{
		Definition: filters.FilterDefinition{Name: "classe", FieldPath: "classe"},
		Value:      "Bardo,Mago",
	}})
	if !p(map[string]any{"classe": []any{"bardo"}}) {
		t.Error("expected OR match for first value")
	}
	if !p(map[string]any{"classe": []any{"mago"}}) {
		t.Error("expected OR match for second value")
	}
	if p(map[string]any{"classe": []any{"druido"}}) {
		t.Error("unexpected match for unlisted value")
	}
}

func TestPredicate_MultipleFilters_AND(t *testing.T) {
	p := mustPred(t, []filters.FilterValue{
		{
			Definition: filters.FilterDefinition{Name: "scuola", FieldPath: "scuola"},
			Value:      "Necromanzia",
		},
		{
			Definition: filters.FilterDefinition{Name: "livello", FieldPath: "livello"},
			Value:      "3",
		},
	})
	if !p(map[string]any{"scuola": "Necromanzia", "livello": 3}) {
		t.Error("expected both conditions to match")
	}
	if p(map[string]any{"scuola": "Necromanzia", "livello": 4}) {
		t.Error("unexpected match when one condition fails")
	}
}

func TestPredicate_NumericFieldFormatting(t *testing.T) {
	p := mustPred(t, []filters.FilterValue{{
		Definition: filters.FilterDefinition{Name: "livello", FieldPath: "livello"},
		Value:      "0",
	}})
	if !p(map[string]any{"livello": 0}) {
		t.Error("expected match for int 0 vs string \"0\"")
	}
}

func TestPredicate_NilOnEmptyValue(t *testing.T) {
	fs := filters.NewFilterSet(collections.Incantesimi)
	fs.AddFilter(filters.FilterValue{
		Definition: filters.FilterDefinition{Name: "scuola", FieldPath: "scuola"},
		Value:      "",
	})
	p, err := NewPredicateBuilder().BuildPredicate(fs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if p != nil {
		t.Error("expected nil predicate when no effective values")
	}
}
