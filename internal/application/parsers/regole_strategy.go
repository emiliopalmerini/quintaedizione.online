package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

type RegoleStrategy struct{}

func NewRegoleStrategy() *RegoleStrategy {
	return &RegoleStrategy{}
}

func (s *RegoleStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
	if len(content) == 0 {
		return nil, ErrEmptyContent
	}

	if err := context.Validate(); err != nil {
		return nil, err
	}

	var entities []domain.ParsedEntity
	currentSection := []string{}
	inSection := false

	for _, line := range content {
		line = strings.TrimSpace(line)
		
		// Skip empty lines and main title
		if line == "" || strings.HasPrefix(line, "# ") {
			continue
		}

		// Check for new rule section (H2)
		if strings.HasPrefix(line, "## ") {
			// Process previous section if exists
			if inSection && len(currentSection) > 0 {
				regola, err := s.parseRegoleSection(currentSection)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse regola section: %v", err))
					continue
				}
				entities = append(entities, regola)
			}

			// Start new section
			currentSection = []string{line}
			inSection = true
		} else if inSection {
			// Add line to current section
			currentSection = append(currentSection, line)
		}
		// Skip lines before first H2 (introduction text)
	}

	// Process last section
	if inSection && len(currentSection) > 0 {
		regola, err := s.parseRegoleSection(currentSection)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last regola section: %v", err))
		} else {
			entities = append(entities, regola)
		}
	}

	if len(entities) == 0 {
		return nil, ErrEmptyContent
	}

	return entities, nil
}

func (s *RegoleStrategy) parseRegoleSection(section []string) (*domain.Regola, error) {
	if len(section) == 0 {
		return nil, ErrEmptySectionContent
	}

	// Extract name from header
	header := section[0]
	if !strings.HasPrefix(header, "## ") {
		return nil, ErrMissingSectionTitle
	}
	nome := strings.TrimSpace(strings.TrimPrefix(header, "## "))

	// Build content from all remaining lines
	var contentLines []string
	
	for i := 1; i < len(section); i++ {
		line := section[i]
		// Add non-empty lines or preserve structure for tables and paragraphs
		if line != "" || (len(contentLines) > 0 && len(contentLines) < len(section)-2) {
			contentLines = append(contentLines, line)
		}
	}

	// Join with newlines and trim final whitespace
	contenutoStr := strings.Join(contentLines, "\n")
	contenutoStr = strings.TrimSpace(contenutoStr)
	if contenutoStr != "" {
		contenutoStr += "\n" // Add single trailing newline
	}

	regola := domain.NewRegola(
		nome,
		contenutoStr,
	)

	return regola, nil
}

func (s *RegoleStrategy) ContentType() ContentType {
	return ContentTypeRegole
}

func (s *RegoleStrategy) Name() string {
	return "Regole Strategy"
}

func (s *RegoleStrategy) Description() string {
	return "Parses Italian D&D 5e rules (regole) from markdown content"
}

func (s *RegoleStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}