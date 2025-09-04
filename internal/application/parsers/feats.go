package parsers

import (
	"strings"
)

// ParseFeats parses Italian D&D 5e feat data from markdown
func ParseFeats(lines []string) ([]map[string]interface{}, error) {
	items := splitItemsByH2(lines)
	var feats []map[string]interface{}

	for _, item := range items {
		if len(item.lines) == 0 {
			continue
		}

		feat := parseFeatItem(item.title, item.lines)
		if feat != nil {
			feats = append(feats, feat)
		}
	}

	return feats, nil
}

// parseFeatItem parses a single feat
func parseFeatItem(title string, lines []string) map[string]interface{} {
	name := strings.TrimSpace(title)
	if name == "" {
		return nil
	}

	// Extract prerequisite if present
	var prerequisito string
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "prerequisito") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 1 {
				prerequisito = strings.TrimSpace(parts[1])
			}
			break
		}
	}

	// Build feat object
	feat := map[string]interface{}{
		"slug":               name,
		"nome":               name,
		"contenuto_markdown": strings.Join(append([]string{"## " + title}, lines...), "\n"),
		"fonte":              "SRD",
		"versione":           "1.0",
	}

	if prerequisito != "" {
		feat["prerequisito"] = prerequisito
	}

	// Extract description (all text except prerequisites)
	description := extractFeatDescription(lines)
	if description != "" {
		feat["descrizione"] = description
	}

	return feat
}

// extractFeatDescription extracts feat description from lines
func extractFeatDescription(lines []string) string {
	var descLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip prerequisite lines
		if strings.Contains(strings.ToLower(line), "prerequisito") {
			continue
		}

		if line != "" {
			descLines = append(descLines, line)
		}
	}

	return strings.Join(descLines, "\n")
}
