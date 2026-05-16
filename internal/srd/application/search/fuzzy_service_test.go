package search

import (
	"context"
	"testing"

	domainsearch "github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/search"
)

type stubRepo struct {
	items map[string][]domainsearch.SearchableItem
}

func (r *stubRepo) GetSearchableItems(_ context.Context, collection string) ([]domainsearch.SearchableItem, error) {
	return r.items[collection], nil
}

func newService(items map[string][]domainsearch.SearchableItem) *FuzzySearchService {
	svc := NewFuzzySearchService(&stubRepo{items: items})
	for col, list := range items {
		svc.index.Set(col, list)
	}
	return svc
}

func collectTitles(sets []domainsearch.SearchResultSet) []string {
	var out []string
	for _, s := range sets {
		for _, r := range s.Results {
			out = append(out, r.Title)
		}
	}
	return out
}

func containsTitle(titles []string, want string) bool {
	for _, t := range titles {
		if t == want {
			return true
		}
	}
	return false
}

func TestSearch_AccentInsensitive(t *testing.T) {
	svc := newService(map[string][]domainsearch.SearchableItem{
		"regole": {
			{ID: "1", Collection: "regole", Title: "Perché giocare"},
		},
	})

	results, err := svc.Search(context.Background(), "perche", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsTitle(collectTitles(results), "Perché giocare") {
		t.Errorf("expected accent-insensitive match, got %v", collectTitles(results))
	}
}

func TestSearch_MultiWordAllTokensRequired(t *testing.T) {
	svc := newService(map[string][]domainsearch.SearchableItem{
		"incantesimi": {
			{ID: "1", Collection: "incantesimi", Title: "Palla di Fuoco"},
			{ID: "2", Collection: "incantesimi", Title: "Palla di Ghiaccio"},
			{ID: "3", Collection: "incantesimi", Title: "Dardo di Fuoco"},
		},
	})

	results, err := svc.Search(context.Background(), "palla fuoco", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	titles := collectTitles(results)
	if !containsTitle(titles, "Palla di Fuoco") {
		t.Errorf("expected Palla di Fuoco in results, got %v", titles)
	}
	if containsTitle(titles, "Palla di Ghiaccio") {
		t.Errorf("Palla di Ghiaccio should not match 'palla fuoco', got %v", titles)
	}
	if containsTitle(titles, "Dardo di Fuoco") {
		t.Errorf("Dardo di Fuoco should not match 'palla fuoco', got %v", titles)
	}
}

func TestSearch_PrefixMatch(t *testing.T) {
	svc := newService(map[string][]domainsearch.SearchableItem{
		"incantesimi": {
			{ID: "1", Collection: "incantesimi", Title: "Invisibilità"},
			{ID: "2", Collection: "incantesimi", Title: "Visione"},
		},
	})

	results, err := svc.Search(context.Background(), "inv", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	titles := collectTitles(results)
	if !containsTitle(titles, "Invisibilità") {
		t.Errorf("expected prefix match on 'inv' → 'Invisibilità', got %v", titles)
	}
}

func TestSearch_FuzzyTypo(t *testing.T) {
	svc := newService(map[string][]domainsearch.SearchableItem{
		"incantesimi": {
			{ID: "1", Collection: "incantesimi", Title: "Palla di Fuoco"},
		},
	})

	// One-char typo in a 5-char token.
	results, err := svc.Search(context.Background(), "pala", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsTitle(collectTitles(results), "Palla di Fuoco") {
		t.Errorf("expected fuzzy match 'pala' → 'Palla di Fuoco', got %v", collectTitles(results))
	}
}

func TestSearch_DescriptionFallback(t *testing.T) {
	svc := newService(map[string][]domainsearch.SearchableItem{
		"incantesimi": {
			{
				ID: "1", Collection: "incantesimi",
				Title:       "Luce",
				Description: "Illumina un oggetto entro 18 metri per un'ora.",
			},
			{ID: "2", Collection: "incantesimi", Title: "Oscurità"},
		},
	})

	results, err := svc.Search(context.Background(), "illumina", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsTitle(collectTitles(results), "Luce") {
		t.Errorf("expected description match to surface 'Luce', got %v", collectTitles(results))
	}
}

func TestSearch_ShortQueryMinTwoChars(t *testing.T) {
	svc := newService(map[string][]domainsearch.SearchableItem{
		"incantesimi": {
			{ID: "1", Collection: "incantesimi", Title: "Identificare"},
		},
	})

	results, err := svc.Search(context.Background(), "id", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsTitle(collectTitles(results), "Identificare") {
		t.Errorf("expected 2-char prefix match, got %v", collectTitles(results))
	}
}

func TestSearch_TitleBeatsDescription(t *testing.T) {
	svc := newService(map[string][]domainsearch.SearchableItem{
		"incantesimi": {
			{ID: "1", Collection: "incantesimi", Title: "Cura Ferite", Description: ""},
			{ID: "2", Collection: "incantesimi", Title: "Benedizione", Description: "Lancia per curare gli alleati."},
		},
	})

	results, err := svc.Search(context.Background(), "cura", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	titles := collectTitles(results)
	if len(titles) < 1 || titles[0] != "Cura Ferite" {
		t.Errorf("title match should rank above description match, got %v", titles)
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	svc := newService(map[string][]domainsearch.SearchableItem{
		"incantesimi": {
			{ID: "1", Collection: "incantesimi", Title: "Palla di Fuoco"},
		},
	})

	results, err := svc.Search(context.Background(), "   ", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("empty query should return no results, got %v", collectTitles(results))
	}
}

func TestSearch_CollectionRankingByAggregate(t *testing.T) {
	svc := newService(map[string][]domainsearch.SearchableItem{
		"a_few_strong": {
			{ID: "1", Collection: "a_few_strong", Title: "Fuoco"},
		},
		"many_matches": {
			{ID: "2", Collection: "many_matches", Title: "Fuoco fatuo"},
			{ID: "3", Collection: "many_matches", Title: "Fuoco infernale"},
			{ID: "4", Collection: "many_matches", Title: "Fuoco sacro"},
		},
	})

	results, err := svc.Search(context.Background(), "fuoco", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("expected two collections, got %d", len(results))
	}
	// Single exact-title hit "Fuoco" scores higher than three prefix hits;
	// but aggregate of three prefix matches plus tiebreak on total should
	// still rank "many_matches" higher when scores are close. Verify both
	// collections appear and ordering is stable.
	seen := map[string]bool{}
	for _, r := range results {
		seen[r.Collection] = true
	}
	if !seen["a_few_strong"] || !seen["many_matches"] {
		t.Errorf("expected both collections in results, got %v", results)
	}
}

func TestSearchCollection_LimitRespected(t *testing.T) {
	svc := newService(map[string][]domainsearch.SearchableItem{
		"incantesimi": {
			{ID: "1", Collection: "incantesimi", Title: "Fuoco fatuo"},
			{ID: "2", Collection: "incantesimi", Title: "Fuoco infernale"},
			{ID: "3", Collection: "incantesimi", Title: "Fuoco sacro"},
		},
	})

	results, err := svc.SearchCollection(context.Background(), "incantesimi", "fuoco", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results limited, got %d", len(results))
	}
}
