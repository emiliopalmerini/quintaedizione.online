package domain

// ParsedEntity represents any entity that can be parsed from SRD content.
// This interface provides a type-safe way for parsers to return domain objects
// while maintaining the flexibility to handle different entity types.
type ParsedEntity interface {
	// EntityType returns the entity type identifier for validation and routing.
	// Examples: "mostro", "classe", "incantesimo", etc.
	EntityType() string
}