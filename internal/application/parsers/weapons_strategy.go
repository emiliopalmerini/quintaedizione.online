package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/google/uuid"
)

// WeaponsStrategy implements the Strategy pattern for parsing weapons
type WeaponsStrategy struct {
	*BaseParser
}

// NewWeaponsStrategy creates a new weapons parsing strategy
func NewWeaponsStrategy() ParsingStrategy {
	return &WeaponsStrategy{
		BaseParser: NewBaseParser(
			ContentTypeWeapons,
			"Weapons Parser",
			"Parses D&D 5e weapons from Italian SRD markdown content",
		),
	}
}

// Parse processes weapon content and returns domain Arma objects
func (w *WeaponsStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
	if err := w.Validate(content); err != nil {
		return nil, err
	}

	sections := w.ExtractSections(content, 2) // H2 level for weapons
	var weapons []domain.ParsedEntity

	for _, section := range sections {
		if !section.HasContent() {
			continue
		}

		weapon, err := w.parseWeaponSection(section)
		if err != nil {
			w.LogParsingProgress("Error parsing weapon %s: %v", section.Title, err)
			continue
		}

		if weapon != nil {
			weapons = append(weapons, weapon)
		}
	}

	return weapons, nil
}

func (w *WeaponsStrategy) parseWeaponSection(section Section) (*domain.Arma, error) {
	if section.Title == "" {
		return nil, fmt.Errorf("weapon section has no title")
	}

	content := section.GetCleanContent()
	if len(content) == 0 {
		return nil, fmt.Errorf("weapon section has no content")
	}

	// Parse weapon information from content
	weaponContent := strings.Join(content, "\n")
	
	// TODO: Parse weapon stats from content (cost, damage, weight, properties)
	// This is simplified - in reality we'd parse the structured content

	// Create domain object - using placeholder values for now
	weapon := domain.NewArma(
		uuid.New(),
		section.Title,
		domain.Costo{Valore: 0, Valuta: domain.ValutaOro}, // TODO: parse cost from content
		domain.Peso{Valore: 0.0, Unita: domain.UnitaLibbre}, // TODO: parse weight from content
		"", // danno - TODO: parse damage from content
		domain.CategoriaArma("Arma Semplice"), // TODO: parse category from content
		[]domain.ProprietaArma{}, // TODO: parse properties from content
		"", // maestria - TODO: parse mastery from content
		nil, // gittata - TODO: parse range from content
		weaponContent,
	)

	return weapon, nil
}