package domain

type HTMLContent string

func NewHTMLContent(html string) HTMLContent {
	return HTMLContent(html)
}

func (h HTMLContent) String() string {
	return string(h)
}

func (h HTMLContent) IsEmpty() bool {
	return len(h) == 0
}
