package srd

import (
	"context"
	"errors"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/monster"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/repositories"
)

type fakeDocs struct {
	all []map[string]any
}

func (f *fakeDocs) FindByID(_ context.Context, collection string, compositeID string) (*domain.Document, error) {
	for _, d := range f.all {
		id, _ := d["_id"].(string)
		short, _ := d["_source_short"].(string)
		if collection == "mostri" && short+"/"+id == compositeID {
			return domain.DocumentFromMap(d), nil
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeDocs) FindByPredicate(_ context.Context, collection string, match repositories.DocumentPredicate, skip, limit int64) ([]*domain.Document, int64, error) {
	if collection != "mostri" {
		return nil, 0, nil
	}
	var out []*domain.Document
	for _, d := range f.all {
		if match == nil || match(d) {
			out = append(out, domain.DocumentFromMap(d))
		}
	}
	total := int64(len(out))
	if skip > 0 {
		if int(skip) >= len(out) {
			return nil, total, nil
		}
		out = out[skip:]
	}
	if limit > 0 && int64(len(out)) > limit {
		out = out[:limit]
	}
	return out, total, nil
}

func monsterDoc(source, id, title, cr string) map[string]any {
	return map[string]any{
		"_id":           id,
		"_source_short": source,
		"title":         title,
		"cr":            cr,
	}
}

// monsterDocFull builds a doc with bare CR (grado_sfida) plus type, matching
// the loader's real output. Tests for CR-range / type filtering use this.
func monsterDocFull(source, id, title, gradoSfida, cr, tipo string) map[string]any {
	return map[string]any{
		"_id":           id,
		"_source_short": source,
		"title":         title,
		"grado_sfida":   gradoSfida,
		"cr":            cr,
		"tipo":          tipo,
	}
}

func TestMonsterReader_ParsesItalianThousandsSeparator(t *testing.T) {
	docs := &fakeDocs{all: []map[string]any{
		monsterDoc("5.5e", "goblin", "Goblin", "1/4 (PE 50; BC +2)"),
		monsterDoc("5.5e", "lich", "Lich", "21 (PE 33.000; BC +7)"),
	}}
	r := NewMonsterReader(docs)

	got, err := r.Search(context.Background(), monster.SearchQuery{Source: "5.5e"})
	if err != nil {
		t.Fatalf("Search err: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Name != "Goblin" || got[0].XP != 50 {
		t.Errorf("entry[0] = %+v, want Goblin/50", got[0])
	}
	if got[1].Name != "Lich" || got[1].XP != 33000 {
		t.Errorf("entry[1] = %+v, want Lich/33000", got[1])
	}
}

func TestMonsterReader_FiltersBySource(t *testing.T) {
	docs := &fakeDocs{all: []map[string]any{
		monsterDoc("5.5e", "goblin", "Goblin", "1/4 (PE 50; BC +2)"),
		monsterDoc("5e", "goblin", "Goblin (vecchio)", "1/4 (PE 50; BC +2)"),
	}}
	r := NewMonsterReader(docs)

	got, _ := r.Search(context.Background(), monster.SearchQuery{Source: "5.5e"})
	if len(got) != 1 || got[0].Source != "5.5e" {
		t.Fatalf("got %+v, want single 5.5e entry", got)
	}
}

func TestMonsterReader_SubstringSearchIsCaseInsensitive(t *testing.T) {
	docs := &fakeDocs{all: []map[string]any{
		monsterDoc("5.5e", "goblin", "Goblin", "1/4 (PE 50; BC +2)"),
		monsterDoc("5.5e", "hobgoblin", "Hobgoblin", "1/2 (PE 100; BC +2)"),
		monsterDoc("5.5e", "lich", "Lich", "21 (PE 33.000; BC +7)"),
	}}
	r := NewMonsterReader(docs)

	got, _ := r.Search(context.Background(), monster.SearchQuery{
		Source: "5.5e",
		Query:  "GOBL",
	})
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
}

func TestMonsterReader_SortsByXPThenName(t *testing.T) {
	docs := &fakeDocs{all: []map[string]any{
		monsterDoc("5.5e", "beholder", "Beholder", "13 (PE 10.000; BC +6)"),
		monsterDoc("5.5e", "goblin", "Goblin", "1/4 (PE 50; BC +2)"),
		monsterDoc("5.5e", "kobold", "Kobold", "1/4 (PE 50; BC +2)"),
	}}
	r := NewMonsterReader(docs)

	got, _ := r.Search(context.Background(), monster.SearchQuery{Source: "5.5e"})
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[0].Name != "Goblin" || got[1].Name != "Kobold" || got[2].Name != "Beholder" {
		t.Errorf("order = %q,%q,%q, want Goblin,Kobold,Beholder", got[0].Name, got[1].Name, got[2].Name)
	}
}

func TestMonsterReader_FindByIDComposesKey(t *testing.T) {
	docs := &fakeDocs{all: []map[string]any{
		monsterDoc("5.5e", "goblin", "Goblin", "1/4 (PE 50; BC +2)"),
	}}
	r := NewMonsterReader(docs)

	m, err := r.FindByID(context.Background(), "5.5e", "goblin")
	if err != nil {
		t.Fatalf("FindByID err: %v", err)
	}
	if m.Name != "Goblin" || m.XP != 50 {
		t.Errorf("got %+v, want Goblin/50", m)
	}
}

func TestMonsterReader_EmptySourceReturnsNil(t *testing.T) {
	docs := &fakeDocs{all: []map[string]any{
		monsterDoc("5.5e", "goblin", "Goblin", "1/4 (PE 50; BC +2)"),
	}}
	r := NewMonsterReader(docs)

	got, err := r.Search(context.Background(), monster.SearchQuery{Source: ""})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != nil {
		t.Fatalf("got %+v, want nil", got)
	}
}

func TestMonsterReader_MissingXPReturnsZero(t *testing.T) {
	docs := &fakeDocs{all: []map[string]any{
		monsterDoc("5.5e", "mystery", "Mystery", "???"),
	}}
	r := NewMonsterReader(docs)

	got, _ := r.Search(context.Background(), monster.SearchQuery{Source: "5.5e"})
	if len(got) != 1 || got[0].XP != 0 {
		t.Fatalf("got %+v, want one entry with XP=0", got)
	}
}

func TestMonsterReader_Parses5eFormat(t *testing.T) {
	// 5e dataset renders CR as "1/4 (50 PE)" (digits before PE), whereas 5.5e
	// uses "10 (PE 5.900; ...)". Both must parse.
	docs := &fakeDocs{all: []map[string]any{
		monsterDoc("5e", "goblin", "Goblin", "1/4 (50 PE)"),
		monsterDoc("5e", "dragon", "Drago", "20 (25.000 PE)"),
	}}
	r := NewMonsterReader(docs)

	got, _ := r.Search(context.Background(), monster.SearchQuery{Source: "5e"})
	if len(got) != 2 {
		t.Fatalf("len=%d, want 2", len(got))
	}
	if got[0].XP != 50 {
		t.Errorf("goblin XP = %d, want 50", got[0].XP)
	}
	if got[1].XP != 25000 {
		t.Errorf("dragon XP = %d, want 25000", got[1].XP)
	}
}

func TestMonsterReader_FiltersByCRRange(t *testing.T) {
	docs := &fakeDocs{all: []map[string]any{
		monsterDocFull("5.5e", "topo", "Topo", "1/8", "1/8 (PE 25)", "Bestia"),
		monsterDocFull("5.5e", "goblin", "Goblin", "1", "1 (PE 200)", "Umanoide"),
		monsterDocFull("5.5e", "orso", "Orso", "2", "2 (PE 450)", "Bestia"),
		monsterDocFull("5.5e", "wyvern", "Wyvern", "5", "5 (PE 1.800)", "Drago"),
		monsterDocFull("5.5e", "yougoloth", "Ultroloth", "8", "8 (PE 3.900)", "Immondo"),
	}}
	r := NewMonsterReader(docs)

	got, _ := r.Search(context.Background(), monster.SearchQuery{
		Source: "5.5e",
		MinCR:  1,
		MaxCR:  5,
	})
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3 (goblin/orso/wyvern)", len(got))
	}
	names := []string{got[0].Name, got[1].Name, got[2].Name}
	want := []string{"Goblin", "Orso", "Wyvern"}
	for i, n := range names {
		if n != want[i] {
			t.Errorf("names[%d] = %q, want %q", i, n, want[i])
		}
	}
}

func TestMonsterReader_CRRangeOnlyUpperBound(t *testing.T) {
	docs := &fakeDocs{all: []map[string]any{
		monsterDocFull("5.5e", "topo", "Topo", "1/8", "1/8 (PE 25)", "Bestia"),
		monsterDocFull("5.5e", "wyvern", "Wyvern", "5", "5 (PE 1.800)", "Drago"),
		monsterDocFull("5.5e", "yougoloth", "Ultroloth", "8", "8 (PE 3.900)", "Immondo"),
	}}
	r := NewMonsterReader(docs)

	got, _ := r.Search(context.Background(), monster.SearchQuery{
		Source: "5.5e",
		MaxCR:  5,
	})
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
}

func TestMonsterReader_FiltersByType(t *testing.T) {
	docs := &fakeDocs{all: []map[string]any{
		monsterDocFull("5.5e", "goblin", "Goblin", "1", "1 (PE 200)", "Umanoide"),
		monsterDocFull("5.5e", "wyvern", "Wyvern", "5", "5 (PE 1.800)", "Drago"),
		monsterDocFull("5.5e", "drago-rosso", "Drago rosso", "17", "17 (PE 18.000)", "Drago"),
	}}
	r := NewMonsterReader(docs)

	got, _ := r.Search(context.Background(), monster.SearchQuery{
		Source: "5.5e",
		Type:   "Drago",
	})
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	for _, m := range got {
		if m.Type != "Drago" {
			t.Errorf("entry type = %q, want Drago", m.Type)
		}
	}
}

func TestMonsterReader_TypeFilterIsCaseInsensitive(t *testing.T) {
	docs := &fakeDocs{all: []map[string]any{
		monsterDocFull("5.5e", "wyvern", "Wyvern", "5", "5 (PE 1.800)", "Drago"),
	}}
	r := NewMonsterReader(docs)
	got, _ := r.Search(context.Background(), monster.SearchQuery{Source: "5.5e", Type: "DRAGO"})
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
}

func TestMonsterReader_CombinesAllFiltersWithAND(t *testing.T) {
	docs := &fakeDocs{all: []map[string]any{
		monsterDocFull("5.5e", "goblin", "Goblin", "1/4", "1/4 (PE 50)", "Umanoide"),
		monsterDocFull("5.5e", "goblin-boss", "Goblin Capo", "1", "1 (PE 200)", "Umanoide"),
		monsterDocFull("5.5e", "wyvern", "Wyvern", "5", "5 (PE 1.800)", "Drago"),
	}}
	r := NewMonsterReader(docs)

	got, _ := r.Search(context.Background(), monster.SearchQuery{
		Source: "5.5e",
		Query:  "gob",
		Type:   "Umanoide",
		MinCR:  1,
		MaxCR:  2,
	})
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Name != "Goblin Capo" {
		t.Errorf("got %q, want Goblin Capo", got[0].Name)
	}
}

func TestMonsterReader_Facets(t *testing.T) {
	docs := &fakeDocs{all: []map[string]any{
		monsterDocFull("5.5e", "goblin", "Goblin", "1/4", "1/4 (PE 50)", "Umanoide"),
		monsterDocFull("5.5e", "wyvern", "Wyvern", "5", "5 (PE 1.800)", "Drago"),
		monsterDocFull("5.5e", "drago-rosso", "Drago rosso", "17", "17 (PE 18.000)", "Drago"),
		monsterDocFull("5.5e", "topo", "Topo", "1/8", "1/8 (PE 25)", "Bestia"),
		monsterDocFull("5e", "lich", "Lich", "21", "21 (PE 33.000)", "Non morto"),
	}}
	r := NewMonsterReader(docs)

	facets, err := r.Facets(context.Background(), "5.5e")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	want := []string{"Bestia", "Drago", "Umanoide"}
	if len(facets.Types) != len(want) {
		t.Fatalf("types = %v, want %v", facets.Types, want)
	}
	for i, ty := range facets.Types {
		if ty != want[i] {
			t.Errorf("types[%d] = %q, want %q", i, ty, want[i])
		}
	}
}

func TestMonsterReader_FacetsEmptySource(t *testing.T) {
	docs := &fakeDocs{all: []map[string]any{
		monsterDocFull("5.5e", "goblin", "Goblin", "1/4", "1/4 (PE 50)", "Umanoide"),
	}}
	r := NewMonsterReader(docs)
	facets, err := r.Facets(context.Background(), "")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if len(facets.Types) != 0 {
		t.Errorf("types = %v, want empty", facets.Types)
	}
}

func TestMonsterReader_ExposesTypeOnMonster(t *testing.T) {
	docs := &fakeDocs{all: []map[string]any{
		monsterDocFull("5.5e", "wyvern", "Wyvern", "5", "5 (PE 1.800)", "Drago"),
	}}
	r := NewMonsterReader(docs)
	got, _ := r.Search(context.Background(), monster.SearchQuery{Source: "5.5e"})
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Type != "Drago" {
		t.Errorf("type = %q, want Drago", got[0].Type)
	}
}

func TestMonsterReader_FallsBackToRawContentWhenCRLacksXP(t *testing.T) {
	// 5e "cr" field sometimes holds only the bare rating; XP lives in
	// raw_content's "GS 1/4 (50 PE) ..." line.
	docs := &fakeDocs{all: []map[string]any{
		{
			"_id":           "sprite",
			"_source_short": "5e",
			"title":         "Folletto",
			"cr":            "1/4",
			"raw_content":   "## Folletto\n**GS** 1/4 (50 PE) Invisibilità...",
		},
	}}
	r := NewMonsterReader(docs)

	got, _ := r.Search(context.Background(), monster.SearchQuery{Source: "5e"})
	if len(got) != 1 || got[0].XP != 50 {
		t.Fatalf("got %+v, want XP=50", got)
	}
}
