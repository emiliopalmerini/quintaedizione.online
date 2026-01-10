package filters

import (
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/collections"
)

func TestBuildSearchFilter(t *testing.T) {
	builder := NewMongoFilterBuilder()

	tests := []struct {
		name           string
		searchTerm     string
		shouldHaveText bool
		expectedSearch string
	}{
		{
			name:           "empty search term",
			searchTerm:     "",
			shouldHaveText: false,
		},
		{
			name:           "single word search",
			searchTerm:     "fuoco",
			shouldHaveText: true,
			expectedSearch: "fuoco",
		},
		{
			name:           "multi-word search",
			searchTerm:     "palla di fuoco",
			shouldHaveText: true,
			expectedSearch: "palla di fuoco",
		},
		{
			name:           "quoted phrase search",
			searchTerm:     `"palla di fuoco"`,
			shouldHaveText: true,
			expectedSearch: `"palla di fuoco"`,
		},
		{
			name:           "whitespace handling",
			searchTerm:     "fuoco",
			shouldHaveText: true,
			expectedSearch: "fuoco",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.BuildSearchFilter(collections.Incantesimi, tt.searchTerm)

			textOp, ok := result["$text"]
			if !ok && tt.shouldHaveText {
				t.Errorf("$text operator not found in result")
				return
			}

			if !tt.shouldHaveText && len(result) > 0 {
				t.Errorf("expected empty filter, got %v", result)
				return
			}

			if tt.shouldHaveText {
				textSearch, ok := textOp.(map[string]any)
				if !ok {
					t.Errorf("$text value is not a map[string]any: %T", textOp)
					return
				}

				search, ok := textSearch["$search"]
				if !ok {
					t.Errorf("$search field not found in $text operator")
					return
				}

				if search != tt.expectedSearch {
					t.Errorf("search term mismatch: got %v, want %v", search, tt.expectedSearch)
				}
			}
		})
	}
}

func TestBuildSearchFilter_Collections(t *testing.T) {
	builder := NewMongoFilterBuilder()
	searchTerm := "test"

	collectionList := []collections.CollectionName{
		collections.Incantesimi,
		collections.Mostri,
		collections.Animali,
		collections.Armi,
		collections.Armature,
		collections.OggettiMagici,
		collections.Classi,
		collections.Backgrounds,
		collections.Talenti,
	}

	for _, col := range collectionList {
		t.Run(col.String(), func(t *testing.T) {
			result := builder.BuildSearchFilter(col, searchTerm)

			textOp, ok := result["$text"]
			if !ok {
				t.Errorf("collection %s: $text operator not found", col)
				return
			}

			textSearch, ok := textOp.(map[string]any)
			if !ok {
				t.Errorf("collection %s: $text is not a map[string]any: %T", col, textOp)
				return
			}

			search, ok := textSearch["$search"]
			if !ok {
				t.Errorf("collection %s: $search not found", col)
				return
			}

			if search != searchTerm {
				t.Errorf("collection %s: search term mismatch: got %v, want %v",
					col, search, searchTerm)
			}
		})
	}
}

func TestBuildSearchFilter_SpecialCharacters(t *testing.T) {
	builder := NewMongoFilterBuilder()

	tests := []struct {
		name       string
		searchTerm string
	}{
		{
			name:       "accented characters",
			searchTerm: "miracolo",
		},
		{
			name:       "hyphenated words",
			searchTerm: "non-morto",
		},
		{
			name:       "negation operator",
			searchTerm: "-invisibile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.BuildSearchFilter(collections.Incantesimi, tt.searchTerm)

			if _, ok := result["$text"]; !ok {
				t.Errorf("$text operator not found")
			}

			textSearch := result["$text"].(map[string]any)
			if search, ok := textSearch["$search"]; ok && search != tt.searchTerm {
				t.Errorf("search term mismatch: got %v, want %v", search, tt.searchTerm)
			}
		})
	}
}

// TestSearchTermValidation tests the input validation function against injection attacks
func TestSearchTermValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid inputs
		{
			name:     "simple word",
			input:    "hello",
			expected: true,
		},
		{
			name:     "multiple words",
			input:    "hello world",
			expected: true,
		},
		{
			name:     "quoted phrase",
			input:    `"hello world"`,
			expected: true,
		},
		{
			name:     "hyphenated words",
			input:    "non-morto",
			expected: true,
		},
		{
			name:     "accented characters",
			input:    "miracolo",
			expected: true,
		},
		{
			name:     "negation operator",
			input:    "-invisibile",
			expected: true,
		},
		{
			name:     "and operator in search",
			input:    "hello AND world",
			expected: true,
		},
		{
			name:     "or operator in search",
			input:    "hello OR world",
			expected: true,
		},
		// Invalid inputs - MongoDB operators
		{
			name:     "$where operator",
			input:    `$where: "test"`,
			expected: false,
		},
		{
			name:     "$regex operator",
			input:    `$regex: "pattern"`,
			expected: false,
		},
		{
			name:     "$ne operator",
			input:    `test $ne value`,
			expected: false,
		},
		{
			name:     "$gt operator",
			input:    `value $gt 10`,
			expected: false,
		},
		{
			name:     "$lt operator",
			input:    `value $lt 10`,
			expected: false,
		},
		{
			name:     "$in operator",
			input:    `value $in [1,2,3]`,
			expected: false,
		},
		{
			name:     "$or operator",
			input:    `$or: [{"a": 1}]`,
			expected: false,
		},
		{
			name:     "$and operator",
			input:    `$and: [{"a": 1}]`,
			expected: false,
		},
		// Invalid inputs - dangerous patterns
		{
			name:     "shell pipe",
			input:    `test | whoami`,
			expected: false,
		},
		{
			name:     "shell and operator",
			input:    `test && whoami`,
			expected: false,
		},
		{
			name:     "shell or operator",
			input:    `test || whoami`,
			expected: false,
		},
		{
			name:     "semicolon",
			input:    `test; command`,
			expected: false,
		},
		{
			name:     "less than greater than",
			input:    `test<>string`,
			expected: false,
		},
		{
			name:     "dollar sign in middle",
			input:    `test$value`,
			expected: false,
		},
		{
			name:     "command substitution",
			input:    `test $(whoami)`,
			expected: false,
		},
		{
			name:     "backticks",
			input:    "`whoami`",
			expected: false,
		},
		// Length validation
		{
			name:     "length limit exceeded",
			input:    string(make([]byte, 501)),
			expected: false,
		},
		{
			name:     "length at limit",
			input:    string(make([]byte, 500)),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidSearchTerm(tt.input)
			if result != tt.expected {
				t.Errorf("isValidSearchTerm(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestSearchTermSanitization tests the sanitization function
func TestSearchTermSanitization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple word",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "trim whitespace",
			input:    "  hello  ",
			expected: "hello",
		},
		{
			name:     "remove semicolon",
			input:    "hello;world",
			expected: "helloworld",
		},
		{
			name:     "remove pipe",
			input:    "hello|world",
			expected: "helloworld",
		},
		{
			name:     "remove dollar sign",
			input:    "hello$world",
			expected: "helloworld",
		},
		{
			name:     "remove angle brackets",
			input:    "hello<world>test",
			expected: "helloworldtest",
		},
		{
			name:     "remove parentheses",
			input:    "hello(world)",
			expected: "helloworld",
		},
		{
			name:     "remove curly braces",
			input:    "hello{world}",
			expected: "helloworld",
		},
		{
			name:     "remove square brackets",
			input:    "hello[world]",
			expected: "helloworld",
		},
		{
			name:     "remove backslash",
			input:    `hello\world`,
			expected: "helloworld",
		},
		{
			name:     "keep quotes",
			input:    `"hello world"`,
			expected: `"hello world"`,
		},
		{
			name:     "keep hyphens",
			input:    "hello-world",
			expected: "hello-world",
		},
		{
			name:     "mixed dangerous and safe",
			input:    `hello"world-test;danger`,
			expected: `hello"world-testdanger`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeSearchTerm(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeSearchTerm(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestBuildSearchFilter_InjectionPrevention tests that injection attempts are blocked
func TestBuildSearchFilter_InjectionPrevention(t *testing.T) {
	builder := NewMongoFilterBuilder()

	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{
			name:      "valid search",
			input:     "fireball",
			shouldErr: false,
		},
		{
			name:      "$where injection",
			input:     `$where: "return true"`,
			shouldErr: true,
		},

		{
			name:      "logical operator injection with pipe",
			input:     `test || echo pwned`,
			shouldErr: true,
		},
		{
			name:      "$or operator injection",
			input:     `spell $or magic`,
			shouldErr: true,
		},
		{
			name:      "quoted phrase is safe",
			input:     `"fireball spell"`,
			shouldErr: false,
		},
		{
			name:      "valid AND search",
			input:     "fireball AND spell",
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.BuildSearchFilter(collections.Incantesimi, tt.input)

			hasText := len(result) > 0 && result["$text"] != nil

			if !tt.shouldErr && !hasText {
				t.Errorf("expected valid search filter, got empty result")
			}

			if tt.shouldErr && hasText {
				t.Errorf("injection attempt was not blocked: %q", tt.input)
			}
		})
	}
}
