package filters

import (
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/collections"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/filters"
)

type fakeAggregator map[string]map[string]map[string]int64 // collection → field → value → count

func (f fakeAggregator) Aggregate(coll, field string, _ func(map[string]any) bool) map[string]int64 {
	if c := f[coll]; c != nil {
		return c[field]
	}
	return nil
}

func TestDeriveEnumValues_CaseFold_MostCommonDisplayWins(t *testing.T) {
	registry := NewInMemoryFilterRegistry()
	registry.AddFilter(filters.FilterDefinition{
		Name:        "scuola",
		FieldPath:   "scuola",
		Collections: []collections.CollectionName{collections.Incantesimi},
	})
	DeriveEnumValues(registry, fakeAggregator{
		"incantesimi": {"scuola": {"Necromanzia": 10, "necromanzia": 1, "Evocazione": 5}},
	})

	def, _ := registry.GetFilterByName("scuola")
	if len(def.EnumValues) != 2 {
		t.Fatalf("expected 2 distinct values after case-folding, got %v", def.EnumValues)
	}
	// Most-common value first (sorted by count desc).
	if def.EnumValues[0] != "Necromanzia" {
		t.Errorf("expected 'Necromanzia' first (higher count), got %s", def.EnumValues[0])
	}
}

func TestDeriveEnumValues_NumericSort(t *testing.T) {
	registry := NewInMemoryFilterRegistry()
	registry.AddFilter(filters.FilterDefinition{
		Name:        "livello",
		FieldPath:   "livello",
		Collections: []collections.CollectionName{collections.Incantesimi},
	})
	DeriveEnumValues(registry, fakeAggregator{
		"incantesimi": {"livello": {"3": 1, "0": 1, "9": 1, "1": 1}},
	})
	def, _ := registry.GetFilterByName("livello")
	want := []string{"0", "1", "3", "9"}
	for i, w := range want {
		if def.EnumValues[i] != w {
			t.Errorf("at index %d: want %s, got %s (full: %v)", i, w, def.EnumValues[i], def.EnumValues)
		}
	}
}

func TestDeriveEnumValues_FractionalCRSort(t *testing.T) {
	registry := NewInMemoryFilterRegistry()
	registry.AddFilter(filters.FilterDefinition{
		Name:        "grado_sfida",
		FieldPath:   "grado_sfida",
		Collections: []collections.CollectionName{collections.Mostri},
	})
	DeriveEnumValues(registry, fakeAggregator{
		"mostri": {"grado_sfida": {"1": 1, "1/2": 1, "1/4": 1, "0": 1, "1/8": 1, "10": 1}},
	})
	def, _ := registry.GetFilterByName("grado_sfida")
	want := []string{"0", "1/8", "1/4", "1/2", "1", "10"}
	for i, w := range want {
		if def.EnumValues[i] != w {
			t.Errorf("at index %d: want %s, got %v", i, w, def.EnumValues)
		}
	}
}

func TestDeriveEnumValues_TitleCaseDisplay(t *testing.T) {
	registry := NewInMemoryFilterRegistry()
	registry.AddFilter(filters.FilterDefinition{
		Name:        "rarita",
		FieldPath:   "rarita",
		Collections: []collections.CollectionName{collections.OggettiMagici},
	})
	DeriveEnumValues(registry, fakeAggregator{
		"oggetti_magici": {"rarita": {"non comune": 3, "comune": 5}},
	})
	def, _ := registry.GetFilterByName("rarita")
	if def.EnumValues[0] != "Comune" {
		t.Errorf("expected title-cased 'Comune', got %s", def.EnumValues[0])
	}
	if def.EnumValues[1] != "Non Comune" {
		t.Errorf("expected title-cased 'Non Comune', got %s", def.EnumValues[1])
	}
}

func TestDeriveEnumValues_SkipsPreSetEnums(t *testing.T) {
	registry := NewInMemoryFilterRegistry()
	registry.AddFilter(filters.FilterDefinition{
		Name:        "_source_short",
		FieldPath:   "_source_short",
		EnumValues:  []string{"5.5e", "5e"},
		Description: "Edizione",
	})
	DeriveEnumValues(registry, fakeAggregator{})
	def, _ := registry.GetFilterByName("_source_short")
	if len(def.EnumValues) != 2 || def.EnumValues[0] != "5.5e" {
		t.Errorf("pre-set enums should be preserved, got %v", def.EnumValues)
	}
}
