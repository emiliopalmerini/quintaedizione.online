package domain

import (
	"testing"
)

func TestDocumentFromMap(t *testing.T) {
	m := map[string]any{
		"_id":           "palla-di-fuoco",
		"title":         "Palla di Fuoco",
		"content":       "<p>Una palla di fuoco.</p>",
		"raw_content":   "Una palla di fuoco.",
		"_source_short": "5.5e",
		"_source":       "srd-5.5e",
		"livello":       3,
		"scuola":        "Evocazione",
	}

	doc := DocumentFromMap(m)

	if doc.ID != "palla-di-fuoco" {
		t.Errorf("expected ID 'palla-di-fuoco', got %q", doc.ID)
	}
	if doc.Title != "Palla di Fuoco" {
		t.Errorf("expected title 'Palla di Fuoco', got %q", doc.Title)
	}
	if doc.Source != "5.5e" {
		t.Errorf("expected source '5.5e', got %q", doc.Source)
	}
	if doc.Content.String() != "<p>Una palla di fuoco.</p>" {
		t.Errorf("expected content, got %q", doc.Content)
	}
	if doc.RawContent.String() != "Una palla di fuoco." {
		t.Errorf("expected raw_content, got %q", doc.RawContent)
	}

	// Core fields should NOT be in Fields
	if _, ok := doc.Fields["_id"]; ok {
		t.Error("_id should not be in Fields")
	}
	if _, ok := doc.Fields["title"]; ok {
		t.Error("title should not be in Fields")
	}
	if _, ok := doc.Fields["_source_short"]; ok {
		t.Error("_source_short should not be in Fields")
	}

	// Collection-specific fields should be in Fields
	if doc.Fields["livello"] != 3 {
		t.Errorf("expected Fields[livello]=3, got %v", doc.Fields["livello"])
	}
	if doc.Fields["scuola"] != "Evocazione" {
		t.Errorf("expected Fields[scuola]=Evocazione, got %v", doc.Fields["scuola"])
	}
}

func TestDocument_GetField(t *testing.T) {
	doc := &Document{
		Fields: map[string]any{
			"livello": 3,
			"scuola":  "Evocazione",
		},
	}

	if doc.GetField("livello") != 3 {
		t.Error("expected livello=3")
	}
	if doc.GetField("missing") != nil {
		t.Error("expected nil for missing field")
	}
}

func TestDocument_GetFieldString(t *testing.T) {
	doc := &Document{
		Fields: map[string]any{
			"scuola":  "Evocazione",
			"livello": 3,
		},
	}

	if doc.GetFieldString("scuola") != "Evocazione" {
		t.Error("expected scuola='Evocazione'")
	}
	if doc.GetFieldString("livello") != "" {
		t.Error("expected empty string for non-string field")
	}
	if doc.GetFieldString("missing") != "" {
		t.Error("expected empty string for missing field")
	}
}
