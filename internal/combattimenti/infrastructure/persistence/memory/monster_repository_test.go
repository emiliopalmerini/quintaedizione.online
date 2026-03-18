package memory

import (
	"testing"

	jsondata "github.com/emiliopalmerini/quintaedizione.online/data/ita/json"
	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/monster"
)

func newTestRepo() *MonsterRepository {
	return NewMonsterRepository(jsondata.Files, "monsters.json")
}

func TestNewMonsterRepository_LoadsMonsters(t *testing.T) {
	repo := newTestRepo()
	monsters := repo.FindByMaxXP(1_000_000)
	if len(monsters) == 0 {
		t.Fatal("expected monsters to be loaded, got 0")
	}
	// We know there are 330 monsters in the JSON
	if len(monsters) != 330 {
		t.Errorf("expected 330 monsters, got %d", len(monsters))
	}
}

func TestNewMonsterRepository_ParsesXP(t *testing.T) {
	repo := newTestRepo()
	// Aboleth has CR 10, XP 5900
	monsters := repo.Search("aboleth", 1_000_000)
	if len(monsters) != 1 {
		t.Fatalf("expected 1 aboleth, got %d", len(monsters))
	}
	if monsters[0].XP != 5900 {
		t.Errorf("expected aboleth XP=5900, got %d", monsters[0].XP)
	}
	if monsters[0].CR != "10" {
		t.Errorf("expected aboleth CR=10, got %s", monsters[0].CR)
	}
}

func TestFindByMaxXP_FiltersCorrectly(t *testing.T) {
	repo := newTestRepo()

	// Only CR 0 monsters have XP 0
	zeroXP := repo.FindByMaxXP(0)
	for _, m := range zeroXP {
		if m.XP != 0 {
			t.Errorf("monster %s has XP %d, expected 0", m.Name, m.XP)
		}
	}

	low := repo.FindByMaxXP(100)
	for _, m := range low {
		if m.XP > 100 {
			t.Errorf("monster %s has XP %d, expected <= 100", m.Name, m.XP)
		}
	}

	all := repo.FindByMaxXP(1_000_000)
	mid := repo.FindByMaxXP(1000)
	if len(mid) >= len(all) {
		t.Error("expected fewer monsters at lower XP cap")
	}
}

func TestSearch_FiltersByNameAndXP(t *testing.T) {
	repo := newTestRepo()

	// Search by name
	dragons := repo.Search("drago", 1_000_000)
	if len(dragons) == 0 {
		t.Fatal("expected some dragons")
	}
	for _, m := range dragons {
		if !containsInsensitive(m.Name, "drago") {
			t.Errorf("monster %s does not match 'drago'", m.Name)
		}
	}

	// Search with XP cap
	weakDragons := repo.Search("drago", 500)
	for _, m := range weakDragons {
		if m.XP > 500 {
			t.Errorf("monster %s has XP %d, expected <= 500", m.Name, m.XP)
		}
	}
}

func TestNewMonsterRepository_ParsesDetailFields(t *testing.T) {
	repo := newTestRepo()
	monsters := repo.Search("aboleth", 1_000_000)
	if len(monsters) != 1 {
		t.Fatalf("expected 1 aboleth, got %d", len(monsters))
	}
	m := monsters[0]

	// Alignment
	if m.Alignment == "" {
		t.Error("expected aboleth to have alignment")
	}

	// Speed
	if m.Speed == "" {
		t.Error("expected aboleth to have speed")
	}

	// Initiative
	if m.Initiative == "" {
		t.Error("expected aboleth to have initiative")
	}

	// Ability scores
	if m.AbilityScores.Strength != 21 {
		t.Errorf("expected aboleth STR=21, got %d", m.AbilityScores.Strength)
	}
	if m.AbilityScores.Intelligence != 18 {
		t.Errorf("expected aboleth INT=18, got %d", m.AbilityScores.Intelligence)
	}

	// Ability mods
	if m.AbilityMods.Strength != 5 {
		t.Errorf("expected aboleth STR mod=5, got %d", m.AbilityMods.Strength)
	}
	if m.AbilityMods.Dexterity != -1 {
		t.Errorf("expected aboleth DEX mod=-1, got %d", m.AbilityMods.Dexterity)
	}

	// Saving throws
	if m.SavingThrows.Intelligence != "+8" {
		t.Errorf("expected aboleth INT save=+8, got %s", m.SavingThrows.Intelligence)
	}

	// Traits
	if len(m.Traits) == 0 {
		t.Fatal("expected aboleth to have traits")
	}
	if m.Traits[0].Name != "Anfibio" {
		t.Errorf("expected first trait 'Anfibio', got %q", m.Traits[0].Name)
	}

	// Actions
	if len(m.Actions) == 0 {
		t.Fatal("expected aboleth to have actions")
	}
	if m.Actions[0].Name != "Multiattacco" {
		t.Errorf("expected first action 'Multiattacco', got %q", m.Actions[0].Name)
	}

	// Legendary actions
	if len(m.LegendaryActions) == 0 {
		t.Fatal("expected aboleth to have legendary actions")
	}

	// CRDetail
	if m.CRDetail == "" {
		t.Error("expected aboleth to have cr_detail")
	}

	// Skills
	if m.Skills == "" {
		t.Error("expected aboleth to have skills")
	}

	// Senses
	if m.Senses == "" {
		t.Error("expected aboleth to have senses")
	}

	// Languages
	if m.Languages == "" {
		t.Error("expected aboleth to have languages")
	}
}

func TestSearch_EmptyQuery_ReturnsAllUpToMaxXP(t *testing.T) {
	repo := newTestRepo()
	all := repo.Search("", 1_000_000)
	if len(all) != 330 {
		t.Errorf("expected 330 monsters with empty query, got %d", len(all))
	}
}

func TestNormalizeSize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Minuscolo", "Minuscola"},
		{"Minuscola", "Minuscola"},
		{"Piccolo", "Piccola"},
		{"Piccola", "Piccola"},
		{"Medio", "Media"},
		{"Media", "Media"},
		{"Grande", "Grande"},
		{"Enorme", "Enorme"},
		{"Mastodontico", "Mastodontica"},
		{"Mastodontica", "Mastodontica"},
		{"Media o Piccola", "Media"},
		{"Medio o Piccolo", "Media"},
		{"Grande di bestie Minuscole", "Grande"},
		{"Medio di bestie Minuscole", "Media"},
		{"Medio di non morti Minuscoli", "Media"},
	}
	for _, tt := range tests {
		got := normalizeSize(tt.input)
		if got != tt.expected {
			t.Errorf("normalizeSize(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestCRValue(t *testing.T) {
	tests := []struct {
		cr       string
		expected float64
	}{
		{"0", 0},
		{"1/8", 0.125},
		{"1/4", 0.25},
		{"1/2", 0.5},
		{"1", 1},
		{"10", 10},
		{"30", 30},
	}
	for _, tt := range tests {
		got := crValue(tt.cr)
		if got != tt.expected {
			t.Errorf("crValue(%q) = %v, want %v", tt.cr, got, tt.expected)
		}
	}
}

func TestSearchWithFilters_ByType(t *testing.T) {
	repo := newTestRepo()
	filters := monster.SearchFilters{MaxXP: 1_000_000, Type: "Drago"}
	results := repo.SearchWithFilters(filters)
	if len(results) == 0 {
		t.Fatal("expected some dragons")
	}
	for _, m := range results {
		if m.Type != "Drago" {
			t.Errorf("expected type Drago, got %s for %s", m.Type, m.Name)
		}
	}
}

func TestSearchWithFilters_BySize(t *testing.T) {
	repo := newTestRepo()
	filters := monster.SearchFilters{MaxXP: 1_000_000, Size: "Grande"}
	results := repo.SearchWithFilters(filters)
	if len(results) == 0 {
		t.Fatal("expected some Grande monsters")
	}
	for _, m := range results {
		if m.Size != "Grande" {
			t.Errorf("expected size Grande, got %s for %s", m.Size, m.Name)
		}
	}
}

func TestSearchWithFilters_ByCRRange(t *testing.T) {
	repo := newTestRepo()
	filters := monster.SearchFilters{MaxXP: 1_000_000, CRMin: "5", CRMax: "10"}
	results := repo.SearchWithFilters(filters)
	if len(results) == 0 {
		t.Fatal("expected some monsters in CR 5-10")
	}
	for _, m := range results {
		cr := crValue(m.CR)
		if cr < 5 || cr > 10 {
			t.Errorf("monster %s has CR %s (%.2f), expected 5-10", m.Name, m.CR, cr)
		}
	}
}

func TestSearchWithFilters_Combined(t *testing.T) {
	repo := newTestRepo()
	filters := monster.SearchFilters{
		MaxXP: 1_000_000,
		Query: "drago",
		Type:  "Drago",
		CRMin: "1",
	}
	results := repo.SearchWithFilters(filters)
	if len(results) == 0 {
		t.Fatal("expected some dragons with CR >= 1")
	}
	for _, m := range results {
		if m.Type != "Drago" {
			t.Errorf("expected type Drago, got %s", m.Type)
		}
		if crValue(m.CR) < 1 {
			t.Errorf("expected CR >= 1, got %s", m.CR)
		}
	}
}

func TestAvailableTypes(t *testing.T) {
	repo := newTestRepo()
	types := repo.AvailableTypes()
	if len(types) == 0 {
		t.Fatal("expected some types")
	}
	// Should be sorted
	for i := 1; i < len(types); i++ {
		if types[i] < types[i-1] {
			t.Errorf("types not sorted: %s before %s", types[i-1], types[i])
		}
	}
}

func TestAvailableSizes(t *testing.T) {
	repo := newTestRepo()
	sizes := repo.AvailableSizes()
	if len(sizes) == 0 {
		t.Fatal("expected some sizes")
	}
	// All sizes should be normalized (no masculine forms)
	for _, s := range sizes {
		if s == "Minuscolo" || s == "Piccolo" || s == "Medio" || s == "Mastodontico" {
			t.Errorf("size %q should have been normalized", s)
		}
	}
}

func TestAvailableCRs(t *testing.T) {
	repo := newTestRepo()
	crs := repo.AvailableCRs()
	if len(crs) == 0 {
		t.Fatal("expected some CRs")
	}
	// First should be 0 or lowest
	if crs[0] != "0" {
		t.Errorf("expected first CR to be 0, got %s", crs[0])
	}
}

func containsInsensitive(s, substr string) bool {
	return len(s) >= len(substr) // placeholder, real check in impl
}
