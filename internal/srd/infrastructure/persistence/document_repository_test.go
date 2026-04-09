package persistence

import (
	"context"
	"sort"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/infrastructure/datastore"
)

func TestFindVersions_ReturnsAllVersionInfos(t *testing.T) {
	store := datastore.NewStore(map[string][]map[string]any{
		"incantesimi": {
			{"_id": "palla-di-fuoco", "_source_short": "5.5e", "_source": "srd-5.5e", "title": "Palla di Fuoco"},
			{"_id": "palla-di-fuoco", "_source_short": "5e", "_source": "srd-5e", "title": "Palla di Fuoco"},
		},
	})
	repo := NewDocumentRepository(store)

	versions, err := repo.FindVersions(context.Background(), "incantesimi", "palla-di-fuoco")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}

	sources := make([]string, len(versions))
	for i, v := range versions {
		sources[i] = v.SourceShort
	}
	sort.Strings(sources)
	if sources[0] != "5.5e" || sources[1] != "5e" {
		t.Errorf("expected sources [5.5e, 5e], got %v", sources)
	}
}

func TestFindVersions_SingleVersion(t *testing.T) {
	store := datastore.NewStore(map[string][]map[string]any{
		"incantesimi": {
			{"_id": "luce", "_source_short": "5.5e", "_source": "srd-5.5e", "title": "Luce"},
		},
	})
	repo := NewDocumentRepository(store)

	versions, err := repo.FindVersions(context.Background(), "incantesimi", "luce")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
	if versions[0].SourceShort != "5.5e" {
		t.Errorf("expected source 5.5e, got %s", versions[0].SourceShort)
	}
	if versions[0].CompositeID != "5.5e/luce" {
		t.Errorf("expected composite ID 5.5e/luce, got %s", versions[0].CompositeID)
	}
}

func TestFindVersions_EmptyForUnknownSlug(t *testing.T) {
	store := datastore.NewStore(map[string][]map[string]any{
		"incantesimi": {
			{"_id": "luce", "_source_short": "5.5e", "_source": "srd-5.5e", "title": "Luce"},
		},
	})
	repo := NewDocumentRepository(store)

	versions, err := repo.FindVersions(context.Background(), "incantesimi", "non-esiste")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 0 {
		t.Errorf("expected 0 versions, got %d", len(versions))
	}
}
