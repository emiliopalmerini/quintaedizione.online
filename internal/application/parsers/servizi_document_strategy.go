package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// ServiziDocumentStrategy parses services and returns Document entities with HTML content
type ServiziDocumentStrategy struct {
	*BaseDocumentParser
}

// NewServiziDocumentStrategy creates a new Document-based servizi strategy
func NewServiziDocumentStrategy() *ServiziDocumentStrategy {
	return &ServiziDocumentStrategy{
		BaseDocumentParser: NewBaseDocumentParser(),
	}
}

func (s *ServiziDocumentStrategy) ParseDocument(content []string, context *ParsingContext) ([]*domain.Document, error) {
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
		// Skip main title only (check before trimming to preserve structure)
		if strings.HasPrefix(line, "# ") {
			continue
		}

		// Trim whitespace but preserve empty lines for proper markdown structure
		line = strings.TrimSpace(line)

		// Check for new service section (H2)
		if strings.HasPrefix(line, "## ") {
			// Process previous section if exists
			if inSection && len(currentSection) > 0 {
				doc, err := s.parseServiceSection(currentSection, context)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse service section: %v", err))
					currentSection = []string{}
					inSection = false
					continue
				}
				documents = append(documents, doc)
			}

			// Start new section
			currentSection = []string{line}
			inSection = true
		} else if inSection {
			// Add line to current section
			currentSection = append(currentSection, line)
		}
	}

	// Process last section
	if inSection && len(currentSection) > 0 {
		doc, err := s.parseServiceSection(currentSection, context)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last service section: %v", err))
		} else {
			documents = append(documents, doc)
		}
	}

	if len(documents) == 0 {
		return nil, ErrEmptyContent
	}

	return documents, nil
}

func (s *ServiziDocumentStrategy) parseServiceSection(section []string, context *ParsingContext) (*domain.Document, error) {
	if len(section) == 0 {
		return nil, ErrEmptySectionContent
	}

	// Extract title from header
	header := section[0]
	if !strings.HasPrefix(header, "## ") {
		return nil, ErrMissingSectionTitle
	}
	title := strings.TrimPrefix(header, "## ")
	title = strings.TrimSpace(title)

	// Collect content as markdown
	markdownContent := strings.Builder{}
	for i := 1; i < len(section); i++ {
		markdownContent.WriteString(section[i] + "\n")
	}

	// Create filters
	filters := map[string]any{
		"type": "service",
	}

	// Create document
	doc, err := s.CreateDocument(
		title,
		"servizi",
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

func (s *ServiziDocumentStrategy) ContentType() ContentType {
	return ContentTypeServizi
}

func (s *ServiziDocumentStrategy) Name() string {
	return "Servizi Document Strategy"
}

func (s *ServiziDocumentStrategy) Description() string {
	return "Parses Italian D&D 5e services and returns Document entities with HTML content"
}

func (s *ServiziDocumentStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}
