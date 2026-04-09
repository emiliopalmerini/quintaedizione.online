package datastore

import (
	"sort"
	"testing"
)

func newTestStore(data map[string][]map[string]any) *Store {
	return NewStore(data)
}

func TestGetBySlug_MultipleVersions(t *testing.T) {
	store := newTestStore(map[string][]map[string]any{
		"incantesimi": {
			{"_id": "palla-di-fuoco", "_source_short": "5.5e", "title": "Palla di Fuoco"},
			{"_id": "palla-di-fuoco", "_source_short": "5e", "title": "Palla di Fuoco"},
		},
	})

	docs := store.GetBySlug("incantesimi", "palla-di-fuoco")
	if len(docs) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(docs))
	}

	// Verify both sources are present
	sources := make([]string, len(docs))
	for i, doc := range docs {
		sources[i] = doc["_source_short"].(string)
	}
	sort.Strings(sources)
	if sources[0] != "5.5e" || sources[1] != "5e" {
		t.Errorf("expected sources [5.5e, 5e], got %v", sources)
	}
}

func TestGetBySlug_SingleVersion(t *testing.T) {
	store := newTestStore(map[string][]map[string]any{
		"incantesimi": {
			{"_id": "luce", "_source_short": "5.5e", "title": "Luce"},
		},
	})

	docs := store.GetBySlug("incantesimi", "luce")
	if len(docs) != 1 {
		t.Fatalf("expected 1 version, got %d", len(docs))
	}
	if docs[0]["_id"].(string) != "luce" {
		t.Errorf("expected _id=luce, got %v", docs[0]["_id"])
	}
}

func TestGetBySlug_NotFound(t *testing.T) {
	store := newTestStore(map[string][]map[string]any{
		"incantesimi": {
			{"_id": "luce", "_source_short": "5.5e", "title": "Luce"},
		},
	})

	docs := store.GetBySlug("incantesimi", "non-esiste")
	if len(docs) != 0 {
		t.Errorf("expected 0 versions for unknown slug, got %d", len(docs))
	}
}

func TestGetBySlug_NoSourcePrefix(t *testing.T) {
	store := newTestStore(map[string][]map[string]any{
		"regole": {
			{"_id": "combat-rules", "title": "Combat Rules"},
		},
	})

	docs := store.GetBySlug("regole", "combat-rules")
	if len(docs) != 1 {
		t.Fatalf("expected 1 version for bare key, got %d", len(docs))
	}
	if docs[0]["_id"].(string) != "combat-rules" {
		t.Errorf("expected _id=combat-rules, got %v", docs[0]["_id"])
	}
}

func TestGetBySlug_UnknownCollection(t *testing.T) {
	store := newTestStore(map[string][]map[string]any{})

	docs := store.GetBySlug("nonexistent", "anything")
	if len(docs) != 0 {
		t.Errorf("expected 0 for unknown collection, got %d", len(docs))
	}
}

func TestDeduplicatePredicate_FiltersNonPreferredDuplicates(t *testing.T) {
	store := newTestStore(map[string][]map[string]any{
		"classi": {
			{"_id": "barbaro", "_source_short": "5.5e", "title": "Barbaro"},
			{"_id": "barbaro", "_source_short": "5e", "title": "Barbaro"},
			{"_id": "ladro", "_source_short": "5.5e", "title": "Ladro"},
		},
	})

	pred := store.DeduplicatePredicate("classi", "5.5e")
	if pred == nil {
		t.Fatal("expected non-nil predicate")
	}

	// 5.5e barbaro: preferred source, should pass
	if !pred(map[string]any{"_id": "barbaro", "_source_short": "5.5e"}) {
		t.Error("preferred source doc should pass")
	}
	// 5e barbaro: non-preferred, has duplicate -> should be filtered out
	if pred(map[string]any{"_id": "barbaro", "_source_short": "5e"}) {
		t.Error("non-preferred duplicate should be filtered out")
	}
	// 5.5e ladro: only version, should pass
	if !pred(map[string]any{"_id": "ladro", "_source_short": "5.5e"}) {
		t.Error("single-version doc should pass")
	}
}

func TestDeduplicatePredicate_KeepsUniqueNonPreferred(t *testing.T) {
	store := newTestStore(map[string][]map[string]any{
		"classi": {
			{"_id": "monaco", "_source_short": "5e", "title": "Monaco"},
		},
	})

	pred := store.DeduplicatePredicate("classi", "5.5e")

	// 5e monaco: not preferred source, but only version -> should pass
	if !pred(map[string]any{"_id": "monaco", "_source_short": "5e"}) {
		t.Error("unique non-preferred doc should pass")
	}
}

func TestDeduplicatePredicate_EmptyPreferred(t *testing.T) {
	store := newTestStore(map[string][]map[string]any{
		"classi": {
			{"_id": "barbaro", "_source_short": "5.5e", "title": "Barbaro"},
			{"_id": "barbaro", "_source_short": "5e", "title": "Barbaro"},
		},
	})

	// Empty preferred source -> no dedup, all pass
	pred := store.DeduplicatePredicate("classi", "")
	if pred != nil {
		t.Error("expected nil predicate for empty preferred source")
	}
}
