package domain

import (
	"strings"
	"testing"
)

func TestNewDocumentID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected DocumentID
		hasError bool
	}{
		{
			name:     "simple title",
			input:    "Fireball",
			expected: "fireball",
			hasError: false,
		},
		{
			name:     "title with spaces",
			input:    "Magic Missile",
			expected: "magic-missile",
			hasError: false,
		},
		{
			name:     "title with special characters",
			input:    "Fire Ball (Level 3)",
			expected: "fire-ball-level-3",
			hasError: false,
		},
		{
			name:     "title with unicode",
			input:    "Mágic Míssilé",
			expected: "magic-missile",
			hasError: false,
		},
		{
			name:     "empty title",
			input:    "",
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewDocumentID(tt.input)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error but got %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}

			if !isValidSlug(string(result)) {
				t.Errorf("Generated ID %s is not a valid slug", result)
			}
		})
	}
}

func TestDocumentID_String(t *testing.T) {
	id := DocumentID("test-slug")
	expected := "test-slug"

	if id.String() != expected {
		t.Errorf("Expected %s, got %s", expected, id.String())
	}
}

func isValidSlug(s string) bool {
	if s == "" {
		return false
	}

	if s != strings.ToLower(s) {
		return false
	}

	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}

	if strings.HasPrefix(s, "-") || strings.HasSuffix(s, "-") {
		return false
	}

	if strings.Contains(s, "--") {
		return false
	}

	return true
}
