package persistence

import (
	"context"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/infrastructure/datastore"
)

func TestGetSearchableItems_DeduplicatesMultiSourceDocuments(t *testing.T) {
	store := datastore.NewStore(map[string][]map[string]any{
		"incantesimi": {
			{"_id": "palla-di-fuoco", "_source_short": "5.5e", "_source": "srd-5.5e", "title": "Palla di Fuoco"},
			{"_id": "palla-di-fuoco", "_source_short": "5e", "_source": "srd-5e", "title": "Palla di Fuoco"},
			{"_id": "luce", "_source_short": "5.5e", "_source": "srd-5.5e", "title": "Luce"},
		},
	})

	repo := NewSearchRepository(store, "5.5e")

	items, err := repo.GetSearchableItems(context.Background(), "incantesimi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return 2 items (palla-di-fuoco deduplicated to preferred 5.5e, plus luce)
	if len(items) != 2 {
		t.Fatalf("expected 2 items after dedup, got %d", len(items))
	}

	titles := make(map[string]bool)
	for _, item := range items {
		titles[item.Title] = true
	}
	if !titles["Palla di Fuoco"] {
		t.Error("expected Palla di Fuoco in results")
	}
	if !titles["Luce"] {
		t.Error("expected Luce in results")
	}
}

func TestGetSearchableItems_NoDedup_WhenNoPreferredSource(t *testing.T) {
	store := datastore.NewStore(map[string][]map[string]any{
		"incantesimi": {
			{"_id": "palla-di-fuoco", "_source_short": "5.5e", "_source": "srd-5.5e", "title": "Palla di Fuoco"},
			{"_id": "palla-di-fuoco", "_source_short": "5e", "_source": "srd-5e", "title": "Palla di Fuoco"},
		},
	})

	repo := NewSearchRepository(store, "")

	items, err := repo.GetSearchableItems(context.Background(), "incantesimi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No preferred source means no dedup; both versions should appear
	if len(items) != 2 {
		t.Fatalf("expected 2 items without dedup, got %d", len(items))
	}
}
