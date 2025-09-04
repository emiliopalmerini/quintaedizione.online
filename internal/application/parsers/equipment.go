package parsers

import (
	"strings"
)

// Generic equipment parser for tools, services, gear

// Italian field mappings for tools
var toolFieldsIT = map[string]string{
	"Costo":      "costo",
	"Peso":       "peso",
	"Categoria":  "categoria",
}

// ParseTools parses Italian D&D 5e tool data from markdown
func ParseTools(lines []string) ([]map[string]interface{}, error) {
	return parseGenericEquipment(lines, "strumento")
}

// ParseServices parses Italian D&D 5e service data from markdown
func ParseServices(lines []string) ([]map[string]interface{}, error) {
	return parseGenericEquipment(lines, "servizio")
}

// ParseGear parses Italian D&D 5e gear data from markdown
func ParseGear(lines []string) ([]map[string]interface{}, error) {
	return parseGenericEquipment(lines, "equipaggiamento")
}

// parseGenericEquipment parses generic equipment items
func parseGenericEquipment(lines []string, itemType string) ([]map[string]interface{}, error) {
	items := splitItemsByH2(lines)
	var equipment []map[string]interface{}

	for _, item := range items {
		if len(item.lines) == 0 {
			continue
		}

		equip := parseEquipmentItem(item.title, item.lines, itemType)
		if equip != nil {
			equipment = append(equipment, equip)
		}
	}

	return equipment, nil
}

// parseEquipmentItem parses a single equipment item
func parseEquipmentItem(title string, lines []string, itemType string) map[string]interface{} {
	name := strings.TrimSpace(title)
	if name == "" {
		return nil
	}

	// Collect labeled fields
	fields := collectLabeledFieldsFromLines(lines)
	
	// Map Italian fields to database keys
	mapped := mapEquipmentFields(fields)

	// Build equipment object
	equipment := map[string]interface{}{
		"slug":                name,
		"nome":                name,
		"tipo":                itemType,
		"contenuto_markdown":  strings.Join(append([]string{"## " + title}, lines...), "\n"),
		"fonte":               "SRD",
		"versione":            "1.0",
	}

	// Add mapped fields
	for key, value := range mapped {
		equipment[key] = value
	}

	// Extract description
	description := extractItemDescription(lines)
	if description != "" {
		equipment["descrizione"] = description
	}

	return equipment
}

// mapEquipmentFields maps Italian equipment fields to database structure
func mapEquipmentFields(fields map[string]string) map[string]interface{} {
	mapped := make(map[string]interface{})

	for italian, english := range toolFieldsIT {
		if value, exists := fields[italian]; exists && value != "" {
			mapped[english] = value
		}
	}

	return mapped
}

// extractItemDescription extracts item description from lines
func extractItemDescription(lines []string) string {
	var descLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip labeled fields (lines with ":")
		if strings.Contains(line, ":") && strings.Index(line, ":") < len(line)/2 {
			continue
		}

		if line != "" {
			descLines = append(descLines, line)
		}
	}

	return strings.Join(descLines, " ")
}