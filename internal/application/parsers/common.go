package parsers

import (
	"strings"
)

// Item represents a parsed item with title and content
type Item struct {
	title string
	lines []string
}

// splitItemsByH2 splits lines into items by H2 headers
func splitItemsByH2(lines []string) []Item {
	var items []Item
	var currentItem *Item

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for H2 header
		if strings.HasPrefix(line, "## ") {
			// Save previous item
			if currentItem != nil && len(currentItem.lines) > 0 {
				items = append(items, *currentItem)
			}

			// Start new item
			title := strings.TrimPrefix(line, "## ")
			currentItem = &Item{
				title: title,
				lines: []string{},
			}
		} else if currentItem != nil {
			// Add line to current item
			if line != "" {
				currentItem.lines = append(currentItem.lines, line)
			}
		}
	}

	// Add final item
	if currentItem != nil && len(currentItem.lines) > 0 {
		items = append(items, *currentItem)
	}

	return items
}

// collectLabeledFieldsFromLines extracts labeled fields from content lines
func collectLabeledFieldsFromLines(lines []string) map[string]string {
	fields := make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for "Label: Value" format
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				label := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Remove all markdown formatting from label (** or *)
				label = strings.Trim(label, "*")
				label = strings.TrimSpace(label)

				// Clean value - remove any leading markdown that might have leaked
				value = strings.TrimLeft(value, "* ")
				value = strings.TrimSpace(value)

				if label != "" && value != "" {
					fields[label] = value
				}
			}
		}
	}

	return fields
}
