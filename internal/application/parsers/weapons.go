package parsers

import (
	"strings"
)

// Italian field mappings for weapons
var weaponFieldsIT = map[string]string{
	"Costo":     "costo",
	"Danno":     "danno",
	"Peso":      "peso",
	"Proprietà": "proprieta",
}

// ParseWeapons parses Italian D&D 5e weapon data from markdown
func ParseWeapons(lines []string) ([]map[string]interface{}, error) {
	items := splitItemsByH2(lines)
	var weapons []map[string]interface{}

	for _, item := range items {
		if len(item.lines) == 0 {
			continue
		}

		weapon := parseWeaponItem(item.title, item.lines)
		if weapon != nil {
			weapons = append(weapons, weapon)
		}
	}

	return weapons, nil
}

// parseWeaponItem parses a single weapon
func parseWeaponItem(title string, lines []string) map[string]interface{} {
	name := strings.TrimSpace(title)
	if name == "" {
		return nil
	}

	// Collect labeled fields
	fields := collectLabeledFieldsFromLines(lines)

	// Map Italian fields to database keys
	mapped := mapWeaponFields(fields)

	// Build weapon object
	weapon := map[string]interface{}{
		"slug":               name,
		"nome":               name,
		"contenuto_markdown": strings.Join(append([]string{"## " + title}, lines...), "\n"),
		"fonte":              "SRD",
		"versione":           "1.0",
	}

	// Add mapped fields
	for key, value := range mapped {
		weapon[key] = value
	}

	// Extract description
	description := extractItemDescription(lines)
	if description != "" {
		weapon["descrizione"] = description
	}

	return weapon
}

// mapWeaponFields maps Italian weapon fields to database structure
func mapWeaponFields(fields map[string]string) map[string]interface{} {
	mapped := make(map[string]interface{})

	for italian, english := range weaponFieldsIT {
		if value, exists := fields[italian]; exists && value != "" {
			mapped[english] = value
		}
	}

	return mapped
}
