package persistence

import (
	"testing"
	"testing/fstest"

	"github.com/emiliopalmerini/quintaedizione.online/internal/mappe/domain"
)

const testJSON = `[
  {
    "slug": "il-tempio-sommerso",
    "nome": "Il Tempio Sommerso",
    "nome_originale": "The Sunken Temple",
    "immagine": "il-tempio-sommerso.png",
    "tag": ["dungeon", "acqua", "trappola"],
    "autore": "Dyson Logos",
    "licenza": "Commercial Use Allowed",
    "url_originale": "https://example.com/sunken-temple"
  },
  {
    "slug": "la-caverna-dei-funghi",
    "nome": "La Caverna dei Funghi",
    "nome_originale": "The Mushroom Cavern",
    "immagine": "la-caverna-dei-funghi.png",
    "tag": ["caverna", "funghi", "sotterraneo"],
    "autore": "Dyson Logos",
    "licenza": "Commercial Use Allowed",
    "url_originale": "https://example.com/mushroom-cavern"
  },
  {
    "slug": "la-piazza-del-mercato",
    "nome": "La Piazza del Mercato",
    "nome_originale": "The Market Square",
    "immagine": "la-piazza-del-mercato.png",
    "tag": ["citta", "mercato", "urbano"],
    "autore": "Dyson Logos",
    "licenza": "Commercial Use Allowed",
    "url_originale": "https://example.com/market-square"
  }
]`

func newTestRepo() *MappaRepository {
	fs := fstest.MapFS{
		"mappe.json": &fstest.MapFile{Data: []byte(testJSON)},
	}
	repo, err := NewMappaRepository(fs, "mappe.json")
	if err != nil {
		panic(err)
	}
	return repo
}

func TestNewMappaRepository_LoadsMaps(t *testing.T) {
	repo := newTestRepo()
	all := repo.FindAll()
	if len(all) != 3 {
		t.Fatalf("expected 3 maps, got %d", len(all))
	}
}

func TestNewMappaRepository_SortsByName(t *testing.T) {
	repo := newTestRepo()
	all := repo.FindAll()
	for i := 1; i < len(all); i++ {
		if all[i].Nome < all[i-1].Nome {
			t.Errorf("maps not sorted: %q before %q", all[i-1].Nome, all[i].Nome)
		}
	}
}

func TestFindBySlug_Found(t *testing.T) {
	repo := newTestRepo()
	m, ok := repo.FindBySlug("il-tempio-sommerso")
	if !ok {
		t.Fatal("expected to find map by slug")
	}
	if m.Nome != "Il Tempio Sommerso" {
		t.Errorf("expected 'Il Tempio Sommerso', got %q", m.Nome)
	}
	if m.NomeOriginale != "The Sunken Temple" {
		t.Errorf("expected 'The Sunken Temple', got %q", m.NomeOriginale)
	}
}

func TestFindBySlug_NotFound(t *testing.T) {
	repo := newTestRepo()
	_, ok := repo.FindBySlug("non-esiste")
	if ok {
		t.Error("expected not found for unknown slug")
	}
}

func TestSearch_ByQuery(t *testing.T) {
	repo := newTestRepo()
	results, total := repo.Search(domain.SearchFilters{Query: "tempio"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if results[0].Slug != "il-tempio-sommerso" {
		t.Errorf("expected il-tempio-sommerso, got %s", results[0].Slug)
	}
}

func TestSearch_ByQueryCaseInsensitive(t *testing.T) {
	repo := newTestRepo()
	results, _ := repo.Search(domain.SearchFilters{Query: "CAVERNA"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestSearch_BySingleTag(t *testing.T) {
	repo := newTestRepo()
	results, _ := repo.Search(domain.SearchFilters{Tags: []string{"acqua"}})
	if len(results) != 1 {
		t.Fatalf("expected 1 result with tag acqua, got %d", len(results))
	}
	if results[0].Slug != "il-tempio-sommerso" {
		t.Errorf("expected il-tempio-sommerso, got %s", results[0].Slug)
	}
}

func TestSearch_ByMultipleTags(t *testing.T) {
	repo := newTestRepo()
	// dungeon + acqua = only il-tempio-sommerso
	results, total := repo.Search(domain.SearchFilters{Tags: []string{"dungeon", "acqua"}})
	if len(results) != 1 {
		t.Fatalf("expected 1 result with tags dungeon+acqua, got %d", len(results))
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if results[0].Slug != "il-tempio-sommerso" {
		t.Errorf("expected il-tempio-sommerso, got %s", results[0].Slug)
	}
}

func TestSearch_ByMultipleTags_NoMatch(t *testing.T) {
	repo := newTestRepo()
	// dungeon + funghi = no map has both
	results, total := repo.Search(domain.SearchFilters{Tags: []string{"dungeon", "funghi"}})
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
	if total != 0 {
		t.Errorf("expected total 0, got %d", total)
	}
}

func TestSearch_Combined(t *testing.T) {
	repo := newTestRepo()
	results, _ := repo.Search(domain.SearchFilters{
		Query: "piazza",
		Tags:  []string{"citta", "urbano"},
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Slug != "la-piazza-del-mercato" {
		t.Errorf("expected la-piazza-del-mercato, got %s", results[0].Slug)
	}
}

func TestSearch_NoResults(t *testing.T) {
	repo := newTestRepo()
	results, total := repo.Search(domain.SearchFilters{Query: "non-esiste"})
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
	if total != 0 {
		t.Errorf("expected total 0, got %d", total)
	}
}

func TestSearch_EmptyFilters_ReturnsAll(t *testing.T) {
	repo := newTestRepo()
	results, total := repo.Search(domain.SearchFilters{})
	if len(results) != 3 {
		t.Errorf("expected 3 results with empty filters, got %d", len(results))
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
}

func TestSearch_ZeroLimitReturnsAll(t *testing.T) {
	repo := newTestRepo()
	results, total := repo.Search(domain.SearchFilters{Limit: 0})
	if len(results) != 3 {
		t.Errorf("expected 3 results with zero limit, got %d", len(results))
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
}

func TestSearch_WithLimit(t *testing.T) {
	repo := newTestRepo()
	results, total := repo.Search(domain.SearchFilters{Limit: 2})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
}

func TestSearch_WithOffset(t *testing.T) {
	repo := newTestRepo()
	results, total := repo.Search(domain.SearchFilters{Offset: 1})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
}

func TestSearch_WithOffsetAndLimit(t *testing.T) {
	repo := newTestRepo()
	results, total := repo.Search(domain.SearchFilters{Offset: 1, Limit: 1})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
	// Sorted by name: Il Tempio, La Caverna, La Piazza — offset 1 should be La Caverna
	if results[0].Slug != "la-caverna-dei-funghi" {
		t.Errorf("expected la-caverna-dei-funghi at offset 1, got %s", results[0].Slug)
	}
}

func TestSearch_OffsetBeyondTotal(t *testing.T) {
	repo := newTestRepo()
	results, total := repo.Search(domain.SearchFilters{Offset: 10})
	if len(results) != 0 {
		t.Errorf("expected 0 results for offset beyond total, got %d", len(results))
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
}

func TestSearch_TagFilterAndPagination(t *testing.T) {
	repo := newTestRepo()
	results, total := repo.Search(domain.SearchFilters{
		Tags:  []string{"dungeon"},
		Limit: 10,
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 dungeon, got %d", len(results))
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
}

func TestTags(t *testing.T) {
	repo := newTestRepo()
	tags := repo.Tags()
	// 9 unique tags: acqua, caverna, citta, dungeon, funghi, mercato, sotterraneo, trappola, urbano
	if len(tags) != 9 {
		t.Fatalf("expected 9 tags, got %d: %v", len(tags), tags)
	}
	// Should be sorted
	for i := 1; i < len(tags); i++ {
		if tags[i] < tags[i-1] {
			t.Errorf("tags not sorted: %q before %q", tags[i-1], tags[i])
		}
	}
}

func TestNewMappaRepository_SkipsMissingImmagine(t *testing.T) {
	json := `[
		{"slug": "no-image", "nome": "No Image", "tag": []},
		{"slug": "has-image", "nome": "Has Image", "immagine": "img.png", "tag": []}
	]`
	fs := fstest.MapFS{
		"mappe.json": &fstest.MapFile{Data: []byte(json)},
	}
	repo, err := NewMappaRepository(fs, "mappe.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	all := repo.FindAll()
	if len(all) != 1 {
		t.Fatalf("expected 1 map (skipping missing immagine), got %d", len(all))
	}
	if all[0].Slug != "has-image" {
		t.Errorf("expected has-image, got %s", all[0].Slug)
	}
}

func TestNewMappaRepository_SkipsDuplicateSlug(t *testing.T) {
	json := `[
		{"slug": "dup", "nome": "First", "immagine": "a.png", "tag": []},
		{"slug": "dup", "nome": "Second", "immagine": "b.png", "tag": []}
	]`
	fs := fstest.MapFS{
		"mappe.json": &fstest.MapFile{Data: []byte(json)},
	}
	repo, err := NewMappaRepository(fs, "mappe.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	all := repo.FindAll()
	if len(all) != 1 {
		t.Fatalf("expected 1 map (skipping duplicate), got %d", len(all))
	}
	if all[0].Nome != "First" {
		t.Errorf("expected First (keep first), got %s", all[0].Nome)
	}
}

func TestNewMappaRepository_ReturnsLoadError(t *testing.T) {
	_, err := NewMappaRepository(fstest.MapFS{}, "missing.json")
	if err == nil {
		t.Fatal("expected load error, got nil")
	}
}
