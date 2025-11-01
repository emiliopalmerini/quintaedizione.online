package parsers

import (
	"testing"
)

func TestIsValidContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType ContentType
		expected    bool
	}{
		{"valid - incantesimi", ContentTypeIncantesimi, true},
		{"valid - mostri", ContentTypeMostri, true},
		{"valid - classi", ContentTypeClassi, true},
		{"valid - armi", ContentTypeArmi, true},
		{"valid - armature", ContentTypeArmature, true},
		{"valid - equipaggiamenti", ContentTypeEquipaggiamenti, true},
		{"valid - oggetti_magici", ContentTypeOggettiMagici, true},
		{"valid - talenti", ContentTypeTalenti, true},
		{"valid - animali", ContentTypeAnimali, true},
		{"valid - regole", ContentTypeRegole, true},
		{"valid - backgrounds", ContentTypeBackgrounds, true},
		{"valid - strumenti", ContentTypeStrumenti, true},
		{"valid - servizi", ContentTypeServizi, true},
		{"valid - cavalcature_veicoli", ContentTypeCavalcatureVeicoli, true},
		{"valid - documenti", ContentTypeDocuments, true},
		{"invalid - empty", ContentType(""), false},
		{"invalid - unknown", ContentType("unknown"), false},
		{"invalid - invalid_type", ContentType("invalid_type"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidContentType(tt.contentType)
			if result != tt.expected {
				t.Errorf("IsValidContentType(%s) = %t, expected %t", tt.contentType, result, tt.expected)
			}
		})
	}
}

func TestGetAllContentTypes(t *testing.T) {
	contentTypes := GetAllContentTypes()

	expectedCount := 15 // Based on the validContentTypes map
	if len(contentTypes) != expectedCount {
		t.Errorf("Expected %d content types, got %d", expectedCount, len(contentTypes))
	}

	// Verify all types are valid
	for _, ct := range contentTypes {
		if !IsValidContentType(ct) {
			t.Errorf("Content type %s should be valid", ct)
		}
	}

	// Verify no duplicates (convert to map)
	seen := make(map[ContentType]bool)
	for _, ct := range contentTypes {
		if seen[ct] {
			t.Errorf("Duplicate content type: %s", ct)
		}
		seen[ct] = true
	}
}

func TestGetContentTypeFromCollection(t *testing.T) {
	tests := []struct {
		name         string
		collection   string
		expected     ContentType
		expectError  bool
	}{
		{"incantesimi", "incantesimi", ContentTypeIncantesimi, false},
		{"mostri", "mostri", ContentTypeMostri, false},
		{"classi", "classi", ContentTypeClassi, false},
		{"armi", "armi", ContentTypeArmi, false},
		{"armature", "armature", ContentTypeArmature, false},
		{"equipaggiamenti", "equipaggiamenti", ContentTypeEquipaggiamenti, false},
		{"oggetti_magici", "oggetti_magici", ContentTypeOggettiMagici, false},
		{"talenti", "talenti", ContentTypeTalenti, false},
		{"animali", "animali", ContentTypeAnimali, false},
		{"regole", "regole", ContentTypeRegole, false},
		{"backgrounds", "backgrounds", ContentTypeBackgrounds, false},
		{"strumenti", "strumenti", ContentTypeStrumenti, false},
		{"servizi", "servizi", ContentTypeServizi, false},
		{"cavalcature_veicoli", "cavalcature_veicoli", ContentTypeCavalcatureVeicoli, false},
		{"documenti", "documenti", ContentTypeDocuments, false},
		{"unknown collection", "unknown", "", true},
		{"empty collection", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetContentTypeFromCollection(tt.collection)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for collection %s, got none", tt.collection)
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error for collection %s, got %v", tt.collection, err)
				return
			}

			if result != tt.expected {
				t.Errorf("GetContentTypeFromCollection(%s) = %s, expected %s", tt.collection, result, tt.expected)
			}
		})
	}
}

func TestGetCollectionFromContentType(t *testing.T) {
	tests := []struct {
		name         string
		contentType  ContentType
		expected     string
		expectError  bool
	}{
		{"incantesimi", ContentTypeIncantesimi, "incantesimi", false},
		{"mostri", ContentTypeMostri, "mostri", false},
		{"classi", ContentTypeClassi, "classi", false},
		{"armi", ContentTypeArmi, "armi", false},
		{"armature", ContentTypeArmature, "armature", false},
		{"equipaggiamenti", ContentTypeEquipaggiamenti, "equipaggiamenti", false},
		{"oggetti_magici", ContentTypeOggettiMagici, "oggetti_magici", false},
		{"talenti", ContentTypeTalenti, "talenti", false},
		{"animali", ContentTypeAnimali, "animali", false},
		{"regole", ContentTypeRegole, "regole", false},
		{"backgrounds", ContentTypeBackgrounds, "backgrounds", false},
		{"strumenti", ContentTypeStrumenti, "strumenti", false},
		{"servizi", ContentTypeServizi, "servizi", false},
		{"cavalcature_veicoli", ContentTypeCavalcatureVeicoli, "cavalcature_veicoli", false},
		{"documenti", ContentTypeDocuments, "documenti", false},
		{"invalid content type", ContentType("invalid"), "", true},
		{"empty content type", ContentType(""), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetCollectionFromContentType(tt.contentType)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for content type %s, got none", tt.contentType)
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error for content type %s, got %v", tt.contentType, err)
				return
			}

			if result != tt.expected {
				t.Errorf("GetCollectionFromContentType(%s) = %s, expected %s", tt.contentType, result, tt.expected)
			}
		})
	}
}
