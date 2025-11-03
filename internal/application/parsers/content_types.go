package parsers

// ContentType definitions and utilities

// ContentType represents the type of D&D content being parsed
type ContentType string

// Content type constants for Italian D&D 5e SRD
const (
	ContentTypeDocuments          ContentType = "documenti"
	ContentTypeIncantesimi        ContentType = "incantesimi"
	ContentTypeMostri             ContentType = "mostri"
	ContentTypeClassi             ContentType = "classi"
	ContentTypeBackgrounds        ContentType = "backgrounds"
	ContentTypeArmi               ContentType = "armi"
	ContentTypeArmature           ContentType = "armature"
	ContentTypeStrumenti          ContentType = "strumenti"
	ContentTypeServizi            ContentType = "servizi"
	ContentTypeEquipaggiamenti    ContentType = "equipaggiamenti"
	ContentTypeOggettiMagici      ContentType = "oggetti_magici"
	ContentTypeTalenti            ContentType = "talenti"
	ContentTypeAnimali            ContentType = "animali"
	ContentTypeRegole             ContentType = "regole"
	ContentTypeCavalcatureVeicoli ContentType = "cavalcature_veicoli"
)

// validContentTypes provides O(1) content type validation
var validContentTypes = map[ContentType]bool{
	ContentTypeDocuments:          true,
	ContentTypeIncantesimi:        true,
	ContentTypeMostri:             true,
	ContentTypeClassi:             true,
	ContentTypeBackgrounds:        true,
	ContentTypeArmi:               true,
	ContentTypeArmature:           true,
	ContentTypeStrumenti:          true,
	ContentTypeServizi:            true,
	ContentTypeEquipaggiamenti:    true,
	ContentTypeOggettiMagici:      true,
	ContentTypeTalenti:            true,
	ContentTypeAnimali:            true,
	ContentTypeRegole:             true,
	ContentTypeCavalcatureVeicoli: true,
}

// IsValidContentType checks if a content type is valid
func IsValidContentType(contentType ContentType) bool {
	return validContentTypes[contentType]
}

// GetAllContentTypes returns all valid content types
func GetAllContentTypes() []ContentType {
	types := make([]ContentType, 0, len(validContentTypes))
	for contentType := range validContentTypes {
		types = append(types, contentType)
	}
	return types
}

// GetContentTypeFromCollection maps MongoDB collection names to content types
func GetContentTypeFromCollection(collection string) (ContentType, error) {
	switch collection {
	case "documenti":
		return ContentTypeDocuments, nil
	case "incantesimi":
		return ContentTypeIncantesimi, nil
	case "mostri":
		return ContentTypeMostri, nil
	case "classi":
		return ContentTypeClassi, nil
	case "backgrounds":
		return ContentTypeBackgrounds, nil
	case "armi":
		return ContentTypeArmi, nil
	case "armature":
		return ContentTypeArmature, nil
	case "strumenti":
		return ContentTypeStrumenti, nil
	case "servizi":
		return ContentTypeServizi, nil
	case "equipaggiamenti":
		return ContentTypeEquipaggiamenti, nil
	case "oggetti_magici":
		return ContentTypeOggettiMagici, nil
	case "talenti":
		return ContentTypeTalenti, nil
	case "animali":
		return ContentTypeAnimali, nil
	case "regole":
		return ContentTypeRegole, nil
	case "cavalcature_veicoli":
		return ContentTypeCavalcatureVeicoli, nil
	default:
		return "", ErrInvalidContentType
	}
}

// GetCollectionFromContentType maps content types to MongoDB collection names
func GetCollectionFromContentType(contentType ContentType) (string, error) {
	if !IsValidContentType(contentType) {
		return "", ErrInvalidContentType
	}

	return string(contentType), nil
}
