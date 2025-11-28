package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain"
)

type StrumentiDocumentStrategy struct {
	*BaseDocumentParser
}

func NewStrumentiDocumentStrategy() *StrumentiDocumentStrategy {
	return &StrumentiDocumentStrategy{
		BaseDocumentParser: NewBaseDocumentParser(),
	}
}

func (s *StrumentiDocumentStrategy) ParseDocument(content []string, context *ParsingContext) ([]*domain.Document, error) {
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
				doc, err := s.parseToolSection(currentSection, context)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse tool section: %v", err))
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
		doc, err := s.parseToolSection(currentSection, context)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last tool section: %v", err))
		} else {
			documents = append(documents, doc)
		}
	}

	if len(documents) == 0 {
		return nil, ErrEmptyContent
	}

	return documents, nil
}

func (s *StrumentiDocumentStrategy) parseToolSection(section []string, context *ParsingContext) (*domain.Document, error) {
	if len(section) == 0 {
		return nil, ErrEmptySectionContent
	}

	header := section[0]
	if !strings.HasPrefix(header, "## ") {
		return nil, ErrMissingSectionTitle
	}
	title := strings.TrimPrefix(header, "## ")
	title = strings.TrimSpace(title)

	markdownContent := strings.Builder{}
	for i := 1; i < len(section); i++ {
		markdownContent.WriteString(section[i] + "\n")
	}

	filters := map[string]any{
		"type": "tool",
	}

	doc, err := s.CreateDocument(
		title,
		"strumenti",
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

func (s *StrumentiDocumentStrategy) ContentType() ContentType {
	return ContentTypeStrumenti
}

func (s *StrumentiDocumentStrategy) Name() string {
	return "Strumenti Document Strategy"
}

func (s *StrumentiDocumentStrategy) Description() string {
	return "Parses Italian Quintaedizione 5e tools and returns Document entities with HTML content"
}

func (s *StrumentiDocumentStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}
