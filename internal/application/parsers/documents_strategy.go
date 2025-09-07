package parsers

import (
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/google/uuid"
)

// DocumentsStrategy implements the Strategy pattern for parsing documents
type DocumentsStrategy struct {
	*BaseParser
}

// NewDocumentsStrategy creates a new documents parsing strategy
func NewDocumentsStrategy() ParsingStrategy {
	return &DocumentsStrategy{
		BaseParser: NewBaseParser(
			ContentTypeDocuments,
			"Documents Parser",
			"Parses D&D 5e documents from Italian SRD markdown content",
		),
	}
}

// Parse processes document content and returns domain Documento objects
func (d *DocumentsStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
	if err := d.Validate(content); err != nil {
		return nil, err
	}

	cleanContent := d.CleanContent(content)
	if len(cleanContent) == 0 {
		return []domain.ParsedEntity{}, nil
	}

	title := d.extractTitle(cleanContent)
	body := d.extractBody(cleanContent)

	documento := domain.NewDocumento(
		uuid.New(),
		1, // pagina number - could be extracted from filename
		title,
		body,
	)

	return []domain.ParsedEntity{documento}, nil
}

func (d *DocumentsStrategy) extractTitle(content []string) string {
	for _, line := range content {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimSpace(trimmed[2:])
		}
	}

	return "Documento Senza Titolo"
}

func (d *DocumentsStrategy) extractBody(content []string) string {
	var bodyLines []string
	skipFirst := false

	for _, line := range content {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "# ") && !skipFirst {
			skipFirst = true
			continue
		}

		if trimmed != "" {
			bodyLines = append(bodyLines, line)
		}
	}

	return strings.Join(bodyLines, "\n")
}
