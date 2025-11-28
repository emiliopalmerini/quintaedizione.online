package domain

import (
	"testing"
)

func TestNewDocumentFilters(t *testing.T) {
	filters := NewDocumentFilters()

	if filters == nil {
		t.Error("Expected non-nil DocumentFilters")
	}

	if len(filters) != 0 {
		t.Errorf("Expected empty filters, got %d items", len(filters))
	}
}

func TestDocumentFilters_Set(t *testing.T) {
	filters := NewDocumentFilters()

	filters.Set("collection", "spells")
	if filters["collection"] != "spells" {
		t.Errorf("Expected collection to be 'spells', got %v", filters["collection"])
	}

	filters.Set("level", 3)
	if filters["level"] != 3 {
		t.Errorf("Expected level to be 3, got %v", filters["level"])
	}

	filters.Set("collection", "weapons")
	if filters["collection"] != "weapons" {
		t.Errorf("Expected collection to be 'weapons', got %v", filters["collection"])
	}
}

func TestDocumentFilters_Get(t *testing.T) {
	filters := DocumentFilters{
		"collection": "spells",
		"level":      3,
		"school":     "evocation",
	}

	tests := []struct {
		name        string
		key         string
		expected    interface{}
		shouldExist bool
	}{
		{"existing string", "collection", "spells", true},
		{"existing int", "level", 3, true},
		{"existing school", "school", "evocation", true},
		{"non-existing key", "rarity", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, exists := filters.Get(tt.key)
			if exists != tt.shouldExist {
				t.Errorf("Expected exists=%t, got %t", tt.shouldExist, exists)
			}
			if exists && result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDocumentFilters_GetString(t *testing.T) {
	filters := DocumentFilters{
		"collection": "spells",
		"level":      3,
		"missing":    nil,
	}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"existing string", "collection", "spells"},
		{"int value", "level", ""},
		{"nil value", "missing", ""},
		{"non-existing key", "nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filters.GetString(tt.key)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDocumentFilters_Has(t *testing.T) {
	filters := DocumentFilters{
		"collection": "spells",
		"level":      3,
		"empty":      "",
		"nil":        nil,
	}

	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"existing non-empty", "collection", true},
		{"existing int", "level", true},
		{"existing empty string", "empty", true},
		{"nil value", "nil", true},
		{"non-existing key", "nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filters.Has(tt.key)
			if result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestDocumentFilters_Delete(t *testing.T) {
	filters := DocumentFilters{
		"collection": "spells",
		"level":      3,
		"school":     "evocation",
	}

	filters.Delete("level")
	if _, exists := filters.Get("level"); exists {
		t.Error("Expected level to be deleted")
	}

	if collection, exists := filters.Get("collection"); !exists || collection != "spells" {
		t.Error("Expected collection to still exist")
	}
	if school, exists := filters.Get("school"); !exists || school != "evocation" {
		t.Error("Expected school to still exist")
	}

	filters.Delete("nonexistent")
}
