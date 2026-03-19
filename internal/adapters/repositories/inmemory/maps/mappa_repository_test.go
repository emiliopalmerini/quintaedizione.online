package maps

import (
	"testing"
	"testing/fstest"

	domain "github.com/emiliopalmerini/quintaedizione.online/internal/domain/maps"
)

const testJSON = `[
  {
    "slug": "il-tempio-sommerso",
    "nome": "Il Tempio Sommerso",
    "nome_originale": "The Sunken Temple",
    "immagine": "il-tempio-sommerso.png",
    "categoria": "dungeon",
    "tag": ["acqua", "trappola"],
    "autore": "Dyson Logos",
    "licenza": "Commercial Use Allowed",
    "url_originale": "https://example.com/sunken-temple"
  },
  {
    "slug": "la-caverna-dei-funghi",
    "nome": "La Caverna dei Funghi",
    "nome_originale": "The Mushroom Cavern",
    "immagine": "la-caverna-dei-funghi.png",
    "categoria": "caverna",
    "tag": ["funghi", "sotterraneo"],
    "autore": "Dyson Logos",
    "licenza": "Commercial Use Allowed",
    "url_originale": "https://example.com/mushroom-cavern"
  },
  {
    "slug": "la-piazza-del-mercato",
    "nome": "La Piazza del Mercato",
    "nome_originale": "The Market Square",
    "immagine": "la-piazza-del-mercato.png",
    "categoria": "citta",
    "tag": ["mercato", "urbano"],
    "autore": "Dyson Logos",
    "licenza": "Commercial Use Allowed",
    "url_originale": "https://example.com/market-square"
  }
]`

func newTestRepo() *MappaRepository {
	fs := fstest.MapFS{
		"mappe.json": &fstest.MapFile{Data: []byte(testJSON)},
	}
	return NewMappaRepository(fs, "mappe.json")
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
	results := repo.Search(domain.SearchFilters{Query: "tempio"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Slug != "il-tempio-sommerso" {
		t.Errorf("expected il-tempio-sommerso, got %s", results[0].Slug)
	}
}

func TestSearch_ByQueryCaseInsensitive(t *testing.T) {
	repo := newTestRepo()
	results := repo.Search(domain.SearchFilters{Query: "CAVERNA"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestSearch_ByCategoria(t *testing.T) {
	repo := newTestRepo()
	results := repo.Search(domain.SearchFilters{Categoria: "dungeon"})
	if len(results) != 1 {
		t.Fatalf("expected 1 dungeon, got %d", len(results))
	}
	if results[0].Categoria != "dungeon" {
		t.Errorf("expected categoria dungeon, got %s", results[0].Categoria)
	}
}

func TestSearch_ByTag(t *testing.T) {
	repo := newTestRepo()
	results := repo.Search(domain.SearchFilters{Tag: "acqua"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result with tag acqua, got %d", len(results))
	}
	if results[0].Slug != "il-tempio-sommerso" {
		t.Errorf("expected il-tempio-sommerso, got %s", results[0].Slug)
	}
}

func TestSearch_Combined(t *testing.T) {
	repo := newTestRepo()
	results := repo.Search(domain.SearchFilters{
		Query:     "piazza",
		Categoria: "citta",
		Tag:       "urbano",
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
	results := repo.Search(domain.SearchFilters{Query: "non-esiste"})
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_EmptyFilters_ReturnsAll(t *testing.T) {
	repo := newTestRepo()
	results := repo.Search(domain.SearchFilters{})
	if len(results) != 3 {
		t.Errorf("expected 3 results with empty filters, got %d", len(results))
	}
}

func TestCategorie(t *testing.T) {
	repo := newTestRepo()
	cats := repo.Categorie()
	if len(cats) != 3 {
		t.Fatalf("expected 3 categories, got %d", len(cats))
	}
	// Should be sorted
	expected := []string{"caverna", "citta", "dungeon"}
	for i, c := range cats {
		if c != expected[i] {
			t.Errorf("expected category %q at index %d, got %q", expected[i], i, c)
		}
	}
}

func TestTags(t *testing.T) {
	repo := newTestRepo()
	tags := repo.Tags()
	if len(tags) != 6 {
		t.Fatalf("expected 6 tags, got %d: %v", len(tags), tags)
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
		{"slug": "no-image", "nome": "No Image", "categoria": "dungeon", "tag": []},
		{"slug": "has-image", "nome": "Has Image", "immagine": "img.png", "categoria": "dungeon", "tag": []}
	]`
	fs := fstest.MapFS{
		"mappe.json": &fstest.MapFile{Data: []byte(json)},
	}
	repo := NewMappaRepository(fs, "mappe.json")
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
		{"slug": "dup", "nome": "First", "immagine": "a.png", "categoria": "dungeon", "tag": []},
		{"slug": "dup", "nome": "Second", "immagine": "b.png", "categoria": "dungeon", "tag": []}
	]`
	fs := fstest.MapFS{
		"mappe.json": &fstest.MapFile{Data: []byte(json)},
	}
	repo := NewMappaRepository(fs, "mappe.json")
	all := repo.FindAll()
	if len(all) != 1 {
		t.Fatalf("expected 1 map (skipping duplicate), got %d", len(all))
	}
	if all[0].Nome != "First" {
		t.Errorf("expected First (keep first), got %s", all[0].Nome)
	}
}

func TestNewMappaRepository_DefaultsCategoria(t *testing.T) {
	json := `[{"slug": "test", "nome": "Test", "immagine": "t.png", "tag": []}]`
	fs := fstest.MapFS{
		"mappe.json": &fstest.MapFile{Data: []byte(json)},
	}
	repo := NewMappaRepository(fs, "mappe.json")
	m, ok := repo.FindBySlug("test")
	if !ok {
		t.Fatal("expected to find map")
	}
	if m.Categoria != "senza_categoria" {
		t.Errorf("expected default categoria 'senza_categoria', got %q", m.Categoria)
	}
}
