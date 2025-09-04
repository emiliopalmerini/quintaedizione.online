package parsers

import (
	"regexp"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

var (
	BoldFieldRE = regexp.MustCompile(`\*\*([^*]+)\*\*`)
)

// ParseBackgrounds parses D&D 5e background data from markdown
func ParseBackgrounds(lines []string) ([]map[string]interface{}, error) {
	sections := splitSections(lines)
	var backgrounds []map[string]interface{}

	for _, section := range sections {
		if len(section) < 2 {
			continue
		}

		// Check if it's a background section (H2 header)
		match := SectionH2RE.FindStringSubmatch(section[0])
		if len(match) < 2 {
			continue
		}

		backgroundName := strings.TrimSpace(match[1])
		if backgroundName == "" {
			continue
		}

		background, err := parseBackgroundSection(backgroundName, section[1:])
		if err != nil {
			continue // Skip invalid backgrounds but continue processing
		}

		if background != nil {
			backgrounds = append(backgrounds, background)
		}
	}

	return backgrounds, nil
}

// parseBackgroundSection parses a single background section
func parseBackgroundSection(backgroundName string, lines []string) (map[string]interface{}, error) {
	background := map[string]interface{}{
		"nome":               backgroundName,
		"slug":               domain.NormalizeID(backgroundName),
		"descrizione":        "",
		"contenuto_markdown": strings.Join(lines, "\n"),
		"fonte":              "SRD",
		"versione":           "1.0",
	}

	// Parse background fields
	fields := parseBackgroundFields(lines)
	for key, value := range fields {
		background[key] = value
	}

	// Extract description (text before first bold field)
	description := extractBackgroundDescription(lines)
	if description != "" {
		background["descrizione"] = description
	}

	return background, nil
}

// parseBackgroundFields parses labeled fields in the background
func parseBackgroundFields(lines []string) map[string]interface{} {
	fields := make(map[string]interface{})

	var currentField string
	var currentContent []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for bold field markers
		boldMatches := BoldFieldRE.FindAllStringSubmatch(line, -1)
		if len(boldMatches) > 0 {
			// Save previous field
			if currentField != "" && len(currentContent) > 0 {
				fields[normalizeFieldName(currentField)] = strings.Join(currentContent, " ")
			}

			// Start new field
			currentField = boldMatches[0][1]
			// Remove the bold marker from the line and use the rest as content
			cleanLine := BoldFieldRE.ReplaceAllString(line, "")
			cleanLine = strings.TrimSpace(cleanLine)
			if cleanLine != "" {
				currentContent = []string{cleanLine}
			} else {
				currentContent = []string{}
			}
		} else {
			// Add to current content
			if currentField != "" {
				currentContent = append(currentContent, line)
			}
		}
	}

	// Save last field
	if currentField != "" && len(currentContent) > 0 {
		fields[normalizeFieldName(currentField)] = parseFieldValue(currentField, strings.Join(currentContent, " "))
	}

	return fields
}

// normalizeFieldName normalizes field names to database keys
func normalizeFieldName(fieldName string) string {
	field := strings.ToLower(strings.TrimSpace(fieldName))

	// Map Italian field names
	switch field {
	case "abilità competenti":
		return "abilita_competenze"
	case "linguaggi":
		return "linguaggi"
	case "equipaggiamento":
		return "equipaggiamento"
	case "caratteristica: ideali":
		return "ideali"
	case "caratteristica: legami":
		return "legami"
	case "caratteristica: difetti":
		return "difetti"
	case "tratto distintivo":
		return "tratto_distintivo"
	case "variante":
		return "variante"
	default:
		// Generic normalization
		field = strings.ReplaceAll(field, " ", "_")
		field = strings.ReplaceAll(field, ":", "")
		return field
	}
}

// parseFieldValue parses specific field values based on field type
func parseFieldValue(fieldName string, value string) interface{} {
	fieldLower := strings.ToLower(fieldName)

	switch {
	case strings.Contains(fieldLower, "equipaggiamento"):
		return parseEquipmentOptions(value)
	case strings.Contains(fieldLower, "abilità"):
		return parseSkillCompetencies(value)
	case strings.Contains(fieldLower, "linguaggi"):
		return parseLanguages(value)
	case strings.Contains(fieldLower, "ideali"),
		strings.Contains(fieldLower, "legami"),
		strings.Contains(fieldLower, "difetti"):
		return parseCharacteristicsList(value)
	default:
		return value
	}
}

// parseEquipmentOptions parses equipment selection options
func parseEquipmentOptions(value string) []map[string]interface{} {
	if value == "" {
		return []map[string]interface{}{}
	}

	// Strip emphasis markers
	cleanValue := strings.ReplaceAll(value, "*", "")
	cleanValue = strings.ReplaceAll(cleanValue, "**", "")
	cleanValue = strings.ReplaceAll(cleanValue, "  ", " ")

	// Look for option pattern: (A) ... ; oppure (B) ...
	optionRegex := regexp.MustCompile(`\(A\)\s*(.+?);\s*(?:oppure|o)\s*\(B\)\s*(.+?)(?:$|;)`)
	match := optionRegex.FindStringSubmatch(cleanValue)

	if len(match) >= 3 {
		// Parse both options
		aItems := parseItemList(match[1])
		bItems := parseItemList(match[2])

		return []map[string]interface{}{
			{"etichetta": "Opzione A", "oggetti": aItems},
			{"etichetta": "Opzione B", "oggetti": bItems},
		}
	}

	// Fallback: treat as single option
	items := parseItemList(cleanValue)
	if len(items) > 0 {
		return []map[string]interface{}{
			{"etichetta": "Default", "oggetti": items},
		}
	}

	return []map[string]interface{}{}
}

// parseItemList parses a comma-separated list of items
func parseItemList(value string) []string {
	items := strings.Split(value, ",")
	var result []string

	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}

	return result
}

// parseLanguages parses language competencies
func parseLanguages(value string) interface{} {
	// Check for choice pattern
	if strings.Contains(value, "Scegli") || strings.Contains(value, "scegli") {
		// Extract choice information
		choiceRegex := regexp.MustCompile(`(?i)scegli\s+(\d+)`)
		match := choiceRegex.FindStringSubmatch(value)

		if len(match) > 1 {
			// Parse as choice structure
			return map[string]interface{}{
				"tipo":    "scelta",
				"numero":  match[1],
				"opzioni": "qualsiasi",
			}
		}
	}

	// Parse as list
	languages := parseItemList(value)
	return languages
}

// parseCharacteristicsList parses characteristics like ideals, bonds, flaws
func parseCharacteristicsList(value string) []string {
	// Split by numbered list or semicolons
	lines := strings.Split(value, "\n")
	var characteristics []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove numbering (1., 2., etc.)
		numberRegex := regexp.MustCompile(`^\d+\.\s*`)
		line = numberRegex.ReplaceAllString(line, "")
		line = strings.TrimSpace(line)

		if line != "" {
			characteristics = append(characteristics, line)
		}
	}

	return characteristics
}

// extractBackgroundDescription extracts the background description
func extractBackgroundDescription(lines []string) string {
	var descLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Stop at first bold field
		if BoldFieldRE.MatchString(line) {
			break
		}

		if line != "" {
			descLines = append(descLines, line)
		}
	}

	return strings.Join(descLines, "\n")
}
