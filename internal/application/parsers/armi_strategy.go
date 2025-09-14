package parsers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

type ArmiStrategy struct{
	*BaseParser
	validator *Validator
	// Pre-compiled patterns for weapon-specific parsing
	proprietaPattern *regexp.Regexp
}

func NewArmiStrategy() *ArmiStrategy {
	return &ArmiStrategy{
		BaseParser: NewBaseParser(),
		validator:  NewValidator(),
		proprietaPattern: regexp.MustCompile(`([^,(]+)(?:\([^)]+\))?`),
	}
}

func (s *ArmiStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
	if len(content) == 0 {
		return nil, ErrEmptyContent
	}

	if err := context.Validate(); err != nil {
		return nil, err
	}

	// Validate content structure before parsing
	if err := s.validator.ValidateContent(content, ContentTypeArmi); err != nil {
		return nil, fmt.Errorf("content validation failed: %w", err)
	}

	var entities []domain.ParsedEntity
	sections := s.SplitIntoSections(content)

	for _, section := range sections {
		if len(section) == 0 {
			continue
		}

		arma, err := s.parseWeaponSection(section)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse weapon section: %v", err))
			continue
		}

		// Validate parsed entity
		if err := s.validator.ValidateArma(arma); err != nil {
			context.Logger.Error(fmt.Sprintf("Weapon validation failed: %v", err))
			continue
		}

		entities = append(entities, arma)
	}

	if len(entities) == 0 {
		return nil, ErrEmptyContent
	}

	return entities, nil
}

func (s *ArmiStrategy) parseWeaponSection(section []string) (*domain.Arma, error) {
	if len(section) == 0 {
		return nil, ErrEmptySectionContent
	}

	// Extract name from header
	header := section[0]
	nome, err := s.ExtractNameFromHeader(header)
	if err != nil {
		return nil, err
	}

	// Validate section structure
	if err := s.validator.ValidateSection(section, nome); err != nil {
		return nil, err
	}

	// Extract fields using base parser
	fields, contenuto := s.ParseFieldsFromSection(section)

	// Validate required fields
	requiredFields := []string{"Costo", "Peso", "Danno", "Categoria"}
	if err := s.validator.ValidateRequiredFields(fields, requiredFields, nome); err != nil {
		return nil, err
	}

	// Parse required fields using base parser
	costo, err := s.ParseCosto(fields["Costo"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse costo for '%s': %w", nome, err)
	}

	peso, err := s.ParsePeso(fields["Peso"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse peso for '%s': %w", nome, err)
	}

	danno := fields["Danno"]

	categoria, err := s.parseCategoria(fields["Categoria"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse categoria for '%s': %w", nome, err)
	}

	proprieta, err := s.parseProprieta(fields["Proprietà"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse proprieta for '%s': %w", nome, err)
	}

	maestria := fields["Maestria"]
	
	gittata := s.parseGittata(fields["Gittata"], fields["Gittata lunga"])

	arma := domain.NewArma(
		nome,
		costo,
		peso,
		danno,
		categoria,
		proprieta,
		maestria,
		gittata,
		contenuto,
	)

	return arma, nil
}


func (s *ArmiStrategy) parseCategoria(value string) (domain.CategoriaArma, error) {
	switch value {
	case "Semplice da Mischia":
		return domain.CategoriaArmaSimpliceMischia, nil
	case "Semplice a Distanza":
		return domain.CategoriaArmaSempliceDistanza, nil
	case "Marziale da Mischia":
		return domain.CategoriaArmaMarzialeMischia, nil
	case "Marziale a Distanza":
		return domain.CategoriaArmaMarzialeDistanza, nil
	default:
		return "", fmt.Errorf("unknown categoria: %s", value)
	}
}

func (s *ArmiStrategy) parseProprieta(value string) ([]domain.ProprietaArma, error) {
	if value == "" || value == "—" {
		return []domain.ProprietaArma{}, nil
	}

	// Split by comma and parse each property
	parts := strings.Split(value, ",")
	proprieta := make([]domain.ProprietaArma, 0, len(parts))

	for _, part := range parts {
		prop := strings.TrimSpace(part)
		
		// Use regex to extract property name, ignoring parentheses
		matches := s.proprietaPattern.FindStringSubmatch(prop)
		if len(matches) > 1 {
			prop = strings.TrimSpace(matches[1])
		}

		var proprietaArma domain.ProprietaArma
		switch prop {
		case "Accurata":
			proprietaArma = domain.ProprietaAccurata
		case "Leggera":
			proprietaArma = domain.ProprietaLeggera
		case "Da Lancio":
			proprietaArma = domain.ProprietaDaLancio
		case "Versatile":
			proprietaArma = domain.ProprietaVersatile
		case "A Due Mani":
			proprietaArma = domain.ProprietaDueManiBast
		case "Pesante":
			// Map to appropriate enum - using existing pattern
			proprietaArma = domain.ProprietaArmaturaPesante
		case "Portata":
			proprietaArma = domain.ProprietaPortata
		case "Munizioni":
			proprietaArma = domain.ProprietaMunizioni
		case "Ricarica":
			proprietaArma = domain.ProprietaCaricare
		case "Speciale":
			proprietaArma = domain.ProprietaSpeciale
		default:
			// Log unknown properties but continue parsing
			continue
		}
		
		proprieta = append(proprieta, proprietaArma)
	}

	return proprieta, nil
}

func (s *ArmiStrategy) parseGittata(gittata, gittataLunga string) *domain.GittataArma {
	// If both are empty or "0", return nil
	if (gittata == "" || gittata == "0") && (gittataLunga == "" || gittataLunga == "0") {
		return nil
	}

	// If one is "0", convert to empty string for cleaner display
	if gittata == "0" {
		gittata = ""
	}
	if gittataLunga == "0" {
		gittataLunga = ""
	}

	return &domain.GittataArma{
		Normale: gittata,
		Lunga:   gittataLunga,
	}
}

func (s *ArmiStrategy) ContentType() ContentType {
	return ContentTypeArmi
}

func (s *ArmiStrategy) Name() string {
	return "Armi Strategy"
}

func (s *ArmiStrategy) Description() string {
	return "Parses Italian D&D 5e weapons (armi) from markdown content"
}

func (s *ArmiStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}