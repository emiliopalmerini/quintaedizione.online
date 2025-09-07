package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/google/uuid"
)

// ArmorStrategy implements the Strategy pattern for parsing armor
type ArmorStrategy struct {
	*BaseParser
}

// NewArmorStrategy creates a new armor parsing strategy
func NewArmorStrategy() ParsingStrategy {
	return &ArmorStrategy{
		BaseParser: NewBaseParser(
			ContentTypeArmor,
			"Armor Parser",
			"Parses D&D 5e armor from Italian SRD markdown content",
		),
	}
}

// Parse processes armor content and returns domain Armatura objects
func (a *ArmorStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
	if err := a.Validate(content); err != nil {
		return nil, err
	}

	sections := a.ExtractSections(content, 2) // H2 level for armor
	var armors []domain.ParsedEntity

	for _, section := range sections {
		if !section.HasContent() {
			continue
		}

		armor, err := a.parseArmorSection(section)
		if err != nil {
			a.LogParsingProgress("Error parsing armor %s: %v", section.Title, err)
			continue
		}

		if armor != nil {
			armors = append(armors, armor)
		}
	}

	return armors, nil
}

func (a *ArmorStrategy) parseArmorSection(section Section) (*domain.Armatura, error) {
	if section.Title == "" {
		return nil, fmt.Errorf("armor section has no title")
	}

	content := section.GetCleanContent()
	if len(content) == 0 {
		return nil, fmt.Errorf("armor section has no content")
	}

	// Parse armor information from content
	armorContent := strings.Join(content, "\n")
	
	// TODO: Parse armor stats from content (cost, AC, weight, etc.)

	// Create domain object - using placeholder values for now
	armor := domain.NewArmatura(
		uuid.New(),
		section.Title,
		domain.Costo{Valore: 0, Valuta: domain.ValutaOro}, // TODO: parse cost
		domain.Peso{Valore: 0.0, Unita: domain.UnitaLibbre}, // TODO: parse weight
		domain.CategoriaArmatura("Armatura Leggera"), // TODO: parse category
		domain.CAArmatura{Base: 10, ModificatoreDes: true}, // TODO: parse AC
		0,    // forza richiesta - TODO: parse
		false, // svantaggio furtività - TODO: parse
		armorContent,
	)

	return armor, nil
}