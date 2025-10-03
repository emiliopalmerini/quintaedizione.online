package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

type ArmatureStrategy struct{}

func NewArmatureStrategy() *ArmatureStrategy {
	return &ArmatureStrategy{}
}

func (s *ArmatureStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
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

		// Check for new armor section (H2)
		if strings.HasPrefix(line, "## ") {
			// Process previous section if exists
			if inSection && len(currentSection) > 0 {
				armor, err := s.parseArmorSection(currentSection)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse armor section: %v", err))
					continue
				}
				entities = append(entities, armor)
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
		armor, err := s.parseArmorSection(currentSection)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last armor section: %v", err))
		} else {
			entities = append(entities, armor)
		}
	}

	if len(entities) == 0 {
		return nil, ErrEmptyContent
	}

	return entities, nil
}

func (s *ArmatureStrategy) parseArmorSection(section []string) (*domain.Armatura, error) {
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

		// Parse field format: **Field:** value
		if strings.HasPrefix(line, "**") && strings.Contains(line, ":**") {
			parts := strings.SplitN(line, ":**", 2)
			if len(parts) == 2 {
				fieldName := strings.TrimSpace(strings.Trim(parts[0], "*"))
				fieldValue := strings.TrimSpace(parts[1])
				fields[fieldName] = fieldValue

				// Add period if not present and add double newline
				if !strings.HasSuffix(strings.TrimSpace(line), ".") {
					line += "."
				}
				contenuto.WriteString(line + "\n\n")
			}
		} else if line != "" {
			// Non-field lines (descriptions, etc)
			if strings.HasSuffix(strings.TrimSpace(line), ".") {
				contenuto.WriteString(line + "\n\n")
			} else {
				contenuto.WriteString(line + "\n")
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

	categoria, err := s.parseCategoria(fields["Categoria"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse categoria: %w", err)
	}

	classeArmatura, err := s.parseClasseArmatura(fields)
	if err != nil {
		return nil, fmt.Errorf("failed to parse classe armatura: %w", err)
	}

	forzaRichiesta := s.parseForzaRichiesta(fields["Forza richiesta"])
	svantaggioFurtivita := s.parseSvantaggioFurtivita(fields["Svantaggio Furtività"])

	armor := domain.NewArmatura(
		nome,
		costo,
		peso,
		categoria,
		classeArmatura,
		forzaRichiesta,
		svantaggioFurtivita,
		strings.TrimSpace(contenuto.String()),
	)

	return armor, nil
}

func (s *ArmatureStrategy) parseCosto(value string) (domain.Costo, error) {
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

func (s *ArmatureStrategy) parsePeso(value string) (domain.Peso, error) {
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

func (s *ArmatureStrategy) parseCategoria(value string) (domain.CategoriaArmatura, error) {
	switch value {
	case "Leggera":
		return domain.CategoriaArmaturaLeggera, nil
	case "Media":
		return domain.CategoriaArmaturaMedia, nil
	case "Pesante":
		return domain.CategoriaArmaturaPesante, nil
	case "Scudo":
		return domain.CategoriaScudo, nil
	default:
		return "", fmt.Errorf("unknown categoria: %s", value)
	}
}

func (s *ArmatureStrategy) parseClasseArmatura(fields map[string]string) (domain.CAArmatura, error) {
	caBaseStr := fields["CA Base"]
	if caBaseStr == "" {
		return domain.CAArmatura{}, fmt.Errorf("missing CA Base")
	}

	caBase, err := strconv.Atoi(caBaseStr)
	if err != nil {
		return domain.CAArmatura{}, fmt.Errorf("invalid CA Base: %s", caBaseStr)
	}

	// Parse CA + Des
	caDesStr := fields["CA + Des"]
	modificatoreDes := strings.ToLower(caDesStr) == "sì" || strings.ToLower(caDesStr) == "si"

	// Parse Limite Des
	limiteDes := 0
	limiteDesStr := fields["Limite Des"]
	if limiteDesStr != "" && limiteDesStr != "—" {
		// Handle format like "+2"
		limiteDesStr = strings.TrimPrefix(limiteDesStr, "+")
		if limiteDes, err = strconv.Atoi(limiteDesStr); err != nil {
			// If parsing fails, default to 0 (no limit)
			limiteDes = 0
		}
	}

	return domain.CAArmatura{
		Base:            caBase,
		ModificatoreDes: modificatoreDes,
		LimiteDes:       limiteDes,
	}, nil
}

func (s *ArmatureStrategy) parseForzaRichiesta(value string) int {
	if value == "" || value == "—" {
		return 0
	}

	forza, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}

	return forza
}

func (s *ArmatureStrategy) parseSvantaggioFurtivita(value string) bool {
	return strings.ToLower(value) == "sì" || strings.ToLower(value) == "si"
}

func (s *ArmatureStrategy) ContentType() ContentType {
	return ContentTypeArmature
}

func (s *ArmatureStrategy) Name() string {
	return "Armature Strategy"
}

func (s *ArmatureStrategy) Description() string {
	return "Parses Italian D&D 5e armor (armature) from markdown content"
}

func (s *ArmatureStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}