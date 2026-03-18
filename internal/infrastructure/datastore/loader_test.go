package datastore

import (
	"testing"
	"testing/fstest"

	"github.com/emiliopalmerini/quintaedizione.online/internal/application/parsers"
)

func newTestFS() fstest.MapFS {
	return fstest.MapFS{
		"srd-5.5e/source.json": &fstest.MapFile{
			Data: []byte(`{"id":"srd-5.5e","name":"SRD 5.2.1 (2024)","short_name":"5.5e","year":2024,"ruleset":"2024","xp_system":"2024","default":true}`),
		},
		"srd-5.5e/spells.json": &fstest.MapFile{
			Data: []byte(`[{"id":"palla-di-fuoco","name":"Palla di Fuoco","level":3,"school":"Evocazione","classes":["Mago"],"description":"Una palla di fuoco."}]`),
		},
		"srd-5.5e/monsters.json": &fstest.MapFile{
			Data: []byte(`[{"id":"goblin","name":"Goblin","type":"Umanoide","size":"Piccola","cr":"1/4","cr_detail":"1/4 (PE 50; BC +2)","alignment":"neutrale malvagio"}]`),
		},
		"srd-5.5e/classes.json":      &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/backgrounds.json":  &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/equipment.json":    &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/magic_items.json":  &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/feats.json":        &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/rules_gameplay.json": &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/rules_creation.json": &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/rules_tools.json":    &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/glossary.json":     &fstest.MapFile{Data: []byte(`[]`)},
		"srd-5.5e/species.json":      &fstest.MapFile{Data: []byte(`[]`)},
	}
}

func newMultiSourceFS() fstest.MapFS {
	fs := newTestFS()
	fs["srd-5e/source.json"] = &fstest.MapFile{
		Data: []byte(`{"id":"srd-5e","name":"SRD 5.0 (2014)","short_name":"5e","year":2014,"ruleset":"2014","xp_system":"2014","default":false}`),
	}
	fs["srd-5e/spells.json"] = &fstest.MapFile{
		Data: []byte(`[{"id":"palla-di-fuoco","name":"Palla di Fuoco","level":3,"school":"Evocazione","classes":["Mago"],"description":"Una palla di fuoco (versione 2014)."}]`),
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

	// _id should be prefixed with source
	id, ok := spells[0]["_id"].(string)
	if !ok || id != "srd-5.5e--palla-di-fuoco" {
		t.Errorf("expected _id=srd-5.5e--palla-di-fuoco, got %v", spells[0]["_id"])
	}

	// _original_id should be preserved
	origID, ok := spells[0]["_original_id"].(string)
	if !ok || origID != "palla-di-fuoco" {
		t.Errorf("expected _original_id=palla-di-fuoco, got %v", spells[0]["_original_id"])
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
