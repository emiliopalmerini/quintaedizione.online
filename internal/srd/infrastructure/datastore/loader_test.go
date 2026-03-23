package datastore

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/application/parsers"
)

func newTestFS() fstest.MapFS {
	return fstest.MapFS{
		"srd-5.5e/source.json": &fstest.MapFile{
			Data: []byte(`{"id":"srd-5.5e","name":"SRD 5.2.1 (2024)","short_name":"5.5e","year":2024,"ruleset":"2024","xp_system":"2024","default":true}`),
		},
		"srd-5.5e/spells.json": &fstest.MapFile{
			Data: []byte(`[{"id":"palla-di-fuoco","name":"Palla di Fuoco","level":3,"school":"Evocazione","classes":["Mago"],"description":[{"type":"text","text":"Una palla di fuoco."}]}]`),
		},
		"srd-5.5e/monsters.json": &fstest.MapFile{
			Data: []byte(`[{"id":"goblin","name":"Goblin","type":"Umanoide","size":"Piccola","cr":"1/4","cr_detail":"1/4 (PE 50; BC +2)","alignment":"neutrale malvagio"}]`),
		},
		"srd-5.5e/classes.json":        &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/backgrounds.json":    &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/equipment.json":      &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/magic_items.json":    &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/feats.json":          &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/rules_gameplay.json": &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/rules_creation.json": &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/rules_tools.json":    &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/glossary.json":       &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/species.json":        &fstest.MapFile{Data: []byte(`[]`)},
	}
}

func newMultiSourceFS() fstest.MapFS {
	fs := newTestFS()
	fs["srd-5e/source.json"] = &fstest.MapFile{
		Data: []byte(`{"id":"srd-5e","name":"SRD 5.0 (2014)","short_name":"5e","year":2014,"ruleset":"2014","xp_system":"2014","default":false}`),
	}
	fs["srd-5e/spells.json"] = &fstest.MapFile{
		Data: []byte(`[{"id":"palla-di-fuoco","name":"Palla di Fuoco","level":3,"school":"Evocazione","classes":["Mago"],"description":[{"type":"text","text":"Una palla di fuoco (versione 2014)."}]}]`),
	}
	fs["srd-5e/monsters.json"] = &fstest.MapFile{
		Data: []byte(`[{"id":"goblin","name":"Goblin","type":"Umanoide","size":"Piccola","cr":"1/4","cr_detail":"1/4 (PE 50; BC +2)","alignment":"legale malvagio"}]`),
	}
	fs["srd-5e/classes.json"] = &fstest.MapFile{Data: []byte(`[]`)}
	fs["srd-5e/backgrounds.json"] = &fstest.MapFile{Data: []byte(`[]`)}
	fs["srd-5e/equipment.json"] = &fstest.MapFile{Data: []byte(`[]`)}
	fs["srd-5e/magic_items.json"] = &fstest.MapFile{Data: []byte(`[]`)}
	fs["srd-5e/feats.json"] = &fstest.MapFile{Data: []byte(`[]`)}
	fs["srd-5e/rules_gameplay.json"] = &fstest.MapFile{Data: []byte(`[]`)}
	fs["srd-5e/rules_creation.json"] = &fstest.MapFile{Data: []byte(`[]`)}
	fs["srd-5e/rules_tools.json"] = &fstest.MapFile{Data: []byte(`[]`)}
	fs["srd-5e/glossary.json"] = &fstest.MapFile{Data: []byte(`[]`)}
	fs["srd-5e/species.json"] = &fstest.MapFile{Data: []byte(`[]`)}
	return fs
}

func TestLoadAll_DiscoversSources(t *testing.T) {
	loader := NewLoader(newTestFS(), parsers.NewMarkdownRenderer(nil))
	_, sources, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(sources))
	}
	if sources[0].ID != "srd-5.5e" {
		t.Errorf("expected source id srd-5.5e, got %s", sources[0].ID)
	}
	if sources[0].ShortName != "5.5e" {
		t.Errorf("expected short_name 5.5e, got %s", sources[0].ShortName)
	}
}

func TestLoadAll_TagsDocumentsWithSource(t *testing.T) {
	loader := NewLoader(newTestFS(), parsers.NewMarkdownRenderer(nil))
	data, _, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	spells := data["incantesimi"]
	if len(spells) != 1 {
		t.Fatalf("expected 1 spell, got %d", len(spells))
	}

	src, ok := spells[0]["_source"].(string)
	if !ok || src != "srd-5.5e" {
		t.Errorf("expected _source=srd-5.5e, got %v", spells[0]["_source"])
	}

	short, ok := spells[0]["_source_short"].(string)
	if !ok || short != "5.5e" {
		t.Errorf("expected _source_short=5.5e, got %v", spells[0]["_source_short"])
	}

	// _id should stay clean (no prefix)
	id, ok := spells[0]["_id"].(string)
	if !ok || id != "palla-di-fuoco" {
		t.Errorf("expected _id=palla-di-fuoco, got %v", spells[0]["_id"])
	}
}

func TestLoadAll_MultipleSourcesMergeDocs(t *testing.T) {
	loader := NewLoader(newMultiSourceFS(), parsers.NewMarkdownRenderer(nil))
	data, sources, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	if len(sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(sources))
	}

	// Spells from both sources should be merged into the same collection
	spells := data["incantesimi"]
	if len(spells) != 2 {
		t.Fatalf("expected 2 spells (one per source), got %d", len(spells))
	}

	// Both should have the same title but different source tags
	sourceSet := make(map[string]bool)
	for _, spell := range spells {
		sourceSet[spell["_source"].(string)] = true
	}
	if !sourceSet["srd-5.5e"] || !sourceSet["srd-5e"] {
		t.Errorf("expected spells from both sources, got %v", sourceSet)
	}
}

func TestLoadAll_MultipleSourcesMergeMonsters(t *testing.T) {
	loader := NewLoader(newMultiSourceFS(), parsers.NewMarkdownRenderer(nil))
	data, _, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	monsters := data["mostri"]
	if len(monsters) != 2 {
		t.Fatalf("expected 2 monsters (one per source), got %d", len(monsters))
	}
}

func TestLoadAll_SkipsDirectoryWithoutSourceJSON(t *testing.T) {
	fs := newTestFS()
	fs["no-source-dir/spells.json"] = &fstest.MapFile{Data: []byte(`[]`)}

	loader := NewLoader(fs, parsers.NewMarkdownRenderer(nil))
	_, sources, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	if len(sources) != 1 {
		t.Errorf("expected 1 source (skipping dir without source.json), got %d", len(sources))
	}
}

func TestToMarkdown_TextOnly(t *testing.T) {
	c := jsonContent{{Type: "text", Text: "Testo semplice."}}
	got := c.toMarkdown("5.5e")
	want := "Testo semplice."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestToMarkdown_Empty(t *testing.T) {
	var c jsonContent
	if got := c.toMarkdown("5.5e"); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestToMarkdown_SpellReference(t *testing.T) {
	c := jsonContent{
		{Type: "text", Text: "lancia "},
		{Type: "spell", Text: "guida", ID: "guida"},
		{Type: "text", Text: " sul bersaglio"},
	}
	got := c.toMarkdown("5.5e")
	want := "lancia [guida](/srd/incantesimi/5.5e/guida) sul bersaglio"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestToMarkdown_ConditionReference(t *testing.T) {
	c := jsonContent{{Type: "condition", Text: "accecato", ID: "accecato"}}
	got := c.toMarkdown("5e")
	want := "[accecato](/srd/glossario/5e/accecato)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestToMarkdown_RuleLinkTypes(t *testing.T) {
	tests := []struct {
		segType  string
		wantRule string
	}{
		{"damage_type", "tipi-di-danno"},
		{"ability", "le-sei-caratteristiche"},
		{"skill", "competenze-nelle-abilita"},
	}
	for _, tt := range tests {
		c := jsonContent{{Type: tt.segType, Text: "test", ID: "test-id"}}
		got := c.toMarkdown("5e")
		want := "[test](/srd/regole/5e/" + tt.wantRule + ")"
		if got != want {
			t.Errorf("type %s: got %q, want %q", tt.segType, got, want)
		}
	}
}

func TestToMarkdown_CreatureTypeIsPlainText(t *testing.T) {
	c := jsonContent{{Type: "creature_type", Text: "Bestia", ID: "bestia"}}
	got := c.toMarkdown("5e")
	if got != "Bestia" {
		t.Errorf("creature_type should be plain text, got %q", got)
	}
}

func TestToMarkdown_RuleLinkUsesSource(t *testing.T) {
	c := jsonContent{{Type: "ability", Text: "Forza", ID: "forza"}}
	got55 := c.toMarkdown("5.5e")
	got5 := c.toMarkdown("5e")
	if !strings.Contains(got55, `/srd/regole/5.5e/`) {
		t.Errorf("5.5e: expected source in link, got %q", got55)
	}
	if !strings.Contains(got5, `/srd/regole/5e/`) {
		t.Errorf("5e: expected source in link, got %q", got5)
	}
}

func TestToMarkdown_RuleLinkWithoutID(t *testing.T) {
	c := jsonContent{{Type: "ability", Text: "Forza"}}
	got := c.toMarkdown("5e")
	if got != "Forza" {
		t.Errorf("expected plain text for rule-link without ID, got %q", got)
	}
}

func TestToMarkdown_AbilityInsideEmphasis(t *testing.T) {
	// Ability references inside *italic* should produce valid markdown links
	c := jsonContent{
		{Type: "text", Text: "*Tiro salvezza su "},
		{Type: "ability", Text: "Saggezza", ID: "saggezza"},
		{Type: "text", Text: ":* CD 16"},
	}
	md := c.toMarkdown("5.5e")
	if !strings.Contains(md, "[Saggezza](/srd/regole/5.5e/le-sei-caratteristiche)") {
		t.Errorf("expected markdown link for ability, got %q", md)
	}
}

func TestToMarkdown_EquipmentReference(t *testing.T) {
	c := jsonContent{{Type: "equipment", Text: "spada lunga", ID: "spada-lunga"}}
	got := c.toMarkdown("5.5e")
	want := "[spada lunga](/srd/equipaggiamenti/5.5e/spada-lunga)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestToMarkdown_RuleReference(t *testing.T) {
	c := jsonContent{{Type: "rule", Text: "azioni", ID: "azioni"}}
	got := c.toMarkdown("5.5e")
	want := "[azioni](/srd/regole/5.5e/azioni)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestToMarkdown_SourceAwareness(t *testing.T) {
	c := jsonContent{{Type: "spell", Text: "guida", ID: "guida"}}

	got55 := c.toMarkdown("5.5e")
	got5 := c.toMarkdown("5e")

	if got55 == got5 {
		t.Error("expected different URLs for different sources")
	}
	if got55 != "[guida](/srd/incantesimi/5.5e/guida)" {
		t.Errorf("5.5e: got %q", got55)
	}
	if got5 != "[guida](/srd/incantesimi/5e/guida)" {
		t.Errorf("5e: got %q", got5)
	}
}

func TestToMarkdown_EscapesMarkdownChars(t *testing.T) {
	c := jsonContent{{Type: "spell", Text: "test]spell)", ID: "test"}}
	got := c.toMarkdown("5.5e")
	want := `[test\]spell\)](/srd/incantesimi/5.5e/test)`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestToMarkdown_ReferenceWithoutID(t *testing.T) {
	c := jsonContent{{Type: "spell", Text: "incantesimo"}}
	got := c.toMarkdown("5.5e")
	want := "incantesimo"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestToMarkdown_UnknownType(t *testing.T) {
	c := jsonContent{{Type: "unknown", Text: "qualcosa", ID: "id"}}
	got := c.toMarkdown("5.5e")
	want := "qualcosa"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLoadAll_SpellContentHasCrosslinks(t *testing.T) {
	fsys := fstest.MapFS{
		"srd-5.5e/source.json": &fstest.MapFile{
			Data: []byte(`{"id":"srd-5.5e","name":"SRD 5.2.1 (2024)","short_name":"5.5e","year":2024,"ruleset":"2024","xp_system":"2024","default":true}`),
		},
		"srd-5.5e/spells.json": &fstest.MapFile{
			Data: []byte(`[{"id":"aculeo-mentale","name":"Aculeo Mentale","level":2,"school":"Divinazione","classes":["Mago"],"description":[{"type":"text","text":"Se è "},{"type":"condition","text":"accecato","id":"accecato"},{"type":"text","text":", subisce danni psichici."}]}]`),
		},
		"srd-5.5e/monsters.json":       &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/classes.json":        &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/backgrounds.json":    &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/equipment.json":      &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/magic_items.json":    &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/feats.json":          &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/rules_gameplay.json": &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/glossary.json":       &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/species.json":        &fstest.MapFile{Data: []byte(`[]`)},
	}

	loader := NewLoader(fsys, parsers.NewMarkdownRenderer(nil))
	data, _, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	spells := data["incantesimi"]
	if len(spells) != 1 {
		t.Fatalf("expected 1 spell, got %d", len(spells))
	}

	// description_html should contain rendered HTML with crosslinks
	content, ok := spells[0]["description_html"].(string)
	if !ok {
		t.Fatal("description_html is not a string")
	}

	if !strings.Contains(content, `/srd/glossario/5.5e/accecato`) {
		t.Errorf("expected crosslink to glossario/5.5e/accecato in description_html, got:\n%s", content)
	}

	// raw_content should NOT contain links (just plain text)
	raw, ok := spells[0]["raw_content"].(string)
	if !ok {
		t.Fatal("raw_content is not a string")
	}
	if strings.Contains(raw, "/srd/glossario/") {
		t.Errorf("raw_content should not contain crosslinks, got:\n%s", raw)
	}
}

func TestLoadAll_MultiSourceCrosslinksUseCorrectSource(t *testing.T) {
	fsys := newMultiSourceFS()
	// Override 5e spells with a reference segment
	fsys["srd-5e/spells.json"] = &fstest.MapFile{
		Data: []byte(`[{"id":"guida","name":"Guida","level":0,"school":"Divinazione","classes":["Chierico"],"description":[{"type":"text","text":"Se è "},{"type":"condition","text":"accecato","id":"accecato"},{"type":"text","text":" non funziona."}]}]`),
	}

	loader := NewLoader(fsys, parsers.NewMarkdownRenderer(nil))
	data, _, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	spells := data["incantesimi"]
	for _, spell := range spells {
		src := spell["_source_short"].(string)
		content := spell["description_html"].(string)
		if strings.Contains(content, "/srd/glossario/") {
			// The link should use the spell's own source, not the other one
			expectedPath := "/srd/glossario/" + src + "/accecato"
			if !strings.Contains(content, expectedPath) {
				t.Errorf("source %s: expected crosslink with path %s, got:\n%s", src, expectedPath, content)
			}
		}
	}
}

func TestLoadAll_SourceDefaultFlag(t *testing.T) {
	loader := NewLoader(newMultiSourceFS(), parsers.NewMarkdownRenderer(nil))
	_, sources, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	defaultCount := 0
	for _, src := range sources {
		if src.Default {
			defaultCount++
			if src.ID != "srd-5.5e" {
				t.Errorf("expected default source to be srd-5.5e, got %s", src.ID)
			}
		}
	}
	if defaultCount != 1 {
		t.Errorf("expected exactly 1 default source, got %d", defaultCount)
	}
}
