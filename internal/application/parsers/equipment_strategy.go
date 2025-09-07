package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/google/uuid"
)

// EquipmentStrategy implements the Strategy pattern for parsing equipment
type EquipmentStrategy struct {
	*BaseParser
}

// NewEquipmentStrategy creates a new equipment parsing strategy
func NewEquipmentStrategy() ParsingStrategy {
	return &EquipmentStrategy{
		BaseParser: NewBaseParser(
			ContentTypeGear,
			"Equipment Parser",
			"Parses D&D 5e equipment from Italian SRD markdown content",
		),
	}
}

// Parse processes equipment content and returns domain Equipaggiamento objects
func (e *EquipmentStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
	if err := e.Validate(content); err != nil {
		return nil, err
	}

	sections := e.ExtractSections(content, 2) // H2 level for equipment
	var equipment []domain.ParsedEntity

	for _, section := range sections {
		if !section.HasContent() {
			continue
		}

		item, err := e.parseEquipmentSection(section)
		if err != nil {
			e.LogParsingProgress("Error parsing equipment %s: %v", section.Title, err)
			continue
		}

		if item != nil {
			equipment = append(equipment, item)
		}
	}

	return equipment, nil
}

func (e *EquipmentStrategy) parseEquipmentSection(section Section) (*domain.Equipaggiamento, error) {
	if section.Title == "" {
		return nil, fmt.Errorf("equipment section has no title")
	}

	content := section.GetCleanContent()
	if len(content) == 0 {
		return nil, fmt.Errorf("equipment section has no content")
	}

	// Parse equipment information from content
	equipmentContent := strings.Join(content, "\n")
	
	// TODO: Parse equipment stats from content

	// Create domain object - using placeholder values for now
	item := domain.NewEquipaggiamento(
		uuid.New(),
		section.Title,
		domain.Costo{Valore: 0, Valuta: domain.ValutaOro}, // TODO: parse cost
		domain.Peso{Valore: 0.0, Unita: domain.UnitaLibbre}, // TODO: parse weight
		nil, // capacità - TODO: parse
		"", // note - TODO: parse
		equipmentContent,
	)

	return item, nil
}