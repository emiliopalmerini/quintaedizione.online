package domain

// Document represents an SRD content item with typed core fields
// and a flexible Fields map for collection-specific data.
type Document struct {
	ID         DocumentID      `json:"id"`
	Title      string          `json:"title"`
	Source     string          `json:"source"`      // source short name (e.g., "5.5e")
	Content    HTMLContent     `json:"content"`     // rendered HTML
	RawContent MarkdownContent `json:"raw_content"` // markdown for copy-to-clipboard
	Fields     map[string]any  `json:"fields"`      // collection-specific fields (livello, scuola, etc.)
}

// DocumentFromMap converts a raw map[string]any (from the Store) to a typed Document.
// Core fields are extracted into typed struct fields; everything else goes into Fields.
func DocumentFromMap(m map[string]any) *Document {
	id, _ := m["_id"].(string)
	title, _ := m["title"].(string)
	content, _ := m["content"].(string)
	rawContent, _ := m["raw_content"].(string)
	source, _ := m["_source_short"].(string)

	fields := make(map[string]any, len(m))
	for k, v := range m {
		switch k {
		case "_id", "title", "content", "raw_content", "_source_short", "_source":
			// skip — already in typed fields
		default:
			fields[k] = v
		}
	}

	return &Document{
		ID:         DocumentID(id),
		Title:      title,
		Source:     source,
		Content:    HTMLContent(content),
		RawContent: MarkdownContent(rawContent),
		Fields:     fields,
	}
}

// GetField returns a collection-specific field value, or nil if not present.
func (d *Document) GetField(key string) any {
	if d.Fields == nil {
		return nil
	}
	return d.Fields[key]
}

// GetFieldString returns a collection-specific field as a string.
func (d *Document) GetFieldString(key string) string {
	if v, ok := d.Fields[key].(string); ok {
		return v
	}
	return ""
}
