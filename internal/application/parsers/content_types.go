package parsers

// ContentType definitions and utilities

// ContentType represents the type of D&D content being parsed
type ContentType string

// Content type constants for Italian D&D 5e SRD
const (
	ContentTypeDocuments   ContentType = "documenti"
	ContentTypeSpells      ContentType = "incantesimi"
	ContentTypeMonsters    ContentType = "mostri"
	ContentTypeClasses     ContentType = "classi"
	ContentTypeBackgrounds ContentType = "backgrounds"
	ContentTypeWeapons     ContentType = "armi"
	ContentTypeArmor       ContentType = "armature"
	ContentTypeTools       ContentType = "strumenti"
	ContentTypeServices    ContentType = "servizi"
	ContentTypeGear        ContentType = "equipaggiamento"
	ContentTypeMagicItems  ContentType = "oggetti_magici"
	ContentTypeFeats       ContentType = "talenti"
	ContentTypeAnimals     ContentType = "animali"
)

// validContentTypes provides O(1) content type validation
var validContentTypes = map[ContentType]bool{
	ContentTypeDocuments:   true,
	ContentTypeSpells:      true,
	ContentTypeMonsters:    true,
	ContentTypeClasses:     true,
	ContentTypeBackgrounds: true,
	ContentTypeWeapons:     true,
	ContentTypeArmor:       true,
	ContentTypeTools:       true,
	ContentTypeServices:    true,
	ContentTypeGear:        true,
	ContentTypeMagicItems:  true,
	ContentTypeFeats:       true,
	ContentTypeAnimals:     true,
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
		return ContentTypeSpells, nil
	case "mostri":
		return ContentTypeMonsters, nil
	case "classi":
		return ContentTypeClasses, nil
	case "backgrounds":
		return ContentTypeBackgrounds, nil
	case "armi":
		return ContentTypeWeapons, nil
	case "armature":
		return ContentTypeArmor, nil
	case "strumenti":
		return ContentTypeTools, nil
	case "servizi":
		return ContentTypeServices, nil
	case "equipaggiamento":
		return ContentTypeGear, nil
	case "oggetti_magici":
		return ContentTypeMagicItems, nil
	case "talenti":
		return ContentTypeFeats, nil
	case "animali":
		return ContentTypeAnimals, nil
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
