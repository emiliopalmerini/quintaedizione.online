package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain"
)

// BackgroundsDocumentStrategy parses backgrounds and returns Document entities with HTML content
type BackgroundsDocumentStrategy struct {
	*BaseDocumentParser
}

// NewBackgroundsDocumentStrategy creates a new Document-based backgrounds strategy
func NewBackgroundsDocumentStrategy() *BackgroundsDocumentStrategy {
	return &BackgroundsDocumentStrategy{
		BaseDocumentParser: NewBaseDocumentParser(),
	}
}

func (s *BackgroundsDocumentStrategy) ParseDocument(content []string, context *ParsingContext) ([]*domain.Document, error) {
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

		// Check for new background section (H2)
		if strings.HasPrefix(line, "## ") {
			// Process previous section if exists
			if inSection && len(currentSection) > 0 {
				doc, err := s.parseBackgroundSection(currentSection, context)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse background section: %v", err))
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
		doc, err := s.parseBackgroundSection(currentSection, context)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last background section: %v", err))
		} else {
			documents = append(documents, doc)
		}
	}

	if len(documents) == 0 {
		return nil, ErrEmptyContent
	}

	return documents, nil
}

func (s *BackgroundsDocumentStrategy) parseBackgroundSection(section []string, context *ParsingContext) (*domain.Document, error) {
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
		"type": "background",
	}

	// Create document
	doc, err := s.CreateDocument(
		title,
		"backgrounds",
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

func (s *BackgroundsDocumentStrategy) ContentType() ContentType {
	return ContentTypeBackgrounds
}

func (s *BackgroundsDocumentStrategy) Name() string {
	return "Backgrounds Document Strategy"
}

func (s *BackgroundsDocumentStrategy) Description() string {
	return "Parses Italian Quintaedizione 5e backgrounds and returns Document entities with HTML content"
}

func (s *BackgroundsDocumentStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}
