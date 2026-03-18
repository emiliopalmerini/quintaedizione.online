package filters

import (
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/collections"
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/filters"
	"github.com/emiliopalmerini/quintaedizione.online/internal/infrastructure/datastore"
)

func TestRegisterEditionFilter_SingleSource(t *testing.T) {
	registry := NewInMemoryFilterRegistry()
	sources := []datastore.Source{
		{ID: "srd-5.5e", ShortName: "5.5e", Default: true},
	}

	RegisterEditionFilter(registry, sources)

	// With a single source, the edition filter should NOT be registered
	defs, err := registry.GetFiltersForCollection(collections.Incantesimi)
	if err != nil {
		t.Fatalf("GetFiltersForCollection failed: %v", err)
	}

	for _, def := range defs {
		if def.Name == "_source_short" {
			t.Error("edition filter should not be registered with a single source")
		}
	}
}

func TestRegisterEditionFilter_MultipleSources(t *testing.T) {
	registry := NewInMemoryFilterRegistry()
	sources := []datastore.Source{
		{ID: "srd-5.5e", ShortName: "5.5e", Default: true},
		{ID: "srd-5e", ShortName: "5e", Default: false},
	}

	RegisterEditionFilter(registry, sources)

	// With multiple sources, the edition filter should be registered for all collections
	for _, coll := range collections.GetAllCollectionNames() {
		collName, ok := collections.FromString(coll)
		if !ok {
			continue
		}
		defs, err := registry.GetFiltersForCollection(collName)
		if err != nil {
			t.Fatalf("GetFiltersForCollection(%s) failed: %v", coll, err)
		}

		found := false
		for _, def := range defs {
			if def.Name == "_source_short" {
				found = true

				// Should have enum values matching source short names
				if len(def.EnumValues) != 2 {
					t.Errorf("expected 2 enum values, got %d", len(def.EnumValues))
				}

				// Should apply to all collections (empty Collections slice)
				if len(def.Collections) != 0 {
					t.Errorf("expected empty Collections (applies to all), got %d", len(def.Collections))
				}

				break
			}
		}

		if !found {
			t.Errorf("edition filter not found for collection %s", coll)
		}
	}
}

func TestEditionFilter_PredicateMatches(t *testing.T) {
	registry := NewInMemoryFilterRegistry()
	sources := []datastore.Source{
		{ID: "srd-5.5e", ShortName: "5.5e"},
		{ID: "srd-5e", ShortName: "5e"},
	}
	RegisterEditionFilter(registry, sources)

	def, ok := registry.GetFilterByName("_source_short")
	if !ok {
		t.Fatal("edition filter not found")
	}

	// Create a predicate builder to test filtering
	pb := NewPredicateBuilder()
	filterSet := filters.NewFilterSet(collections.Incantesimi)
	filterSet.AddFilter(filters.FilterValue{
		Definition: def,
		Value:      "5.5e",
	})

	predicate, err := pb.BuildPredicate(filterSet)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Should match 5.5e documents
	doc55 := map[string]any{"_source_short": "5.5e", "title": "Palla di Fuoco"}
	if !predicate(doc55) {
		t.Error("expected predicate to match 5.5e document")
	}

	// Should NOT match 5e documents
	doc5 := map[string]any{"_source_short": "5e", "title": "Palla di Fuoco"}
	if predicate(doc5) {
		t.Error("expected predicate to NOT match 5e document")
	}
}
