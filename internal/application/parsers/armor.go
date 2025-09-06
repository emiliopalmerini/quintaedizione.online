package parsers

import (
	"maps"
	"strconv"
	"strings"
)

// Italian field mappings for armor
var armorFieldsIT = map[string]string{
	"Costo":                "costo",
	"Peso":                 "peso",
	"Classe Armatura":      "classe_armatura",
	"Categoria":            "categoria",
	"Forza richiesta":      "forza_richiesta",
	"Svantaggio Furtività": "svantaggio_furtivita",
}

// ParseArmor parses Italian D&D 5e armor data from markdown
func ParseArmor(lines []string) ([]map[string]any, error) {
	items := splitItemsByH2(lines)
	var armors []map[string]any

	for _, item := range items {
		if len(item.lines) == 0 {
			continue
		}

		armor := parseArmorItem(item.title, item.lines)
		if armor != nil {
			armors = append(armors, armor)
		}
	}

	return armors, nil
}

// parseArmorItem parses a single armor item
func parseArmorItem(title string, lines []string) map[string]any {
	name := strings.TrimSpace(title)
	if name == "" {
		return nil
	}

	// Collect labeled fields
	fields := collectLabeledFieldsFromLines(lines)

	// Map Italian fields to database keys
	mapped := mapArmorFields(fields)

	// Build armor object
	armor := map[string]any{
		"slug":               name,
		"nome":               name,
		"contenuto_markdown": strings.Join(append([]string{"## " + title}, lines...), "\n"),
		"fonte":              "SRD",
		"versione":           "1.0",
	}

	// Add mapped fields
	maps.Copy(armor, mapped)

	return armor
}

// mapArmorFields maps Italian armor fields to database structure
func mapArmorFields(fields map[string]string) map[string]any {
	mapped := make(map[string]any)

	for italian, english := range armorFieldsIT {
		if value, exists := fields[italian]; exists && value != "" {
			switch english {
			case "costo":
				// Parse cost as object with amount and currency
				mapped[english] = parseCosto(value)
			case "peso":
				// Parse weight as object with amount and unit
				mapped[english] = parsePeso(value)
			case "forza_richiesta":
				// Parse strength requirement (convert "—" to 0)
				if value == "—" || value == "-" {
					mapped[english] = 0
				} else {
					// Try to parse as integer
					if val, err := strconv.Atoi(value); err == nil {
						mapped[english] = val
					} else {
						mapped[english] = value
					}
				}
			case "svantaggio_furtivita":
				// Convert to boolean
				mapped[english] = strings.ToLower(value) == "sì" || strings.ToLower(value) == "si" || strings.ToLower(value) == "yes"
			default:
				mapped[english] = value
			}
		}
	}

	return mapped
}

// parseCosto parses cost string into structured object
func parseCosto(costStr string) map[string]any {
	// Handle "—" case
	if costStr == "—" || costStr == "-" {
		return map[string]any{
			"valore": 0,
			"valuta": "",
		}
	}

	// Split by space to separate amount and currency
	parts := strings.Fields(costStr)
	if len(parts) >= 2 {
		amountStr := parts[0]
		currency := strings.Join(parts[1:], " ")

		// Try to parse amount (handle decimal comma)
		amountStr = strings.Replace(amountStr, ",", ".", 1)
		if amount, err := strconv.ParseFloat(amountStr, 64); err == nil {
			return map[string]any{
				"valore": amount,
				"valuta": currency,
			}
		}
	}

	// Fallback to original string
	return map[string]any{
		"valore": 0,
		"valuta": "",
	}
}

// parsePeso parses weight string into structured object
func parsePeso(pesoStr string) map[string]any {
	// Handle "—" case
	if pesoStr == "—" || pesoStr == "-" {
		return map[string]any{
			"valore": 0.0,
			"unita":  "",
		}
	}

	// Split by space to separate amount and unit
	parts := strings.Fields(pesoStr)
	if len(parts) >= 2 {
		amountStr := parts[0]
		unit := strings.Join(parts[1:], " ")

		// Try to parse amount (handle decimal comma)
		amountStr = strings.Replace(amountStr, ",", ".", 1)
		if amount, err := strconv.ParseFloat(amountStr, 64); err == nil {
			return map[string]any{
				"valore": amount,
				"unita":  unit,
			}
		}
	}

	// Fallback to original string
	return map[string]any{
		"valore": 0.0,
		"unita":  "",
	}
}
