package maps

import (
	"encoding/json"
	"io/fs"
	"log"
	"slices"
	"sort"
	"strings"

	domain "github.com/emiliopalmerini/quintaedizione.online/internal/domain/maps"
)

type jsonMappa struct {
	Slug          string   `json:"slug"`
	Nome          string   `json:"nome"`
	NomeOriginale string   `json:"nome_originale"`
	Immagine      string   `json:"immagine"`
	Categoria     string   `json:"categoria"`
	Tag           []string `json:"tag"`
	Descrizione   string   `json:"descrizione"`
	Autore        string   `json:"autore"`
	Licenza       string   `json:"licenza"`
	URLOriginale  string   `json:"url_originale"`
}

// MappaRepository provides in-memory access to map data.
type MappaRepository struct {
	mappe     []domain.Mappa
	bySlug    map[string]int
	categorie []string
	tags      []string
}

// NewMappaRepository loads maps from the provided filesystem.
func NewMappaRepository(dataFS fs.FS, filename string) *MappaRepository {
	data, err := fs.ReadFile(dataFS, filename)
	if err != nil {
		log.Fatalf("failed to read %s: %v", filename, err)
	}

	var raw []jsonMappa
	if err := json.Unmarshal(data, &raw); err != nil {
		log.Fatalf("failed to parse %s: %v", filename, err)
	}

	mappe := make([]domain.Mappa, 0, len(raw))
	bySlug := make(map[string]int)

	for _, m := range raw {
		if m.Immagine == "" {
			log.Printf("WARN: map %q has no immagine, skipping", m.Nome)
			continue
		}

		if m.Categoria == "" {
			m.Categoria = "senza_categoria"
		}

		if _, exists := bySlug[m.Slug]; exists {
			log.Printf("WARN: duplicate slug %q, skipping", m.Slug)
			continue
		}

		mappa := domain.Mappa{
			Slug:          m.Slug,
			Nome:          m.Nome,
			NomeOriginale: m.NomeOriginale,
			Immagine:      m.Immagine,
			Categoria:     m.Categoria,
			Tag:           m.Tag,
			Descrizione:   m.Descrizione,
			Autore:        m.Autore,
			Licenza:       m.Licenza,
			URLOriginale:  m.URLOriginale,
		}

		bySlug[m.Slug] = len(mappe)
		mappe = append(mappe, mappa)
	}

	sort.Slice(mappe, func(i, j int) bool {
		return mappe[i].Nome < mappe[j].Nome
	})

	// Rebuild index after sort
	for i, m := range mappe {
		bySlug[m.Slug] = i
	}

	repo := &MappaRepository{
		mappe:  mappe,
		bySlug: bySlug,
	}
	repo.buildFacets()
	return repo
}

func (r *MappaRepository) FindAll() []domain.Mappa {
	result := make([]domain.Mappa, len(r.mappe))
	copy(result, r.mappe)
	return result
}

func (r *MappaRepository) FindBySlug(slug string) (domain.Mappa, bool) {
	idx, ok := r.bySlug[slug]
	if !ok {
		return domain.Mappa{}, false
	}
	return r.mappe[idx], true
}

func (r *MappaRepository) Search(filters domain.SearchFilters) []domain.Mappa {
	query := strings.ToLower(filters.Query)
	var result []domain.Mappa

	for _, m := range r.mappe {
		if query != "" && !strings.Contains(strings.ToLower(m.Nome), query) {
			continue
		}
		if filters.Categoria != "" && m.Categoria != filters.Categoria {
			continue
		}
		if filters.Tag != "" && !slices.Contains(m.Tag, filters.Tag) {
			continue
		}
		result = append(result, m)
	}
	return result
}

func (r *MappaRepository) Categorie() []string { return r.categorie }
func (r *MappaRepository) Tags() []string      { return r.tags }

func (r *MappaRepository) buildFacets() {
	categoriaSet := make(map[string]bool)
	tagSet := make(map[string]bool)

	for _, m := range r.mappe {
		categoriaSet[m.Categoria] = true
		for _, t := range m.Tag {
			tagSet[t] = true
		}
	}

	r.categorie = sortedKeys(categoriaSet)
	r.tags = sortedKeys(tagSet)
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
