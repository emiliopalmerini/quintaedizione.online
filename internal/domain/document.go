package domain

type Document struct {
	ID         DocumentID      `json:"id"          bson:"_id"`
	Title      string          `json:"title"       bson:"title"`
	Filters    DocumentFilters `json:"filters"     bson:"filters"`
	Content    HTMLContent     `json:"content"     bson:"content"`
	RawContent MarkdownContent `json:"raw_content" bson:"raw_content"`
}

func NewDocument(id DocumentID, title string, filters DocumentFilters, content HTMLContent, rawContent MarkdownContent) *Document {
	return &Document{
		ID:         id,
		Title:      title,
		Filters:    filters,
		Content:    content,
		RawContent: rawContent,
	}
}

func NewDocumentFromName(title string, filters DocumentFilters, content HTMLContent, rawContent MarkdownContent) (*Document, error) {
	id, err := NewDocumentID(title)
	if err != nil {
		return nil, err
	}
	return NewDocument(id, title, filters, content, rawContent), nil
}

func (d *Document) EntityType() string {
	if collection, ok := d.Filters["collection"].(string); ok {
		return collection
	}
	return "document"
}

func (d *Document) GetCollection() string {
	return d.Filters.GetString("collection")
}

func (d *Document) GetSourceFile() string {
	return d.Filters.GetString("source_file")
}

func (d *Document) GetLocale() string {
	return d.Filters.GetString("locale")
}
