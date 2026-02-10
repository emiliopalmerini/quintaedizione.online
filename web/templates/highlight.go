package templates

import (
	"html"
	"strings"
)

// HighlightQuery wraps the first case-insensitive occurrence of query in
// the title with <mark> tags. The title is HTML-escaped first for safety.
func HighlightQuery(title, query string) string {
	escaped := html.EscapeString(title)
	if query == "" || title == "" {
		return escaped
	}

	escapedQuery := html.EscapeString(query)
	lower := strings.ToLower(escaped)
	lowerQuery := strings.ToLower(escapedQuery)

	idx := strings.Index(lower, lowerQuery)
	if idx < 0 {
		return escaped
	}

	// Preserve original case from escaped title
	matched := escaped[idx : idx+len(escapedQuery)]
	return escaped[:idx] + "<mark>" + matched + "</mark>" + escaped[idx+len(escapedQuery):]
}
