package parsers

import "github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"

// DocumentParsingStrategy defines the interface for Document-based parsing strategies
// This is the new interface that outputs unified Document entities with HTML content
type DocumentParsingStrategy interface {
	// ParseDocument parses content and returns Document entities
	ParseDocument(content []string, context *ParsingContext) ([]*domain.Document, error)

	// ContentType returns the content type this strategy handles
	ContentType() ContentType

	// Name returns the strategy name
	Name() string

	// Description returns a description of what this strategy parses
	Description() string

	// Validate validates the content structure
	Validate(content []string) error
}

// BaseDocumentParser provides common functionality for Document-based parsers
type BaseDocumentParser struct {
	renderer *MarkdownRenderer
}

// NewBaseDocumentParser creates a new base document parser
func NewBaseDocumentParser() *BaseDocumentParser {
	return &BaseDocumentParser{
		renderer: NewMarkdownRenderer(),
	}
}

// RenderMarkdown converts markdown to HTML
func (p *BaseDocumentParser) RenderMarkdown(markdown string) domain.HTMLContent {
	html := p.renderer.Render(markdown)
	return domain.NewHTMLContent(html)
}

// RenderMarkdownLines converts markdown lines to HTML
func (p *BaseDocumentParser) RenderMarkdownLines(lines []string) domain.HTMLContent {
	html := p.renderer.RenderLines(lines)
	return domain.NewHTMLContent(html)
}

// CreateDocument is a helper to create a Document with common fields
func (p *BaseDocumentParser) CreateDocument(
	title string,
	collection string,
	sourceFile string,
	locale string,
	markdownContent string,
	additionalFilters map[string]any,
) (*domain.Document, error) {
	// Create document ID from title
	id, err := domain.NewDocumentID(title)
	if err != nil {
		return nil, err
	}

	// Build filters
	filters := domain.NewDocumentFilters()
	filters.Set("collection", collection)
	filters.Set("source_file", sourceFile)
	filters.Set("locale", locale)

	// Add additional filters
	for key, value := range additionalFilters {
		filters.Set(key, value)
	}

	// Store both raw markdown and rendered HTML
	rawContent := domain.NewMarkdownContent(markdownContent)
	htmlContent := p.RenderMarkdown(markdownContent)

	return domain.NewDocument(id, title, filters, htmlContent, rawContent), nil
}
