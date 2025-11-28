package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain"
)

type IncantesimiDocumentStrategy struct {
	*BaseDocumentParser
}

func NewIncantesimiDocumentStrategy() *IncantesimiDocumentStrategy {
	return &IncantesimiDocumentStrategy{
		BaseDocumentParser: NewBaseDocumentParser(),
	}
}

func (s *IncantesimiDocumentStrategy) ParseDocument(content []string, context *ParsingContext) ([]*domain.Document, error) {
	if len(content) == 0 {
		return nil, ErrEmptyContent
	}

	if err := context.Validate(); err != nil {
		return nil, err
	}

	var documents []*domain.Document
	currentSection := []string{}
	inSection := false

	for _, line := range content {

		if strings.HasPrefix(line, "# ") {
			continue
		}

		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "## ") {

			if inSection && len(currentSection) > 0 {
				doc, err := s.parseSpellSection(currentSection, context)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse spell section: %v", err))
					currentSection = []string{}
					inSection = false
					continue
				}
				documents = append(documents, doc)
			}

			currentSection = []string{line}
			inSection = true
		} else if inSection {

			currentSection = append(currentSection, line)
		}
	}

	if inSection && len(currentSection) > 0 {
		doc, err := s.parseSpellSection(currentSection, context)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last spell section: %v", err))
		} else {
			documents = append(documents, doc)
		}
	}

	if len(documents) == 0 {
		return nil, ErrEmptyContent
	}

	return documents, nil
}

func (s *IncantesimiDocumentStrategy) parseSpellSection(section []string, context *ParsingContext) (*domain.Document, error) {
	if len(section) == 0 {
		return nil, ErrEmptySectionContent
	}

	header := section[0]
	if !strings.HasPrefix(header, "## ") {
		return nil, ErrMissingSectionTitle
	}
	title := strings.TrimPrefix(header, "## ")
	title = strings.TrimSpace(title)

	title = strings.TrimSuffix(title, ".")
	title = strings.Trim(title, "*")

	markdownContent := strings.Builder{}
	for i := 1; i < len(section); i++ {
		markdownContent.WriteString(section[i] + "\n")
	}

	filters := map[string]any{
		"type": "spell",
	}

	doc, err := s.CreateDocument(
		title,
		"incantesimi",
		context.Filename,
		context.Language,
		strings.TrimSpace(markdownContent.String()),
		filters,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return doc, nil
}

func (s *IncantesimiDocumentStrategy) ContentType() ContentType {
	return ContentTypeIncantesimi
}

func (s *IncantesimiDocumentStrategy) Name() string {
	return "Incantesimi Document Strategy"
}

func (s *IncantesimiDocumentStrategy) Description() string {
	return "Parses Italian Quintaedizione 5e spells and returns Document entities with HTML content"
}

func (s *IncantesimiDocumentStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}
