package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

type ArmiStrategy struct{}

func NewArmiStrategy() *ArmiStrategy {
	return &ArmiStrategy{}
}

func (s *ArmiStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
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

		// Check for new weapon section (H2)
		if strings.HasPrefix(line, "## ") {
			// Process previous section if exists
			if inSection && len(currentSection) > 0 {
				arma, err := s.parseWeaponSection(currentSection)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse weapon section: %v", err))
					continue
				}
				entities = append(entities, arma)
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
		arma, err := s.parseWeaponSection(currentSection)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last weapon section: %v", err))
		} else {
			entities = append(entities, arma)
		}
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
	if !strings.HasPrefix(header, "## ") {
		return nil, ErrMissingSectionTitle
	}
	nome := strings.TrimSpace(strings.TrimPrefix(header, "## "))

	// Parse fields
	fields := make(map[string]string)
	contenuto := strings.Builder{}
	
	for i := 1; i < len(section); i++ {
		line := section[i]
		contenuto.WriteString(line + "\n")
		
		// Parse field format: **Field:** value
		if strings.HasPrefix(line, "**") && strings.Contains(line, ":**") {
			parts := strings.SplitN(line, ":**", 2)
			if len(parts) == 2 {
				fieldName := strings.TrimSpace(strings.Trim(parts[0], "*"))
				fieldValue := strings.TrimSpace(parts[1])
				fields[fieldName] = fieldValue
			}
		}
	}

	// Parse required fields
	costo, err := s.parseCosto(fields["Costo"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse costo: %w", err)
	}

	peso, err := s.parsePeso(fields["Peso"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse peso: %w", err)
	}

	danno := fields["Danno"]
	if danno == "" {
		return nil, fmt.Errorf("missing danno field")
	}

	categoria, err := s.parseCategoria(fields["Categoria"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse categoria: %w", err)
	}

	proprieta, err := s.parseProprieta(fields["Proprietà"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse proprieta: %w", err)
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
		strings.TrimSpace(contenuto.String()),
	)

	return arma, nil
}

func (s *ArmiStrategy) parseCosto(value string) (domain.Costo, error) {
	if value == "" || value == "—" {
		return domain.NewCosto(0, domain.ValutaOro), nil
	}

	// Parse format like "5 mo", "10 ma", etc.
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([a-z]{2})`)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 3 {
		return domain.Costo{}, fmt.Errorf("invalid costo format: %s", value)
	}

	valore, err := strconv.Atoi(matches[1])
	if err != nil {
		return domain.Costo{}, fmt.Errorf("invalid costo value: %s", matches[1])
	}

	var valuta domain.Valuta
	switch matches[2] {
	case "mr":
		valuta = domain.ValutaRame
	case "ma":
		valuta = domain.ValutaArgento
	case "me":
		valuta = domain.ValutaElettro
	case "mo":
		valuta = domain.ValutaOro
	case "mp":
		valuta = domain.ValutaPlatino
	default:
		return domain.Costo{}, fmt.Errorf("unknown currency: %s", matches[2])
	}

	return domain.NewCosto(valore, valuta), nil
}

func (s *ArmiStrategy) parsePeso(value string) (domain.Peso, error) {
	if value == "" || value == "—" {
		return domain.NewPeso(0, domain.UnitaKg), nil
	}

	// Parse format like "3,5 kg", "4.5 kg", etc.
	// Handle both comma and dot as decimal separator
	value = strings.ReplaceAll(value, ",", ".")
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*kg`)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 2 {
		return domain.Peso{}, fmt.Errorf("invalid peso format: %s", value)
	}

	valore, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return domain.Peso{}, fmt.Errorf("invalid peso value: %s", matches[1])
	}

	return domain.NewPeso(valore, domain.UnitaKg), nil
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
		
		// Handle properties with parentheses like "Munizioni (Freccia)"
		if strings.Contains(prop, "(") {
			prop = strings.Split(prop, "(")[0]
			prop = strings.TrimSpace(prop)
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
			// Skip unknown properties rather than failing
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