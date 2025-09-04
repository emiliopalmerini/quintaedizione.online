package parsers

import (
	"strings"
)

// Italian field mappings for magic items
var magicItemFieldsIT = map[string]string{
	"Oggetto magico":      "tipo",
	"Rarità":              "rarita",
	"Sintonizzazione":     "sintonizzazione",
}

// ParseMagicItems parses Italian D&D 5e magic item data from markdown
func ParseMagicItems(lines []string) ([]map[string]interface{}, error) {
	items := splitItemsByH2(lines)
	var magicItems []map[string]interface{}

	for _, item := range items {
		if len(item.lines) == 0 {
			continue
		}

		magicItem := parseMagicItemItem(item.title, item.lines)
		if magicItem != nil {
			magicItems = append(magicItems, magicItem)
		}
	}

	return magicItems, nil
}

// parseMagicItemItem parses a single magic item
func parseMagicItemItem(title string, lines []string) map[string]interface{} {
	name := strings.TrimSpace(title)
	if name == "" {
		return nil
	}

	// Collect labeled fields
	fields := collectLabeledFieldsFromLines(lines)
	
	// Map Italian fields to database keys
	mapped := mapMagicItemFields(fields)

	// Build magic item object
	magicItem := map[string]interface{}{
		"slug":                name,
		"nome":                name,
		"contenuto_markdown":  strings.Join(append([]string{"## " + title}, lines...), "\n"),
		"fonte":               "SRD",
		"versione":            "1.0",
	}

	// Add mapped fields
	for key, value := range mapped {
		magicItem[key] = value
	}

	// Extract description
	description := extractItemDescription(lines)
	if description != "" {
		magicItem["descrizione"] = description
	}

	return magicItem
}

// mapMagicItemFields maps Italian magic item fields to database structure
func mapMagicItemFields(fields map[string]string) map[string]interface{} {
	mapped := make(map[string]interface{})

	for italian, english := range magicItemFieldsIT {
		if value, exists := fields[italian]; exists && value != "" {
			switch english {
			case "sintonizzazione":
				// Convert to boolean
				needsAttune := strings.Contains(strings.ToLower(value), "sì") || 
					         strings.Contains(strings.ToLower(value), "richiesta") ||
					         strings.Contains(strings.ToLower(value), "necessaria")
				mapped[english] = needsAttune
			default:
				mapped[english] = value
			}
		}
	}

	return mapped
}