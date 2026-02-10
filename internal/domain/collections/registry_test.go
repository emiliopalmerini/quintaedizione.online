package collections

import (
	"testing"
)

func TestGetAllCollectionNames(t *testing.T) {
	names := GetAllCollectionNames()

	// Should return all 11 collections
	if len(names) != 11 {
		t.Errorf("Expected 11 collections, got %d", len(names))
	}

	// All names should be strings
	for _, name := range names {
		if name == "" {
			t.Error("Found empty collection name")
		}
	}

	// Verify all expected collections are present
	expectedCollections := map[string]bool{
		"classi":          false,
		"backgrounds":     false,
		"incantesimi":     false,
		"talenti":         false,
		"equipaggiamenti": false,
		"servizi":         false,
		"regole":          false,
		"oggetti_magici":  false,
		"mostri":          false,
		"glossario":       false,
		"specie":          false,
	}

	for _, name := range names {
		if _, exists := expectedCollections[name]; exists {
			expectedCollections[name] = true
		} else {
			t.Errorf("Unexpected collection name: %s", name)
		}
	}

	for name, found := range expectedCollections {
		if !found {
			t.Errorf("Missing expected collection: %s", name)
		}
	}
}

func TestFromString_ValidCollection(t *testing.T) {
	tests := []struct {
		input    string
		expected CollectionName
	}{
		{"incantesimi", Incantesimi},
		{"mostri", Mostri},
		{"classi", Classi},
		{"talenti", Talenti},
		{"backgrounds", Backgrounds},
		{"equipaggiamenti", Equipaggiamenti},
		{"servizi", Servizi},
		{"regole", Regole},
		{"oggetti_magici", OggettiMagici},
		{"glossario", Glossario},
		{"specie", Specie},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, valid := FromString(tt.input)
			if !valid {
				t.Errorf("Expected %s to be valid", tt.input)
			}
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFromString_InvalidCollection(t *testing.T) {
	tests := []string{
		"invalid",
		"",
		"Incantesimi",  // Wrong case
		"incantesimi ", // Trailing space
		" incantesimi", // Leading space
		"non_existent",
		"armi",                // Removed collection
		"armature",            // Removed collection
		"animali",             // Removed collection
		"strumenti",           // Removed collection
		"cavalcature_veicoli", // Removed collection
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, valid := FromString(input)
			if valid {
				t.Errorf("Expected %s to be invalid", input)
			}
		})
	}
}

func TestFromString_RoundTrip(t *testing.T) {
	// Test that converting to string and back gives the same value
	for _, collName := range GetAllCollections() {
		str := collName.String()
		result, valid := FromString(str)
		if !valid {
			t.Errorf("Round trip failed for %s: not valid", str)
		}
		if result != collName {
			t.Errorf("Round trip failed for %s: got %s", collName, result)
		}
	}
}
