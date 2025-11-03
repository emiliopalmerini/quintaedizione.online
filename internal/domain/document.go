package domain

// Document represents a unified D&D 5e SRD content entity.
// This is the core domain model that replaces all specific entity types
// (Mostro, Classe, Incantesimo, etc.) with a flexible document structure.
type Document struct {
	ID         DocumentID      `json:"id"          bson:"_id"`
	Title      string          `json:"title"       bson:"title"`
	Filters    DocumentFilters `json:"filters"     bson:"filters"`
	Content    HTMLContent     `json:"content"     bson:"content"`
	RawContent MarkdownContent `json:"raw_content" bson:"raw_content"`
}

// NewDocument creates a new Document
func NewDocument(id DocumentID, title string, filters DocumentFilters, content HTMLContent, rawContent MarkdownContent) *Document {
	return &Document{
		ID:         id,
		Title:      title,
		Filters:    filters,
		Content:    content,
		RawContent: rawContent,
	}
}

// NewDocumentFromName creates a new Document with auto-generated ID from title
func NewDocumentFromName(title string, filters DocumentFilters, content HTMLContent, rawContent MarkdownContent) (*Document, error) {
	id, err := NewDocumentID(title)
	if err != nil {
		return nil, err
	}
	return NewDocument(id, title, filters, content, rawContent), nil
}

// EntityType implements ParsedEntity interface
// Returns the collection name from filters, or "document" as fallback
func (d *Document) EntityType() string {
	if collection, ok := d.Filters["collection"].(string); ok {
		return collection
	}
	return "document"
}

// GetCollection returns the collection name from filters
func (d *Document) GetCollection() string {
	return d.Filters.GetString("collection")
}

// GetSourceFile returns the source file from filters
func (d *Document) GetSourceFile() string {
	return d.Filters.GetString("source_file")
}

// GetLocale returns the locale from filters
func (d *Document) GetLocale() string {
	return d.Filters.GetString("locale")
}
