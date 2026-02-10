package web

import (
	"regexp"
	"strings"
)

var (
	mdBoldItalic = regexp.MustCompile(`\*{1,3}([^*]+)\*{1,3}`)
	mdLink       = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	mdHeading    = regexp.MustCompile(`(?m)^#{1,6}\s+.*$`)
)

// truncateDescription cleans markdown formatting from raw text,
// extracts the first paragraph, and truncates at a word boundary.
func truncateDescription(raw string, maxLen int) string {
	if raw == "" {
		return ""
	}

	// Split into paragraphs (double newline) and take the first non-empty, non-heading one
	paragraphs := strings.Split(raw, "\n\n")
	text := ""
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// Skip headings
		if strings.HasPrefix(p, "#") {
			continue
		}
		text = p
		break
	}

	if text == "" {
		return ""
	}

	// Strip markdown formatting
	text = mdBoldItalic.ReplaceAllString(text, "$1")
	text = mdLink.ReplaceAllString(text, "$1")
	text = mdHeading.ReplaceAllString(text, "")
	text = strings.TrimSpace(text)

	if len(text) <= maxLen {
		return text
	}

	// Truncate at word boundary
	truncated := text[:maxLen]
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > 0 {
		truncated = truncated[:lastSpace]
	}
	return truncated + "..."
}
