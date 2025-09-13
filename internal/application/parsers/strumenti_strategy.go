package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

type StrumentiStrategy struct{}

func NewStrumentiStrategy() *StrumentiStrategy {
	return &StrumentiStrategy{}
}

func (s *StrumentiStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
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

		// Check for new tool section (H2)
		if strings.HasPrefix(line, "## ") {
			// Process previous section if exists
			if inSection && len(currentSection) > 0 {
				tool, err := s.parseToolSection(currentSection)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse tool section: %v", err))
					currentSection = []string{}
					inSection = false
					continue
				}
				entities = append(entities, tool)
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
		tool, err := s.parseToolSection(currentSection)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last tool section: %v", err))
		} else {
			entities = append(entities, tool)
		}
	}

	if len(entities) == 0 {
		return nil, ErrEmptyContent
	}

	return entities, nil
}

func (s *StrumentiStrategy) parseToolSection(section []string) (*domain.Strumento, error) {
	if len(section) == 0 {
		return nil, ErrEmptySectionContent
	}

	// Extract name and cost from header
	header := section[0]
	if !strings.HasPrefix(header, "## ") {
		return nil, ErrMissingSectionTitle
	}
	
	headerText := strings.TrimSpace(strings.TrimPrefix(header, "## "))
	nome, costoFromHeader := s.parseNomeECosto(headerText)

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

	// Parse fields
	costo := s.parseCosto(fields["Costo"])
	if costo.Valore == 0 && costoFromHeader.Valore > 0 {
		costo = costoFromHeader
	}

	peso, err := s.parsePeso(fields["Peso"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse peso: %w", err)
	}

	abilitaAssociata := s.parseAbilita(fields["Abilità"])
	utilizzi := s.parseUtilizzi(fields["Utilizzo"])
	creazioni := s.parseCreazioni(fields["Creazioni"])

	tool := domain.NewStrumento(
		nome,
		costo,
		peso,
		abilitaAssociata,
		utilizzi,
		creazioni,
		strings.TrimSpace(contenuto.String()),
	)

	return tool, nil
}

func (s *StrumentiStrategy) parseNomeECosto(headerText string) (string, domain.Costo) {
	// Parse format like "Strumenti da alchimista (50 MO)"
	re := regexp.MustCompile(`^(.+?)\s*\(([^)]+)\)$`)
	matches := re.FindStringSubmatch(headerText)
	
	if len(matches) == 3 {
		nome := strings.TrimSpace(matches[1])
		costoStr := strings.TrimSpace(matches[2])
		costo := s.parseCosto(costoStr)
		return nome, costo
	}
	
	return headerText, domain.NewCosto(0, domain.ValutaOro)
}

func (s *StrumentiStrategy) parseCosto(value string) domain.Costo {
	if value == "" || value == "—" {
		return domain.NewCosto(0, domain.ValutaOro)
	}

	// Parse format like "50 MO", "10 mo", etc.
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([A-Za-z]{1,2})`)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 3 {
		return domain.NewCosto(0, domain.ValutaOro)
	}

	valore, err := strconv.Atoi(matches[1])
	if err != nil {
		return domain.NewCosto(0, domain.ValutaOro)
	}

	var valuta domain.Valuta
	valutaStr := strings.ToLower(matches[2])
	switch valutaStr {
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
		valuta = domain.ValutaOro
	}

	return domain.NewCosto(valore, valuta)
}

func (s *StrumentiStrategy) parsePeso(value string) (domain.Peso, error) {
	if value == "" || value == "—" {
		return domain.NewPeso(0, domain.UnitaKg), nil
	}

	// Parse format like "3,5 kg", "4 kg", "250 g", etc.
	value = strings.ReplaceAll(value, ",", ".")
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(kg|g)`)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 3 {
		return domain.Peso{}, fmt.Errorf("invalid peso format: %s", value)
	}

	valore, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return domain.Peso{}, fmt.Errorf("invalid peso value: %s", matches[1])
	}

	unit := matches[2]
	
	// Convert grams to kilograms if needed
	if unit == "g" {
		valore = valore / 1000.0
	}

	return domain.NewPeso(valore, domain.UnitaKg), nil
}

func (s *StrumentiStrategy) parseAbilita(value string) string {
	return strings.TrimSpace(value)
}

func (s *StrumentiStrategy) parseUtilizzi(value string) []domain.UtilizzoStrumento {
	if value == "" {
		return []domain.UtilizzoStrumento{}
	}

	var utilizzi []domain.UtilizzoStrumento
	
	// Split by " o " for multiple uses
	parts := strings.Split(value, " o ")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		// Extract CD from format like "Identificare una sostanza (CD 15)"
		re := regexp.MustCompile(`(.+?)\s*\(CD\s*(\d+)\)`)
		matches := re.FindStringSubmatch(part)
		if len(matches) == 3 {
			descrizione := strings.TrimSpace(matches[1])
			cd, err := strconv.Atoi(matches[2])
			if err == nil && cd >= 5 && cd <= 30 {
				utilizzo, err := domain.NewUtilizzoStrumento(descrizione, cd)
				if err == nil {
					utilizzi = append(utilizzi, utilizzo)
				}
			}
		} else {
			// If no CD found, default to CD 10
			utilizzo, err := domain.NewUtilizzoStrumento(part, 10)
			if err == nil {
				utilizzi = append(utilizzi, utilizzo)
			}
		}
	}
	
	return utilizzi
}

func (s *StrumentiStrategy) parseCreazioni(value string) []string {
	if value == "" {
		return []string{}
	}
	
	// Split by comma and clean up
	parts := strings.Split(value, ",")
	var creazioni []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			// Remove italic formatting if present
			part = strings.Trim(part, "*")
			creazioni = append(creazioni, part)
		}
	}
	
	return creazioni
}

func (s *StrumentiStrategy) ContentType() ContentType {
	return ContentTypeStrumenti
}

func (s *StrumentiStrategy) Name() string {
	return "Strumenti Strategy"
}

func (s *StrumentiStrategy) Description() string {
	return "Parses Italian D&D 5e tools (strumenti) from markdown content"
}

func (s *StrumentiStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}