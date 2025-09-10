package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

type EquipaggiamentiStrategy struct{}

func NewEquipaggiamentiStrategy() *EquipaggiamentiStrategy {
	return &EquipaggiamentiStrategy{}
}

func (s *EquipaggiamentiStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
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

		// Check for new equipment section (H2)
		if strings.HasPrefix(line, "## ") {
			// Process previous section if exists
			if inSection && len(currentSection) > 0 {
				equipment, err := s.parseEquipmentSection(currentSection)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse equipment section: %v", err))
					currentSection = []string{}
					inSection = false
					continue
				}
				entities = append(entities, equipment)
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
		equipment, err := s.parseEquipmentSection(currentSection)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last equipment section: %v", err))
		} else {
			entities = append(entities, equipment)
		}
	}

	if len(entities) == 0 {
		return nil, ErrEmptyContent
	}

	return entities, nil
}

func (s *EquipaggiamentiStrategy) parseEquipmentSection(section []string) (*domain.Equipaggiamento, error) {
	if len(section) == 0 {
		return nil, ErrEmptySectionContent
	}

	// Extract name and cost from header
	header := section[0]
	if !strings.HasPrefix(header, "## ") {
		return nil, ErrMissingSectionTitle
	}
	
	headerText := strings.TrimSpace(strings.TrimPrefix(header, "## "))
	nome, costo := s.parseNomeECosto(headerText)

	// Parse fields and content
	fields := make(map[string]string)
	contenuto := strings.Builder{}
	var description []string
	
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
		} else if line != "" && !strings.HasPrefix(line, "**") {
			// This is description text
			description = append(description, line)
		}
	}

	// Parse peso
	peso, err := s.parsePeso(fields["Peso"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse peso: %w", err)
	}

	// Parse capacity if present
	var capacita *domain.Capacita
	if capacitaStr := fields["Capacità"]; capacitaStr != "" {
		cap := s.parseCapacita(capacitaStr)
		capacita = &cap
	}

	// Combine description as notes
	note := strings.Join(description, " ")

	equipment := domain.NewEquipaggiamento(
		nome,
		costo,
		peso,
		capacita,
		note,
		strings.TrimSpace(contenuto.String()),
	)

	return equipment, nil
}

func (s *EquipaggiamentiStrategy) parseNomeECosto(headerText string) (string, domain.Costo) {
	// Parse format like "Frecce (20) (1 mo)" or "Item Name"
	
	// Extract cost from parentheses at the end
	re := regexp.MustCompile(`^(.+?)\s*\(([^)]+)\)$`)
	matches := re.FindStringSubmatch(headerText)
	
	if len(matches) == 3 {
		nome := strings.TrimSpace(matches[1])
		costoStr := strings.TrimSpace(matches[2])
		
		// Check if there are multiple parentheses - extract the last one for cost
		if strings.Count(headerText, "(") > 1 {
			// More complex parsing for items like "Frecce (20) (1 mo)"
			lastParenIndex := strings.LastIndex(headerText, "(")
			if lastParenIndex != -1 {
				nome = strings.TrimSpace(headerText[:lastParenIndex])
				costoStr = strings.TrimSpace(headerText[lastParenIndex+1:])
				costoStr = strings.TrimSuffix(costoStr, ")")
			}
		}
		
		costo := s.parseCosto(costoStr)
		return nome, costo
	}
	
	// No cost found in header, default to zero cost
	return headerText, domain.NewCosto(0, domain.ValutaOro)
}

func (s *EquipaggiamentiStrategy) parseCosto(value string) domain.Costo {
	if value == "" || value == "—" {
		return domain.NewCosto(0, domain.ValutaOro)
	}

	// Parse format like "1 mo", "4 mr", etc.
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([a-z]{2})`)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 3 {
		return domain.NewCosto(0, domain.ValutaOro)
	}

	valore, err := strconv.Atoi(matches[1])
	if err != nil {
		return domain.NewCosto(0, domain.ValutaOro)
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
		valuta = domain.ValutaOro
	}

	return domain.NewCosto(valore, valuta)
}

func (s *EquipaggiamentiStrategy) parsePeso(value string) (domain.Peso, error) {
	if value == "" || value == "—" {
		return domain.NewPeso(0, domain.UnitaKg), nil
	}

	// Parse format like "0,5 kg", "0.5 kg", etc.
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

func (s *EquipaggiamentiStrategy) parseCapacita(value string) domain.Capacita {
	if value == "" || value == "—" {
		return domain.Capacita{Valore: 0, Unita: domain.UnitaLitri}
	}

	// Parse format like "10 l", "5 litri", etc.
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([a-zA-Z]+)`)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 3 {
		return domain.Capacita{Valore: 0, Unita: domain.UnitaLitri}
	}

	valore, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return domain.Capacita{Valore: 0, Unita: domain.UnitaLitri}
	}

	unita := domain.UnitaLitri
	unitaStr := strings.ToLower(matches[2])
	switch unitaStr {
	case "l", "litri", "litro":
		unita = domain.UnitaLitri
	case "ml", "millilitri", "millilitro":
		unita = domain.UnitaMillilitri
	case "gal", "galloni", "gallone":
		unita = domain.UnitaGalloni
	}

	return domain.Capacita{Valore: valore, Unita: unita}
}

func (s *EquipaggiamentiStrategy) ContentType() ContentType {
	return ContentTypeEquipaggiamenti
}

func (s *EquipaggiamentiStrategy) Name() string {
	return "Equipaggiamenti Strategy"
}

func (s *EquipaggiamentiStrategy) Description() string {
	return "Parses Italian D&D 5e general equipment from markdown content"
}

func (s *EquipaggiamentiStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}