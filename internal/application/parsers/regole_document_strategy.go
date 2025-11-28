package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain"
)

// RegoleDocumentStrategy parses rules and returns Document entities with HTML content
type RegoleDocumentStrategy struct {
	*BaseDocumentParser
}

// NewRegoleDocumentStrategy creates a new Document-based regole strategy
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
		// Skip main title only (check before trimming to preserve structure)
		if strings.HasPrefix(line, "# ") {
			continue
		}

		// Trim whitespace but preserve empty lines for proper markdown structure
		line = strings.TrimSpace(line)

		// Skip introduction content until we reach the first glossary entry (after abbreviations table)
		if !inGlossaryContent {
			// Start parsing when we find the first H2 section after the introduction
			if strings.HasPrefix(line, "## ") && !strings.Contains(line, "Convenzioni") && !strings.Contains(line, "Abbreviazioni") {
				inGlossaryContent = true
			} else {
				continue
			}
		}

		// Check for new rule section (H2)
		if strings.HasPrefix(line, "## ") {
			// Process previous section if exists
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

	// Extract name from header
	header := section[0]
	if !strings.HasPrefix(header, "## ") {
		return nil, ErrMissingSectionTitle
	}
	title := strings.TrimSpace(strings.TrimPrefix(header, "## "))

	// Collect all content as markdown (excluding the title)
	markdownContent := strings.Builder{}
	for i := 1; i < len(section); i++ {
		markdownContent.WriteString(section[i] + "\n")
	}

	// Create filters
	filters := map[string]any{
		"type": "rule",
	}

	// Create document using base parser helper
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
