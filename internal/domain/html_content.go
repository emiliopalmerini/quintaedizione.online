package domain

// HTMLContent represents rendered HTML content
type HTMLContent string

// NewHTMLContent creates HTMLContent from a string
func NewHTMLContent(html string) HTMLContent {
	return HTMLContent(html)
}

// String returns the HTML string
func (h HTMLContent) String() string {
	return string(h)
}

// IsEmpty checks if content is empty
func (h HTMLContent) IsEmpty() bool {
	return len(h) == 0
}
