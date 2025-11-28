package filters

import (
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/filters"
	"go.mongodb.org/mongo-driver/bson"
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
			result := builder.BuildSearchFilter(filters.IncantesimiCollection, tt.searchTerm)

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
				textSearch, ok := textOp.(bson.M)
				if !ok {
					t.Errorf("$text value is not a bson.M: %T", textOp)
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

	collections := []filters.CollectionType{
		filters.IncantesimiCollection,
		filters.MostriCollection,
		filters.AnimaliCollection,
		filters.ArmiCollection,
		filters.ArmatureCollection,
		filters.OggettiMagiciCollection,
		filters.ClassiCollection,
		filters.BackgroundsCollection,
		filters.TalentiCollection,
	}

	for _, col := range collections {
		t.Run(col.String(), func(t *testing.T) {
			result := builder.BuildSearchFilter(col, searchTerm)

			// All collections should use the same $text operator
			textOp, ok := result["$text"]
			if !ok {
				t.Errorf("collection %s: $text operator not found", col)
				return
			}

			textSearch, ok := textOp.(bson.M)
			if !ok {
				t.Errorf("collection %s: $text is not a bson.M: %T", col, textOp)
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
			result := builder.BuildSearchFilter(filters.IncantesimiCollection, tt.searchTerm)

			if _, ok := result["$text"]; !ok {
				t.Errorf("$text operator not found")
			}

			textSearch := result["$text"].(bson.M)
			if search, ok := textSearch["$search"]; ok && search != tt.searchTerm {
				t.Errorf("search term mismatch: got %v, want %v", search, tt.searchTerm)
			}
		})
	}
}
