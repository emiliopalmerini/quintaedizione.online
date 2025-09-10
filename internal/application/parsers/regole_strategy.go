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
	inGlossaryContent := false

	for _, line := range content {
		line = strings.TrimSpace(line)
		
		// Skip empty lines and main title
		if line == "" || strings.HasPrefix(line, "# ") {
			continue
		}

		// Skip introduction content until we reach the glossary definitions
		if !inGlossaryContent && strings.Contains(line, "Di seguito trovi le definizioni") {
			inGlossaryContent = true
			continue
		}

		if !inGlossaryContent {
			continue
		}

		// Check for new rule section (H2)
		if strings.HasPrefix(line, "## ") {
			// Process previous section if exists
			if inSection && len(currentSection) > 0 {
				rule, err := s.parseGlossaryEntry(currentSection)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse glossary entry: %v", err))
					currentSection = []string{}
					inSection = false
					continue
				}
				entities = append(entities, rule)
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
		rule, err := s.parseGlossaryEntry(currentSection)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last glossary entry: %v", err))
		} else {
			entities = append(entities, rule)
		}
	}

	if len(entities) == 0 {
		return nil, ErrEmptyContent
	}

	return entities, nil
}

func (s *RegoleStrategy) parseGlossaryEntry(section []string) (*domain.Regola, error) {
	if len(section) == 0 {
		return nil, ErrEmptySectionContent
	}

	// Extract name from header
	header := section[0]
	if !strings.HasPrefix(header, "## ") {
		return nil, ErrMissingSectionTitle
	}
	nome := strings.TrimSpace(strings.TrimPrefix(header, "## "))

	// Collect all content
	contenuto := strings.Builder{}
	for i := 1; i < len(section); i++ {
		contenuto.WriteString(section[i] + "\n")
	}

	rule := domain.NewRegola(
		nome,
		strings.TrimSpace(contenuto.String()),
	)

	return rule, nil
}

func (s *RegoleStrategy) ContentType() ContentType {
	return ContentTypeRegole
}

func (s *RegoleStrategy) Name() string {
	return "Regole Strategy"
}

func (s *RegoleStrategy) Description() string {
	return "Parses Italian D&D 5e rules from markdown content"
}

func (s *RegoleStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}