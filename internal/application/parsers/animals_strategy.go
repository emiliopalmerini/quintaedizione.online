package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/google/uuid"
)

// AnimalsStrategy implements the Strategy pattern for parsing animals
type AnimalsStrategy struct {
	*BaseParser
}

// NewAnimalsStrategy creates a new animals parsing strategy
func NewAnimalsStrategy() ParsingStrategy {
	return &AnimalsStrategy{
		BaseParser: NewBaseParser(
			ContentTypeAnimals,
			"Animals Parser",
			"Parses D&D 5e animals from Italian SRD markdown content",
		),
	}
}

// Parse processes animals content and returns domain Animale objects
func (a *AnimalsStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
	if err := a.Validate(content); err != nil {
		return nil, err
	}

	sections := a.ExtractSections(content, 2) // H2 level for animals
	var animals []domain.ParsedEntity

	for _, section := range sections {
		if !section.HasContent() {
			continue
		}

		animal, err := a.parseAnimalSection(section)
		if err != nil {
			a.LogParsingProgress("Error parsing animal %s: %v", section.Title, err)
			continue
		}

		if animal != nil {
			animals = append(animals, animal)
		}
	}

	return animals, nil
}

func (a *AnimalsStrategy) parseAnimalSection(section Section) (*domain.Animale, error) {
	if section.Title == "" {
		return nil, fmt.Errorf("animal section has no title")
	}

	content := section.GetCleanContent()
	if len(content) == 0 {
		return nil, fmt.Errorf("animal section has no content")
	}

	// Parse animal information from content
	animalContent := strings.Join(content, "\n")
	
	// TODO: Parse animal stats from content (similar to monsters)

	// Create domain object - using placeholder values for now
	animal := domain.NewAnimale(
		uuid.New(),
		section.Title,
		domain.TagliaMedia, // taglia - TODO: parse from content
		domain.TipoAnimale("Bestia"), // tipo - TODO: parse from content
		domain.ClasseArmatura(10), // CA - TODO: parse from content
		domain.PuntiFerita{Valore: 10}, // PF - TODO: parse from content
		domain.Velocita{Valore: 9, Unita: domain.UnitaMetri}, // velocità - TODO: parse
		[]domain.Caratteristica{}, // caratteristiche - TODO: parse
		[]domain.Tratto{}, // tratti - TODO: parse
		[]domain.Azione{}, // azioni - TODO: parse
		animalContent,
		domain.BonusCompetenza(2), // bonus competenza - TODO: calculate based on CR
	)

	return animal, nil
}