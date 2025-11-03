package domain

// DocumentID is a unique identifier for documents (slug-based)
type DocumentID string

// NewDocumentID creates a DocumentID from a name using slug conversion
func NewDocumentID(name string) (DocumentID, error) {
	slug, err := NewSlug(name)
	if err != nil {
		return "", err
	}
	return DocumentID(slug), nil
}

// String returns the string representation
func (d DocumentID) String() string {
	return string(d)
}

// IsEmpty checks if the ID is empty
func (d DocumentID) IsEmpty() bool {
	return len(d) == 0
}
