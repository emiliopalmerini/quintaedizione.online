package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// AnimaliDocumentStrategy parses animals and returns Document entities with HTML content
type AnimaliDocumentStrategy struct {
	*BaseDocumentParser
}

// NewAnimaliDocumentStrategy creates a new Document-based animali strategy
func NewAnimaliDocumentStrategy() *AnimaliDocumentStrategy {
	return &AnimaliDocumentStrategy{
		BaseDocumentParser: NewBaseDocumentParser(),
	}
}

func (s *AnimaliDocumentStrategy) ParseDocument(content []string, context *ParsingContext) ([]*domain.Document, error) {
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

		// Check for new animal section (H2)
		if strings.HasPrefix(line, "## ") {
			// Process previous section if exists
			if inSection && len(currentSection) > 0 {
				doc, err := s.parseAnimalSection(currentSection, context)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse animal section: %v", err))
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
		doc, err := s.parseAnimalSection(currentSection, context)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last animal section: %v", err))
		} else {
			documents = append(documents, doc)
		}
	}

	if len(documents) == 0 {
		return nil, ErrEmptyContent
	}

	return documents, nil
}

func (s *AnimaliDocumentStrategy) parseAnimalSection(section []string, context *ParsingContext) (*domain.Document, error) {
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
		"type": "animal",
	}

	// Create document
	doc, err := s.CreateDocument(
		title,
		"animali",
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

func (s *AnimaliDocumentStrategy) ContentType() ContentType {
	return ContentTypeAnimali
}

func (s *AnimaliDocumentStrategy) Name() string {
	return "Animali Document Strategy"
}

func (s *AnimaliDocumentStrategy) Description() string {
	return "Parses Italian D&D 5e animals and returns Document entities with HTML content"
}

func (s *AnimaliDocumentStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}
