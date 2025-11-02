package domain

// MarkdownContent represents raw markdown content
type MarkdownContent string

// NewMarkdownContent creates MarkdownContent from a string
func NewMarkdownContent(markdown string) MarkdownContent {
	return MarkdownContent(markdown)
}

// String returns the markdown string
func (m MarkdownContent) String() string {
	return string(m)
}

// IsEmpty checks if content is empty
func (m MarkdownContent) IsEmpty() bool {
	return len(m) == 0
}
