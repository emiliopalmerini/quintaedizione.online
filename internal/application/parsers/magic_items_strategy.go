package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/google/uuid"
)

// MagicItemsStrategy implements the Strategy pattern for parsing magic items
type MagicItemsStrategy struct {
	*BaseParser
}

// NewMagicItemsStrategy creates a new magic items parsing strategy
func NewMagicItemsStrategy() ParsingStrategy {
	return &MagicItemsStrategy{
		BaseParser: NewBaseParser(
			ContentTypeMagicItems,
			"Magic Items Parser",
			"Parses D&D 5e magic items from Italian SRD markdown content",
		),
	}
}

// Parse processes magic items content and returns domain OggettoMagico objects
func (m *MagicItemsStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
	if err := m.Validate(content); err != nil {
		return nil, err
	}

	sections := m.ExtractSections(content, 2) // H2 level for magic items
	var items []domain.ParsedEntity

	for _, section := range sections {
		if !section.HasContent() {
			continue
		}

		item, err := m.parseMagicItemSection(section)
		if err != nil {
			m.LogParsingProgress("Error parsing magic item %s: %v", section.Title, err)
			continue
		}

		if item != nil {
			items = append(items, item)
		}
	}

	return items, nil
}

func (m *MagicItemsStrategy) parseMagicItemSection(section Section) (*domain.OggettoMagico, error) {
	if section.Title == "" {
		return nil, fmt.Errorf("magic item section has no title")
	}

	content := section.GetCleanContent()
	if len(content) == 0 {
		return nil, fmt.Errorf("magic item section has no content")
	}

	// Parse magic item information from content
	itemContent := strings.Join(content, "\n")

	// TODO: Parse magic item stats from content

	// Create domain object - using placeholder values for now
	item := domain.NewOggettoMagico(
		uuid.New(),
		section.Title,
		"",                  // tipo - TODO: parse from content
		domain.RaritaComune, // rarità - TODO: parse from content
		false,               // sintonizzazione - TODO: parse from content
		itemContent,
	)

	return item, nil
}
