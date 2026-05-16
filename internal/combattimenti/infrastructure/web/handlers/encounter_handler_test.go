package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/application/encounter"
	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/monster"
	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/infrastructure/persistence/memory"
	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/infrastructure/web/templates"
)

// stubMonsterReader is a thin in-memory reader sufficient for handler tests.
// Mirrors the stubReader in application/encounter/cart_test.go.
type stubMonsterReader struct {
	monsters map[string]monster.Monster
}

func (s *stubMonsterReader) Search(_ context.Context, q monster.SearchQuery) ([]monster.Monster, error) {
	out := make([]monster.Monster, 0)
	for _, m := range s.monsters {
		if m.Source != q.Source {
			continue
		}
		if q.OnlyAfford && q.MaxXP > 0 && m.XP > q.MaxXP {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

func (s *stubMonsterReader) FindByID(_ context.Context, source, id string) (monster.Monster, error) {
	if m, ok := s.monsters[source+"/"+id]; ok {
		return m, nil
	}
	return monster.Monster{}, errors.New("monster: not found")
}

func (s *stubMonsterReader) Facets(_ context.Context, _ string) (monster.FacetSet, error) {
	return monster.FacetSet{}, nil
}

func newTestHandler(t *testing.T) (*EncounterHandler, []templates.EditionOption) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	repo := memory.NewEncounterRepository()
	svc := encounter.NewService(logger, repo)
	qh := encounter.NewQueryHandler(logger, repo)
	reader := &stubMonsterReader{monsters: map[string]monster.Monster{
		"5.5e/aboleth": {ID: "aboleth", Source: "5.5e", Name: "Aboleth", CR: "10", XP: 5900},
		"5.5e/goblin":  {ID: "goblin", Source: "5.5e", Name: "Goblin", CR: "1/4", XP: 50},
	}}
	pricer := encounter.NewCartPricer(logger, reader, repo)

	editions := []templates.EditionOption{
		{SourceID: "srd-5.5e", Name: "SRD 5.2.1 (2024)", ShortName: "5.5e", Ruleset: "2024", IsDefault: true},
		{SourceID: "srd-5e", Name: "SRD 5.1 (2014)", ShortName: "5e", Ruleset: "2014"},
	}
	return NewEncounterHandler(svc, qh, pricer, reader, logger), editions
}

func TestHomeHandler_EmptyQueryRendersPlaceholder(t *testing.T) {
	h, editions := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/combattimenti", nil)
	rec := httptest.NewRecorder()

	h.HomeHandler(editions).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	// Default form values should appear.
	if !strings.Contains(body, `name="level"`) {
		t.Errorf("missing level input")
	}
	// Empty querystring → no server-side prerender → placeholder visible.
	if !strings.Contains(body, "result-placeholder") {
		t.Errorf("expected placeholder; body did not contain it")
	}
	if strings.Contains(body, "result-card success") {
		t.Errorf("did not expect prerendered result card on empty URL")
	}
}

func TestHomeHandler_HydratesFromURL2024High(t *testing.T) {
	h, editions := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/combattimenti?party=4,4,4,4&ruleset=2024&diff=High", nil)
	rec := httptest.NewRecorder()

	h.HomeHandler(editions).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()

	// 2024 selected.
	if !strings.Contains(body, `value="2024" data-source="srd-5.5e" data-source-short="5.5e" checked`) {
		t.Errorf("2024 radio not checked")
	}
	// Level seeded to 4.
	if !strings.Contains(body, `id="party-level" name="level" value="4"`) {
		t.Errorf("party level not seeded to 4")
	}
	// Difficulty High selected.
	if !strings.Contains(body, `<option value="High" selected>Alta</option>`) {
		t.Errorf("High difficulty not selected")
	}
	// Server-side prerender ran.
	if !strings.Contains(body, "result-card success") {
		t.Errorf("expected prerendered result card; body:\n%s", body[:min(2000, len(body))])
	}
	if strings.Contains(body, "result-placeholder") {
		t.Errorf("unexpected placeholder when URL has params")
	}
}

func TestHomeHandler_HydratesCartFromURL(t *testing.T) {
	h, editions := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet,
		"/combattimenti?ruleset=2024&party=3,3,3,3&cart=goblin@5.5e:3", nil)
	rec := httptest.NewRecorder()

	h.HomeHandler(editions).ServeHTTP(rec, req)

	body := rec.Body.String()
	// Three hidden cart inputs for goblin@5.5e (one per qty).
	count := strings.Count(body, `value="goblin@5.5e"`)
	if count != 3 {
		t.Errorf("want 3 hidden monsters[] inputs, got %d", count)
	}
	// Cart entries appear in prerendered result.
	if !strings.Contains(body, "Goblin") {
		t.Errorf("Goblin not present in rendered cart")
	}
}

func TestHomeHandler_DropsForeignSourceCart(t *testing.T) {
	h, editions := newTestHandler(t)
	// Ruleset 2024 (source 5.5e) but cart points to source 5e — should be dropped.
	req := httptest.NewRequest(http.MethodGet,
		"/combattimenti?ruleset=2024&cart=goblin@5e:2", nil)
	rec := httptest.NewRecorder()

	h.HomeHandler(editions).ServeHTTP(rec, req)

	body := rec.Body.String()
	if strings.Contains(body, `value="goblin@5e"`) {
		t.Errorf("foreign-source cart entry should be dropped")
	}
}

func TestHomeHandler_HydratesFromURL2014(t *testing.T) {
	h, editions := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet,
		"/combattimenti?ruleset=2014&party=5,5,5,5&diff=Difficile", nil)
	rec := httptest.NewRecorder()

	h.HomeHandler(editions).ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, `value="2014" data-source="srd-5e" data-source-short="5e" checked`) {
		t.Errorf("2014 radio not checked")
	}
	if !strings.Contains(body, `<option value="Difficile" selected>Difficile</option>`) {
		t.Errorf("Difficile not selected")
	}
}

func TestHomeHandler_DifferentLevelsMode(t *testing.T) {
	h, editions := newTestHandler(t)
	// Party with mixed levels should flip into "different" mode.
	req := httptest.NewRequest(http.MethodGet,
		"/combattimenti?party=1,2,3,4&ruleset=2024", nil)
	rec := httptest.NewRecorder()

	h.HomeHandler(editions).ServeHTTP(rec, req)

	body := rec.Body.String()
	// "different" radio is checked.
	if !strings.Contains(body, `value="different" checked`) {
		t.Errorf("different mode not selected")
	}
	// Per-character inputs seeded with the right values.
	for _, want := range []string{`value="1"`, `value="2"`, `value="3"`, `value="4"`} {
		if !strings.Contains(body, `name="character_levels" `+want) {
			t.Errorf("missing per-character input with %s", want)
		}
	}
}

func TestCalculateHandler_SetsHXPushURL(t *testing.T) {
	h, _ := newTestHandler(t)
	form := strings.NewReader("ruleset=2024&party_mode=same&level=4&count=4&difficulty_2024=High&source_short=5.5e")
	req := httptest.NewRequest(http.MethodPost, "/combattimenti/calculate", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	h.CalculateHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	got := rec.Header().Get("HX-Push-Url")
	if got == "" {
		t.Fatal("HX-Push-Url not set")
	}
	if !strings.HasPrefix(got, "/combattimenti?") {
		t.Errorf("unexpected HX-Push-Url prefix: %q", got)
	}
	// party=4,4,4,4 is non-default; ruleset and diff are non-default too.
	for _, fragment := range []string{"party=4%2C4%2C4%2C4", "diff=High"} {
		if !strings.Contains(got, fragment) {
			t.Errorf("HX-Push-Url %q missing fragment %q", got, fragment)
		}
	}
}

func TestCalculateHandler_HXPushURLDropsForeignCart(t *testing.T) {
	h, _ := newTestHandler(t)
	// cart entry's source ("5e") doesn't match active ruleset's source ("5.5e")
	// → filterRefsBySource drops it → URL should not contain it either.
	body := "ruleset=2024&party_mode=same&level=3&count=4&difficulty_2024=Moderate&source_short=5.5e&monsters%5B%5D=goblin@5e"
	req := httptest.NewRequest(http.MethodPost, "/combattimenti/calculate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	h.CalculateHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	got := rec.Header().Get("HX-Push-Url")
	if strings.Contains(got, "cart=") {
		t.Errorf("HX-Push-Url should not include cart= when only foreign-source entries: %q", got)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
