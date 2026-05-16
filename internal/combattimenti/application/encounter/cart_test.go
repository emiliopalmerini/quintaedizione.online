package encounter

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/monster"
	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/infrastructure/persistence/memory"
)

type stubReader struct {
	m map[string]monster.Monster
}

func (s *stubReader) Search(context.Context, monster.SearchQuery) ([]monster.Monster, error) {
	return nil, nil
}

func (s *stubReader) FindByID(_ context.Context, source, id string) (monster.Monster, error) {
	if m, ok := s.m[source+"/"+id]; ok {
		return m, nil
	}
	return monster.Monster{}, errors.New("not found")
}

func silentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestParseCartRefs_DropsMalformed(t *testing.T) {
	got := ParseCartRefs([]string{"goblin@5.5e", "no-at-sign", "@missing-id", "missing-src@", "  ancient-dragon@5e "})
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2; got %+v", len(got), got)
	}
	if got[0].ID != "goblin" || got[0].Source != "5.5e" || got[0].Quantity != 1 {
		t.Errorf("got[0] = %+v", got[0])
	}
	if got[1].ID != "ancient-dragon" || got[1].Source != "5e" || got[1].Quantity != 1 {
		t.Errorf("got[1] = %+v", got[1])
	}
}

func TestParseCartRefs_FoldsDuplicatesByIDAndSource(t *testing.T) {
	got := ParseCartRefs([]string{
		"goblin@5e",
		"goblin@5e",
		"goblin@5e",
		"orc@5e",
	})
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2; got %+v", len(got), got)
	}
	if got[0].ID != "goblin" || got[0].Source != "5e" || got[0].Quantity != 3 {
		t.Errorf("got[0] = %+v, want goblin/5e/3", got[0])
	}
	if got[1].ID != "orc" || got[1].Source != "5e" || got[1].Quantity != 1 {
		t.Errorf("got[1] = %+v, want orc/5e/1", got[1])
	}
}

func TestParseCartRefs_DoesNotFoldAcrossSources(t *testing.T) {
	got := ParseCartRefs([]string{"goblin@5e", "goblin@5.5e"})
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2; got %+v", len(got), got)
	}
	for _, ref := range got {
		if ref.Quantity != 1 {
			t.Errorf("ref %+v quantity = %d, want 1", ref, ref.Quantity)
		}
	}
}

func TestCartPricer_2024NoMultiplier(t *testing.T) {
	reader := &stubReader{m: map[string]monster.Monster{
		"5.5e/goblin":  {ID: "goblin", Source: "5.5e", Name: "Goblin", XP: 50},
		"5.5e/bugbear": {ID: "bugbear", Source: "5.5e", Name: "Bugbear", XP: 200},
	}}
	pricer := NewCartPricer(silentLogger(), reader, memory.NewEncounterRepository())

	got, err := pricer.Price(context.Background(), PriceCartRequest{
		Ruleset: "2024",
		Budget:  500,
		Refs:    []CartItemRef{{ID: "goblin", Source: "5.5e", Quantity: 1}, {ID: "bugbear", Source: "5.5e", Quantity: 1}},
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.Subtotal != 250 {
		t.Errorf("subtotal = %d, want 250", got.Subtotal)
	}
	if got.EffectiveCost != 250 {
		t.Errorf("effective = %d, want 250", got.EffectiveCost)
	}
	if got.Remaining != 250 {
		t.Errorf("remaining = %d, want 250", got.Remaining)
	}
}

func TestCartPricer_2014AppliesMultiplier(t *testing.T) {
	reader := &stubReader{m: map[string]monster.Monster{
		"5e/goblin":  {ID: "goblin", Source: "5e", Name: "Goblin", XP: 50},
		"5e/bugbear": {ID: "bugbear", Source: "5e", Name: "Bugbear", XP: 200},
	}}
	pricer := NewCartPricer(silentLogger(), reader, memory.NewEncounterRepository())

	got, err := pricer.Price(context.Background(), PriceCartRequest{
		Ruleset: "2014",
		Budget:  500,
		Refs: []CartItemRef{
			{ID: "goblin", Source: "5e", Quantity: 1},
			{ID: "bugbear", Source: "5e", Quantity: 1},
		},
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.Subtotal != 250 {
		t.Errorf("subtotal = %d, want 250", got.Subtotal)
	}
	// 2 monsters → multiplier 1.5 → 250 * 1.5 = 375
	if got.EffectiveCost != 375 {
		t.Errorf("effective = %d, want 375", got.EffectiveCost)
	}
	if got.Remaining != 125 {
		t.Errorf("remaining = %d, want 125", got.Remaining)
	}
}

func TestCartPricer_DropsUnknownEntries(t *testing.T) {
	reader := &stubReader{m: map[string]monster.Monster{
		"5.5e/goblin": {ID: "goblin", Source: "5.5e", Name: "Goblin", XP: 50},
	}}
	pricer := NewCartPricer(silentLogger(), reader, memory.NewEncounterRepository())

	got, _ := pricer.Price(context.Background(), PriceCartRequest{
		Ruleset: "2024",
		Budget:  500,
		Refs: []CartItemRef{
			{ID: "goblin", Source: "5.5e", Quantity: 1},
			{ID: "ghost-monster", Source: "5.5e", Quantity: 1},
		},
	})
	if len(got.Entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(got.Entries))
	}
}

func TestCartPricer_EmptyCart(t *testing.T) {
	reader := &stubReader{m: map[string]monster.Monster{}}
	pricer := NewCartPricer(silentLogger(), reader, memory.NewEncounterRepository())

	got, err := pricer.Price(context.Background(), PriceCartRequest{
		Ruleset: "2024",
		Budget:  500,
		Refs:    nil,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.EffectiveCost != 0 || got.Remaining != 500 {
		t.Errorf("got %+v, want effective=0/remaining=500", got)
	}
}

func TestCartPricer_InvalidRuleset(t *testing.T) {
	pricer := NewCartPricer(silentLogger(), &stubReader{}, memory.NewEncounterRepository())
	_, err := pricer.Price(context.Background(), PriceCartRequest{Ruleset: "bogus"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCartPricer_2024QuantitiesMultiplyUnitXP(t *testing.T) {
	reader := &stubReader{m: map[string]monster.Monster{
		"5.5e/goblin": {ID: "goblin", Source: "5.5e", Name: "Goblin", XP: 50},
	}}
	pricer := NewCartPricer(silentLogger(), reader, memory.NewEncounterRepository())

	got, err := pricer.Price(context.Background(), PriceCartRequest{
		Ruleset: "2024",
		Budget:  500,
		Refs:    []CartItemRef{{ID: "goblin", Source: "5.5e", Quantity: 3}},
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(got.Entries) != 1 {
		t.Fatalf("entries = %d, want 1 (folded)", len(got.Entries))
	}
	if got.Entries[0].Quantity != 3 {
		t.Errorf("entry quantity = %d, want 3", got.Entries[0].Quantity)
	}
	if got.Subtotal != 150 {
		t.Errorf("subtotal = %d, want 150", got.Subtotal)
	}
	if got.EffectiveCost != 150 {
		t.Errorf("effective = %d, want 150", got.EffectiveCost)
	}
}

func TestCartPricer_2014MultiplierUsesSummedQuantity(t *testing.T) {
	reader := &stubReader{m: map[string]monster.Monster{
		"5e/goblin":  {ID: "goblin", Source: "5e", Name: "Goblin", XP: 50},
		"5e/bugbear": {ID: "bugbear", Source: "5e", Name: "Bugbear", XP: 200},
	}}
	pricer := NewCartPricer(silentLogger(), reader, memory.NewEncounterRepository())

	got, err := pricer.Price(context.Background(), PriceCartRequest{
		Ruleset: "2014",
		Budget:  1000,
		Refs: []CartItemRef{
			{ID: "goblin", Source: "5e", Quantity: 2},
			{ID: "bugbear", Source: "5e", Quantity: 1},
		},
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.Subtotal != 300 {
		t.Errorf("subtotal = %d, want 300", got.Subtotal)
	}
	if got.EffectiveCost != 600 {
		t.Errorf("effective = %d, want 600 (multiplier must see 3, not 2)", got.EffectiveCost)
	}
	if got.Remaining != 400 {
		t.Errorf("remaining = %d, want 400", got.Remaining)
	}
}
