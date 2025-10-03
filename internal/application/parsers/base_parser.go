package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// BaseParser provides common functionality for all parsing strategies
type BaseParser struct {
	// Pre-compiled regex patterns for common field formats
	CostoPattern *regexp.Regexp
	PesoPattern  *regexp.Regexp
}

// NewBaseParser creates a new BaseParser with pre-compiled patterns
func NewBaseParser() *BaseParser {
	return &BaseParser{
		CostoPattern: regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([a-z]{2})`),
		PesoPattern:  regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(kg|g)`),
	}
}

// ParseFieldsFromSection extracts field-value pairs from a markdown section
func (bp *BaseParser) ParseFieldsFromSection(section []string) (map[string]string, string) {
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

	return fields, strings.TrimSpace(contenuto.String())
}

// ExtractNameFromHeader extracts entity name from H2 markdown header
func (bp *BaseParser) ExtractNameFromHeader(header string) (string, error) {
	if !strings.HasPrefix(header, "## ") {
		return "", fmt.Errorf("invalid header format: %s", header)
	}
	name := strings.TrimSpace(strings.TrimPrefix(header, "## "))
	if name == "" {
		return "", fmt.Errorf("empty name in header: %s", header)
	}
	return name, nil
}

// ParseCosto parses cost field into domain.Costo
func (bp *BaseParser) ParseCosto(value string) (domain.Costo, error) {
	if value == "" || value == "—" {
		return domain.NewCosto(0, domain.ValutaOro), nil
	}

	matches := bp.CostoPattern.FindStringSubmatch(value)
	if len(matches) != 3 {
		return domain.Costo{}, fmt.Errorf("invalid costo format: %s", value)
	}

	valore, err := strconv.Atoi(matches[1])
	if err != nil {
		return domain.Costo{}, fmt.Errorf("invalid costo value: %s", matches[1])
	}

	valuta, err := bp.parseValuta(matches[2])
	if err != nil {
		return domain.Costo{}, err
	}

	return domain.NewCosto(valore, valuta), nil
}

// ParsePeso parses weight field into domain.Peso
func (bp *BaseParser) ParsePeso(value string) (domain.Peso, error) {
	if value == "" || value == "—" {
		return domain.NewPeso(0, domain.UnitaKg), nil
	}

	// Handle both comma and dot as decimal separator
	value = strings.ReplaceAll(value, ",", ".")
	matches := bp.PesoPattern.FindStringSubmatch(value)
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

// ValidateRequiredFields checks that all required fields are present
func (bp *BaseParser) ValidateRequiredFields(fields map[string]string, required []string, entityName string) error {
	var missing []string
	for _, field := range required {
		if value, exists := fields[field]; !exists || value == "" {
			missing = append(missing, field)
		}
	}
	
	if len(missing) > 0 {
		return fmt.Errorf("entity '%s' missing required fields: %v", entityName, missing)
	}
	
	return nil
}

// SplitIntoSections splits content into H2 sections
func (bp *BaseParser) SplitIntoSections(content []string) [][]string {
	var sections [][]string
	var currentSection []string
	inSection := false

	for _, line := range content {
		line = strings.TrimSpace(line)
		
		// Skip empty lines and main title
		if line == "" || strings.HasPrefix(line, "# ") {
			continue
		}

		// Check for new section (H2)
		if strings.HasPrefix(line, "## ") {
			// Save previous section if exists
			if inSection && len(currentSection) > 0 {
				sections = append(sections, currentSection)
			}

			// Start new section
			currentSection = []string{line}
			inSection = true
		} else if inSection {
			// Add line to current section
			currentSection = append(currentSection, line)
		}
	}

	// Add last section
	if inSection && len(currentSection) > 0 {
		sections = append(sections, currentSection)
	}

	return sections
}

// parseValuta converts currency string to domain.Valuta
func (bp *BaseParser) parseValuta(currency string) (domain.Valuta, error) {
	switch currency {
	case "mr":
		return domain.ValutaRame, nil
	case "ma":
		return domain.ValutaArgento, nil
	case "me":
		return domain.ValutaElettro, nil
	case "mo":
		return domain.ValutaOro, nil
	case "mp":
		return domain.ValutaPlatino, nil
	default:
		return "", fmt.Errorf("unknown currency: %s", currency)
	}
}