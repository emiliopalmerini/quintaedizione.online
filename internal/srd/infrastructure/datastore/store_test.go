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
