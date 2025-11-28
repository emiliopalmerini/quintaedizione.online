package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain"
)

type MostriDocumentStrategy struct {
	*BaseDocumentParser
}

func NewMostriDocumentStrategy() *MostriDocumentStrategy {
	return &MostriDocumentStrategy{
		BaseDocumentParser: NewBaseDocumentParser(),
	}
}

func (s *MostriDocumentStrategy) ParseDocument(content []string, context *ParsingContext) ([]*domain.Document, error) {
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
				doc, err := s.parseMonsterSection(currentSection, context)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse monster section: %v", err))
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
		doc, err := s.parseMonsterSection(currentSection, context)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last monster section: %v", err))
		} else {
			documents = append(documents, doc)
		}
	}

	if len(documents) == 0 {
		return nil, ErrEmptyContent
	}

	return documents, nil
}

func (s *MostriDocumentStrategy) parseMonsterSection(section []string, context *ParsingContext) (*domain.Document, error) {
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
		"type": "monster",
	}

	doc, err := s.CreateDocument(
		title,
		"mostri",
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

func (s *MostriDocumentStrategy) ContentType() ContentType {
	return ContentTypeMostri
}

func (s *MostriDocumentStrategy) Name() string {
	return "Mostri Document Strategy"
}

func (s *MostriDocumentStrategy) Description() string {
	return "Parses Italian Quintaedizione 5e monsters and returns Document entities with HTML content"
}

func (s *MostriDocumentStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}
