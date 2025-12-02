package parsers

import "github.com/emiliopalmerini/quintaedizione.online/internal/domain"

type DocumentParsingStrategy interface {
	ParseDocument(content []string, context *ParsingContext) ([]*domain.Document, error)

	ContentType() ContentType

	Name() string

	Description() string

	Validate(content []string) error
}

type BaseDocumentParser struct {
	renderer *MarkdownRenderer
}

func NewBaseDocumentParser() *BaseDocumentParser {
	return &BaseDocumentParser{
		renderer: NewMarkdownRenderer(),
	}
}

func (p *BaseDocumentParser) RenderMarkdown(markdown string) domain.HTMLContent {
	html := p.renderer.Render(markdown)
	return domain.NewHTMLContent(html)
}

func (p *BaseDocumentParser) RenderMarkdownLines(lines []string) domain.HTMLContent {
	html := p.renderer.RenderLines(lines)
	return domain.NewHTMLContent(html)
}

func (p *BaseDocumentParser) CreateDocument(
	title string,
	collection string,
	sourceFile string,
	locale string,
	markdownContent string,
	additionalFilters map[string]any,
) (*domain.Document, error) {

	id, err := domain.NewDocumentID(title)
	if err != nil {
		return nil, err
	}

	filters := domain.NewDocumentFilters()
	filters.Set("collection", collection)
	filters.Set("source_file", sourceFile)
	filters.Set("locale", locale)

	for key, value := range additionalFilters {
		filters.Set(key, value)
	}

	rawContent := domain.NewMarkdownContent(markdownContent)
	htmlContent := p.RenderMarkdown(markdownContent)

	return domain.NewDocument(id, title, filters, htmlContent, rawContent), nil
}
