package domain

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDocument_NewDocument(t *testing.T) {
	id := DocumentID("test-id")
	title := "Test Document"
	filters := NewDocumentFilters()
	content := NewHTMLContent("<p>test content</p>")
	rawContent := NewMarkdownContent("test content")

	doc := NewDocument(id, title, filters, content, rawContent)

	if doc.ID != id {
		t.Errorf("Expected ID %s, got %s", id, doc.ID)
	}
	if doc.Title != title {
		t.Errorf("Expected title %s, got %s", title, doc.Title)
	}
	if !cmp.Equal(doc.Filters, filters) {
		t.Errorf("Expected filters to match")
	}
	if string(doc.Content) != string(content) {
		t.Errorf("Expected content to match")
	}
	if string(doc.RawContent) != string(rawContent) {
		t.Errorf("Expected raw content to match")
	}
}

func TestDocument_NewDocumentFromName(t *testing.T) {
	title := "Test Document Title"
	filters := NewDocumentFilters()
	content := NewHTMLContent("<p>test content</p>")
	rawContent := NewMarkdownContent("test content")

	doc, err := NewDocumentFromName(title, filters, content, rawContent)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedID := DocumentID("test-document-title")
	if doc.ID != expectedID {
		t.Errorf("Expected ID %s, got %s", expectedID, doc.ID)
	}
	if doc.Title != title {
		t.Errorf("Expected title %s, got %s", title, doc.Title)
	}
}

func TestDocument_EntityType(t *testing.T) {
	tests := []struct {
		name     string
		filters  DocumentFilters
		expected string
	}{
		{
			name:     "collection filter present",
			filters:  DocumentFilters{"collection": "spells"},
			expected: "spells",
		},
		{
			name:     "no collection filter",
			filters:  DocumentFilters{},
			expected: "document",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &Document{Filters: tt.filters}
			result := doc.EntityType()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDocument_GetCollection(t *testing.T) {
	tests := []struct {
		name     string
		filters  DocumentFilters
		expected string
	}{
		{
			name:     "string collection",
			filters:  DocumentFilters{"collection": "weapons"},
			expected: "weapons",
		},
		{
			name:     "no collection",
			filters:  DocumentFilters{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &Document{Filters: tt.filters}
			result := doc.GetCollection()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
