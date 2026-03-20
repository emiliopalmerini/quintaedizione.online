package domain

type DocumentID string

func NewDocumentID(name string) (DocumentID, error) {
	slug, err := NewSlug(name)
	if err != nil {
		return "", err
	}
	return DocumentID(slug), nil
}

func (d DocumentID) String() string {
	return string(d)
}

func (d DocumentID) IsEmpty() bool {
	return len(d) == 0
}
