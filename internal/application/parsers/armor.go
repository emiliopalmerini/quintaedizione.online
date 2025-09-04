package parsers

import (
	"strings"
)

// Italian field mappings for armor
var armorFieldsIT = map[string]string{
	"Costo":                    "costo",
	"Peso":                     "peso",
	"Classe Armatura":          "classe_armatura",
	"Categoria":                "categoria",
	"Forza richiesta":          "forza_richiesta",
	"Svantaggio Furtività":     "svantaggio_furtivita",
}

// ParseArmor parses Italian D&D 5e armor data from markdown
func ParseArmor(lines []string) ([]map[string]interface{}, error) {
	items := splitItemsByH2(lines)
	var armors []map[string]interface{}

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
func parseArmorItem(title string, lines []string) map[string]interface{} {
	name := strings.TrimSpace(title)
	if name == "" {
		return nil
	}

	// Collect labeled fields
	fields := collectLabeledFieldsFromLines(lines)
	
	// Map Italian fields to database keys
	mapped := mapArmorFields(fields)

	// Build armor object
	armor := map[string]interface{}{
		"slug":                name,
		"nome":                name,
		"contenuto_markdown":  strings.Join(append([]string{"## " + title}, lines...), "\n"),
		"fonte":               "SRD",
		"versione":            "1.0",
	}

	// Add mapped fields
	for key, value := range mapped {
		armor[key] = value
	}

	return armor
}

// mapArmorFields maps Italian armor fields to database structure
func mapArmorFields(fields map[string]string) map[string]interface{} {
	mapped := make(map[string]interface{})

	for italian, english := range armorFieldsIT {
		if value, exists := fields[italian]; exists && value != "" {
			switch english {
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