package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/google/uuid"
)

// FeatsStrategy implements the Strategy pattern for parsing feats
type FeatsStrategy struct {
	*BaseParser
}

// NewFeatsStrategy creates a new feats parsing strategy
func NewFeatsStrategy() ParsingStrategy {
	return &FeatsStrategy{
		BaseParser: NewBaseParser(
			ContentTypeFeats,
			"Feats Parser",
			"Parses D&D 5e feats from Italian SRD markdown content",
		),
	}
}

// Parse processes feats content and returns domain Talento objects
func (f *FeatsStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
	if err := f.Validate(content); err != nil {
		return nil, err
	}

	sections := f.ExtractSections(content, 2) // H2 level for feats
	var feats []domain.ParsedEntity

	for _, section := range sections {
		if !section.HasContent() {
			continue
		}

		feat, err := f.parseFeatSection(section)
		if err != nil {
			f.LogParsingProgress("Error parsing feat %s: %v", section.Title, err)
			continue
		}

		if feat != nil {
			feats = append(feats, feat)
		}
	}

	return feats, nil
}

func (f *FeatsStrategy) parseFeatSection(section Section) (*domain.Talento, error) {
	if section.Title == "" {
		return nil, fmt.Errorf("feat section has no title")
	}

	content := section.GetCleanContent()
	if len(content) == 0 {
		return nil, fmt.Errorf("feat section has no content")
	}

	// Parse feat information from content
	featContent := strings.Join(content, "\n")
	
	// TODO: Parse feat details from content (prerequisites, benefits, etc.)

	// Create domain object - using placeholder values for now
	feat := domain.NewTalento(
		uuid.New(),
		section.Title,
		domain.CategoriaTalento("Generale"), // categoria - TODO: parse from content
		"", // prerequisiti - TODO: parse from content
		[]string{}, // benefici - TODO: parse from content
		featContent,
	)

	return feat, nil
}