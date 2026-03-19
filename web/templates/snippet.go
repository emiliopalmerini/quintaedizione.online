package templates

import (
	"regexp"
	"strings"
)

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)
var markdownRe = regexp.MustCompile(`[*_~` + "`" + `#\[\]]+`)

// ExtractSnippet returns a plain-text snippet of approximately maxLen characters
// centered around the first case-insensitive occurrence of query in body.
// HTML tags and markdown formatting are stripped before extraction.
// Returns "" if there is no match, or if body/query is empty.
func ExtractSnippet(body, query string, maxLen int) string {
	if body == "" || query == "" {
		return ""
	}

	// Strip HTML and markdown
	plain := htmlTagRe.ReplaceAllString(body, "")
	plain = markdownRe.ReplaceAllString(plain, "")
	plain = strings.Join(strings.Fields(plain), " ") // normalize whitespace

	lower := strings.ToLower(plain)
	lowerQuery := strings.ToLower(query)

	idx := strings.Index(lower, lowerQuery)
	if idx < 0 {
		return ""
	}

	// Calculate window around match
	half := (maxLen - len(query)) / 2
	if half < 0 {
		half = 0
	}

	start := idx - half
	end := idx + len(query) + half

	if start < 0 {
		start = 0
	}
	if end > len(plain) {
		end = len(plain)
	}

	snippet := plain[start:end]

	// Trim to word boundaries
	if start > 0 {
		if spaceIdx := strings.Index(snippet, " "); spaceIdx >= 0 && spaceIdx < len(query) {
			snippet = snippet[spaceIdx+1:]
		}
		snippet = "…" + snippet
	}
	if end < len(plain) {
		if spaceIdx := strings.LastIndex(snippet, " "); spaceIdx > len(snippet)-len(query) && spaceIdx >= 0 {
			snippet = snippet[:spaceIdx]
		}
		snippet = snippet + "…"
	}

	return snippet
}
