package persistence

import (
	"fmt"
	"io/fs"
	"log"
	"slices"
	"sort"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/mappe/domain"
	"github.com/emiliopalmerini/quintaedizione.online/pkg/datastore"
)

type jsonMappa struct {
	Slug                 string   `json:"slug"`
	Nome                 string   `json:"nome"`
	NomeOriginale        string   `json:"nome_originale"`
	Immagine             string   `json:"immagine"`
	Tag                  []string `json:"tag"`
	Descrizione          string   `json:"descrizione"`
	Autore               string   `json:"autore"`
	Licenza              string   `json:"licenza"`
	URLOriginale         string   `json:"url_originale"`
	URLImmagineOriginale string   `json:"url_immagine_originale"`
}

// MappaRepository provides in-memory access to map data.
type MappaRepository struct {
	mappe  []domain.Mappa
	bySlug map[string]int
	tags   []string
}

// NewMappaRepository loads maps from the provided filesystem.
func NewMappaRepository(dataFS fs.FS, filename string) (*MappaRepository, error) {
	raw, err := datastore.LoadJSON[jsonMappa](dataFS, filename)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", filename, err)
	}

	mappe := make([]domain.Mappa, 0, len(raw))
	bySlug := make(map[string]int)

	for _, m := range raw {
		if m.Immagine == "" {
			log.Printf("WARN: map %q has no immagine, skipping", m.Nome)
			continue
		}

		if _, exists := bySlug[m.Slug]; exists {
			log.Printf("WARN: duplicate slug %q, skipping", m.Slug)
			continue
		}

		mappa := domain.Mappa{
			Slug:                 m.Slug,
			Nome:                 m.Nome,
			NomeOriginale:        m.NomeOriginale,
			Immagine:             m.Immagine,
			Tag:                  m.Tag,
			Descrizione:          m.Descrizione,
			Autore:               m.Autore,
			Licenza:              m.Licenza,
			URLOriginale:         m.URLOriginale,
			URLImmagineOriginale: m.URLImmagineOriginale,
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
	return repo, nil
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

func (r *MappaRepository) Search(filters domain.SearchFilters) ([]domain.Mappa, int) {
	query := strings.ToLower(filters.Query)
	var filtered []domain.Mappa

	for _, m := range r.mappe {
		if query != "" && !strings.Contains(strings.ToLower(m.Nome), query) {
			continue
		}
		if !hasAllTags(m.Tag, filters.Tags) {
			continue
		}
		filtered = append(filtered, m)
	}

	total := len(filtered)

	if filters.Offset >= total {
		return nil, total
	}

	start := filters.Offset
	end := total
	if filters.Limit > 0 && start+filters.Limit < end {
		end = start + filters.Limit
	}

	return filtered[start:end], total
}

func (r *MappaRepository) Tags() []string { return r.tags }

func (r *MappaRepository) buildFacets() {
	tagSet := make(map[string]bool)
	for _, m := range r.mappe {
		for _, t := range m.Tag {
			tagSet[t] = true
		}
	}
	r.tags = sortedKeys(tagSet)
}

func hasAllTags(mappaTags []string, filterTags []string) bool {
	for _, ft := range filterTags {
		if !slices.Contains(mappaTags, ft) {
			return false
		}
	}
	return true
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
