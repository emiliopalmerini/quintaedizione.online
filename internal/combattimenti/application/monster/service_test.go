package monster

import (
	"strings"
	"testing"

	domain "github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/monster"
)

type mockRepo struct {
	monsters []domain.Monster
}

func (r *mockRepo) FindByMaxXP(maxXP int) []domain.Monster {
	var result []domain.Monster
	for _, m := range r.monsters {
		if m.XP <= maxXP {
			result = append(result, m)
		}
	}
	return result
}

func (r *mockRepo) Search(query string, maxXP int) []domain.Monster {
	var result []domain.Monster
	for _, m := range r.monsters {
		if m.XP <= maxXP {
			result = append(result, m)
		}
	}
	return result
}

func (r *mockRepo) SearchWithFilters(filters domain.SearchFilters) []domain.Monster {
	var result []domain.Monster
	for _, m := range r.monsters {
		if filters.MaxXP > 0 && m.XP > filters.MaxXP {
			continue
		}
		if filters.Query != "" && !strings.Contains(strings.ToLower(m.Name), strings.ToLower(filters.Query)) {
			continue
		}
		if filters.Type != "" && m.Type != filters.Type {
			continue
		}
		if filters.Size != "" && m.Size != filters.Size {
			continue
		}
		result = append(result, m)
	}
	return result
}

func (r *mockRepo) AvailableTypes() []string {
	return []string{"Bestia", "Drago", "Umanoide"}
}

func (r *mockRepo) AvailableSizes() []string {
	return []string{"Grande", "Media", "Piccola"}
}

func (r *mockRepo) AvailableCRs() []string {
	return []string{"0", "1/4", "1/2", "1", "17"}
}

func newTestService() *Service {
	repo := &mockRepo{
		monsters: []domain.Monster{
			{ID: "goblin", Name: "Goblin", Type: "Umanoide", Size: "Piccola", CR: "1/4", XP: 50},
			{ID: "orc", Name: "Orco", Type: "Umanoide", Size: "Media", CR: "1/2", XP: 100},
			{ID: "dragon", Name: "Drago Rosso Adulto", Type: "Drago", Size: "Grande", CR: "17", XP: 18000},
		},
	}
	return NewService(repo)
}

func TestSearchMonsters(t *testing.T) {
	svc := newTestService()

	results := svc.SearchMonsters("", 1_000_000)
	if len(results) != 3 {
		t.Errorf("expected 3 monsters, got %d", len(results))
	}

	results = svc.SearchMonsters("", 100)
	if len(results) != 2 {
		t.Errorf("expected 2 monsters with XP<=100, got %d", len(results))
	}
}

func TestSearchMonstersWithFilters(t *testing.T) {
	svc := newTestService()

	results := svc.SearchMonstersWithFilters(domain.SearchFilters{MaxXP: 1_000_000, Type: "Umanoide"})
	if len(results) != 2 {
		t.Errorf("expected 2 humanoids, got %d", len(results))
	}

	results = svc.SearchMonstersWithFilters(domain.SearchFilters{MaxXP: 1_000_000, Type: "Drago"})
	if len(results) != 1 {
		t.Errorf("expected 1 dragon, got %d", len(results))
	}
}

func TestAvailableTypes(t *testing.T) {
	svc := newTestService()
	types := svc.AvailableTypes()
	if len(types) != 3 {
		t.Errorf("expected 3 types, got %d", len(types))
	}
}

func TestAvailableSizes(t *testing.T) {
	svc := newTestService()
	sizes := svc.AvailableSizes()
	if len(sizes) != 3 {
		t.Errorf("expected 3 sizes, got %d", len(sizes))
	}
}

func TestAvailableCRs(t *testing.T) {
	svc := newTestService()
	crs := svc.AvailableCRs()
	if len(crs) != 5 {
		t.Errorf("expected 5 CRs, got %d", len(crs))
	}
}
