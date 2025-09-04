package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

var (
	SectionH2RE = regexp.MustCompile(`^## (.+)$`)
	SectionH4RE = regexp.MustCompile(`^#### (.+)$`)
)

// ParseClasses parses D&D 5e class data from markdown
func ParseClasses(lines []string) ([]map[string]interface{}, error) {
	sections := splitSections(lines)
	var classes []map[string]interface{}

	for _, section := range sections {
		if len(section) < 2 {
			continue
		}

		// Check if it's a class section (H2 header)
		match := SectionH2RE.FindStringSubmatch(section[0])
		if len(match) < 2 {
			continue
		}

		className := strings.TrimSpace(match[1])
		if className == "" {
			continue
		}

		class, err := parseClassSection(className, section[1:])
		if err != nil {
			continue // Skip invalid classes but continue processing
		}

		if class != nil {
			classes = append(classes, class)
		}
	}

	return classes, nil
}

// parseClassSection parses a single class section
func parseClassSection(className string, lines []string) (map[string]interface{}, error) {
	class := map[string]interface{}{
		"nome":               className,
		"slug":               domain.NormalizeID(className),
		"descrizione":        "",
		"contenuto_markdown": strings.Join(lines, "\n"),
		"fonte":              "SRD",
		"versione":           "1.0",
	}

	// Parse base traits table
	baseTraits := parseBaseTraitsTable(lines)
	for key, value := range baseTraits {
		class[key] = value
	}

	// Parse class features by level
	featuresByLevel := parseFeaturesByLevel(lines)
	if len(featuresByLevel) > 0 {
		class["caratteristiche_per_livello"] = featuresByLevel
	}

	// Parse spellcasting progression if present
	spellcasting := parseSpellcastingProgression(lines)
	if len(spellcasting) > 0 {
		class["progressione_incantesimi"] = spellcasting
	}

	// Extract description (text before first table or feature)
	description := extractClassDescription(lines)
	if description != "" {
		class["descrizione"] = description
	}

	return class, nil
}

// parseBaseTraitsTable parses the base traits table
func parseBaseTraitsTable(lines []string) map[string]interface{} {
	traits := make(map[string]interface{})

	// Find the base traits table
	for i, line := range lines {
		if strings.Contains(line, "Tratti base del") {
			// Look for table starting after this line
			tableStart := findTableStart(lines[i:])
			if tableStart > 0 {
				headers, rows := parseMarkdownTable(lines[i+tableStart:])
				if len(headers) >= 2 && len(rows) > 0 {
					parseTraitsFromTable(rows, traits)
				}
			}
			break
		}
	}

	return traits
}

// parseFeaturesByLevel parses class features organized by level
func parseFeaturesByLevel(lines []string) map[string]interface{} {
	featuresByLevel := make(map[string]interface{})

	// Look for level sections (#### Livello X)
	for i, line := range lines {
		match := SectionH4RE.FindStringSubmatch(line)
		if len(match) < 2 {
			continue
		}

		levelText := strings.TrimSpace(match[1])
		if !strings.HasPrefix(levelText, "Livello ") {
			continue
		}

		// Extract level number
		levelStr := strings.TrimPrefix(levelText, "Livello ")
		levelStr = strings.TrimSpace(levelStr)
		level, err := strconv.Atoi(levelStr)
		if err != nil {
			continue
		}

		// Extract features for this level
		features := extractLevelFeatures(lines, i+1)
		if len(features) > 0 {
			featuresByLevel[strconv.Itoa(level)] = features
		}
	}

	return featuresByLevel
}

// parseSpellcastingProgression parses spellcasting tables
func parseSpellcastingProgression(lines []string) map[string]interface{} {
	spellcasting := make(map[string]interface{})

	// Look for spellcasting table headers
	for i, line := range lines {
		if strings.Contains(line, "Incantesimi") && strings.Contains(line, "Livello") {
			headers, rows := parseMarkdownTable(lines[i:])
			if len(headers) > 0 && len(rows) > 0 {
				spellcasting = parseSpellcastingTable(headers, rows)
				break
			}
		}
	}

	return spellcasting
}

// Helper functions

// splitSections splits lines into sections by H2 headers
func splitSections(lines []string) [][]string {
	var sections [][]string
	var currentSection []string

	for _, line := range lines {
		if SectionH2RE.MatchString(line) {
			if len(currentSection) > 0 {
				sections = append(sections, currentSection)
			}
			currentSection = []string{line}
		} else {
			currentSection = append(currentSection, line)
		}
	}

	if len(currentSection) > 0 {
		sections = append(sections, currentSection)
	}

	return sections
}

// findTableStart finds the start of a markdown table
func findTableStart(lines []string) int {
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "|") {
			return i
		}
	}
	return -1
}

// parseMarkdownTable parses a markdown table
func parseMarkdownTable(lines []string) ([]string, [][]string) {
	var headers []string
	var rows [][]string

	tableStarted := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			if tableStarted {
				break // End of table
			}
			continue
		}

		if !tableStarted {
			// Parse headers
			headers = parseTableRow(line)
			tableStarted = true
			continue
		}

		// Skip separator row
		if strings.Contains(line, "---") {
			continue
		}

		// Parse data row
		row := parseTableRow(line)
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	return headers, rows
}

// parseTableRow parses a single table row
func parseTableRow(line string) []string {
	line = strings.Trim(line, "|")
	parts := strings.Split(line, "|")

	var result []string
	for _, part := range parts {
		result = append(result, strings.TrimSpace(part))
	}

	return result
}

// parseTraitsFromTable extracts traits from base traits table
func parseTraitsFromTable(rows [][]string, traits map[string]interface{}) {
	for _, row := range rows {
		if len(row) < 2 {
			continue
		}

		key := strings.TrimSpace(row[0])
		value := strings.TrimSpace(row[1])

		switch {
		case strings.Contains(key, "Caratteristica primaria"):
			traits["caratteristica_primaria"] = value
		case strings.Contains(key, "Dado Punti Ferita"):
			// Extract hit die (d6, d8, d10, d12)
			if strings.Contains(value, "d") {
				re := regexp.MustCompile(`[dD](\d+)`)
				match := re.FindStringSubmatch(value)
				if len(match) > 1 {
					traits["dado_vita"] = fmt.Sprintf("d%s", match[1])
				}
			}
		case strings.Contains(key, "Tiri salvezza competenti"):
			// Split by 'e' or commas
			saves := strings.Split(strings.ReplaceAll(value, " e ", ", "), ",")
			var cleanSaves []string
			for _, save := range saves {
				cleanSaves = append(cleanSaves, strings.TrimSpace(save))
			}
			traits["salvezze_competenze"] = cleanSaves
		case strings.Contains(key, "Abilità competenti"):
			traits["abilita_competenze_opzioni"] = parseSkillCompetencies(value)
		case strings.Contains(key, "Armi competenti"):
			weapons := strings.Split(strings.ReplaceAll(value, " e ", ", "), ",")
			var cleanWeapons []string
			for _, weapon := range weapons {
				cleanWeapons = append(cleanWeapons, strings.TrimSpace(weapon))
			}
			traits["armi_competenze"] = cleanWeapons
		case strings.Contains(key, "Armature competenti"):
			armor := strings.Split(strings.ReplaceAll(value, " e ", ", "), ",")
			var cleanArmor []string
			for _, a := range armor {
				cleanArmor = append(cleanArmor, strings.TrimSpace(a))
			}
			traits["armature_competenze"] = cleanArmor
		}
	}
}

// parseSkillCompetencies parses skill competency options
func parseSkillCompetencies(value string) map[string]interface{} {
	result := make(map[string]interface{})

	if strings.Contains(value, "Scegli") {
		// Parse "Scegli X tra: ..."
		re := regexp.MustCompile(`Scegli (\d+)`)
		match := re.FindStringSubmatch(value)
		if len(match) > 1 {
			count, err := strconv.Atoi(match[1])
			if err == nil {
				result["scegli"] = count
			}
		}

		// Extract options after colon
		if strings.Contains(value, ":") {
			parts := strings.SplitN(value, ":", 2)
			if len(parts) > 1 {
				optionsText := strings.TrimSpace(parts[1])
				options := strings.Split(optionsText, ",")
				var cleanOptions []string
				for _, option := range options {
					cleanOptions = append(cleanOptions, strings.TrimSpace(option))
				}
				result["opzioni"] = cleanOptions
			}
		}
	} else {
		// All abilities listed
		options := strings.Split(value, ",")
		var cleanOptions []string
		for _, option := range options {
			cleanOptions = append(cleanOptions, strings.TrimSpace(option))
		}
		result["scegli"] = len(cleanOptions)
		result["opzioni"] = cleanOptions
	}

	return result
}

// extractLevelFeatures extracts features for a specific level
func extractLevelFeatures(lines []string, startIdx int) []map[string]interface{} {
	var features []map[string]interface{}

	// Read until next level or section
	for i := startIdx; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Stop at next level or major section
		if SectionH4RE.MatchString(line) || SectionH2RE.MatchString(line) || SectionH3RE.MatchString(line) {
			break
		}

		// Look for feature descriptions
		if line != "" && !strings.HasPrefix(line, "|") {
			// This could be a feature - extract it
			feature := map[string]interface{}{
				"descrizione": line,
			}
			features = append(features, feature)
		}
	}

	return features
}

// parseSpellcastingTable parses spellcasting progression table
func parseSpellcastingTable(headers []string, rows [][]string) map[string]interface{} {
	spellcasting := make(map[string]interface{})

	// Process each row (level)
	for _, row := range rows {
		if len(row) < 2 {
			continue
		}

		level := strings.TrimSpace(row[0])
		if level == "" {
			continue
		}

		levelData := make(map[string]interface{})

		// Map each column to spell slot data
		for j := 1; j < len(row) && j < len(headers); j++ {
			header := strings.TrimSpace(headers[j])
			value := strings.TrimSpace(row[j])

			if value != "" && value != "-" && value != "—" {
				levelData[header] = value
			}
		}

		if len(levelData) > 0 {
			spellcasting[level] = levelData
		}
	}

	return spellcasting
}

// extractClassDescription extracts the class description
func extractClassDescription(lines []string) string {
	var descLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Stop at first table or major section
		if strings.HasPrefix(line, "|") || strings.HasPrefix(line, "###") || strings.HasPrefix(line, "####") {
			break
		}

		if line != "" {
			descLines = append(descLines, line)
		}
	}

	return strings.Join(descLines, "\n")
}
