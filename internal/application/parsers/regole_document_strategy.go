package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain"
)

type RegoleDocumentStrategy struct {
	*BaseDocumentParser
}

func NewRegoleDocumentStrategy() *RegoleDocumentStrategy {
	return &RegoleDocumentStrategy{
		BaseDocumentParser: NewBaseDocumentParser(),
	}
}

func (s *RegoleDocumentStrategy) ParseDocument(content []string, context *ParsingContext) ([]*domain.Document, error) {
	if len(content) == 0 {
		return nil, ErrEmptyContent
	}

	if err := context.Validate(); err != nil {
		return nil, err
	}

	var documents []*domain.Document
	currentSection := []string{}
	inSection := false
	inGlossaryContent := false

	for _, line := range content {

		if strings.HasPrefix(line, "# ") {
			continue
		}

		line = strings.TrimSpace(line)

		if !inGlossaryContent {

			if strings.HasPrefix(line, "## ") && !strings.Contains(line, "Convenzioni") && !strings.Contains(line, "Abbreviazioni") {
				inGlossaryContent = true
			} else {
				continue
			}
		}

		if strings.HasPrefix(line, "## ") {

			if inSection && len(currentSection) > 0 {
				doc, err := s.parseRuleSection(currentSection, context)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse rule section: %v", err))
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
		doc, err := s.parseRuleSection(currentSection, context)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last rule section: %v", err))
		} else {
			documents = append(documents, doc)
		}
	}

	if len(documents) == 0 {
		return nil, ErrEmptyContent
	}

	return documents, nil
}

func (s *RegoleDocumentStrategy) parseRuleSection(section []string, context *ParsingContext) (*domain.Document, error) {
	if len(section) == 0 {
		return nil, ErrEmptySectionContent
	}

	header := section[0]
	if !strings.HasPrefix(header, "## ") {
		return nil, ErrMissingSectionTitle
	}
	title := strings.TrimSpace(strings.TrimPrefix(header, "## "))

	markdownContent := strings.Builder{}
	for i := 1; i < len(section); i++ {
		markdownContent.WriteString(section[i] + "\n")
	}

	filters := map[string]any{
		"type": "rule",
	}

	doc, err := s.CreateDocument(
		title,
		"regole",
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

func (s *RegoleDocumentStrategy) ContentType() ContentType {
	return ContentTypeRegole
}

func (s *RegoleDocumentStrategy) Name() string {
	return "Regole Document Strategy"
}

func (s *RegoleDocumentStrategy) Description() string {
	return "Parses Italian Quintaedizione 5e rules and returns Document entities with HTML content"
}

func (s *RegoleDocumentStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}
