package domain

type MarkdownContent string

func NewMarkdownContent(markdown string) MarkdownContent {
	return MarkdownContent(markdown)
}

func (m MarkdownContent) String() string {
	return string(m)
}

func (m MarkdownContent) IsEmpty() bool {
	return len(m) == 0
}
