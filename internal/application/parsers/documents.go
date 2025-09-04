package parsers

import (
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// ParseDocumentWithFilename creates a parser for a specific document
func ParseDocument(filename string) domain.ParserFunc {
	return func(lines []string) ([]map[string]interface{}, error) {
		return parseDocument(lines, filename)
	}
}

// parseDocument parses a document into a single document entry
func parseDocument(lines []string, filename string) ([]map[string]interface{}, error) {
	if len(lines) == 0 {
		return []map[string]interface{}{}, nil
	}

	// Extract title from first heading or use filename
	title := extractDocumentTitle(lines)
	if title == "" {
		title = strings.TrimSuffix(filename, ".md")
	}

	// Join all content
	content := strings.Join(lines, "\n")
	plainText := extractPlainText(content)

	// Create document
	doc := map[string]interface{}{
		"titolo":             title,
		"slug":               domain.NormalizeID(title), // Use slug for MongoDB compatibility
		"filename":           filename,
		"contenuto_markdown": content,
		"contenuto_testo":    plainText,
		"numero_di_pagina":   extractPageNumber(filename),
		"fonte":              "SRD",
		"versione":           "1.0",
		"lingua":             domain.ExtractLanguageFromPath(filename),
	}

	return []map[string]interface{}{doc}, nil
}

// extractDocumentTitle extracts the title from document lines
func extractDocumentTitle(lines []string) string {
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for markdown headers
		if strings.HasPrefix(line, "# ") {
			return domain.RemoveMarkdownHeaders(line)
		}
		if strings.HasPrefix(line, "## ") {
			return domain.RemoveMarkdownHeaders(line)
		}

		// If we find content, stop looking
		if line != "" && !strings.HasPrefix(line, "#") {
			break
		}
	}

	return ""
}

// extractPlainText removes markdown formatting to create plain text
func extractPlainText(markdown string) string {
	text := markdown

	// Remove markdown headers
	lines := strings.Split(text, "\n")
	var cleanLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Remove markdown formatting
		line = domain.RemoveMarkdownHeaders(line)

		// Remove bold and italic formatting
		line = strings.ReplaceAll(line, "**", "")
		line = strings.ReplaceAll(line, "*", "")
		line = strings.ReplaceAll(line, "_", "")

		// Remove links but keep text
		// Simple regex-like replacement for [text](url) -> text
		for strings.Contains(line, "](") {
			start := strings.Index(line, "[")
			if start == -1 {
				break
			}
			middle := strings.Index(line[start:], "](")
			if middle == -1 {
				break
			}
			end := strings.Index(line[start+middle+2:], ")")
			if end == -1 {
				break
			}

			linkText := line[start+1 : start+middle]
			line = line[:start] + linkText + line[start+middle+2+end+1:]
		}

		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, "\n")
}

// extractPageNumber extracts page number from filename
func extractPageNumber(filename string) int {
	// Extract number from beginning of filename like "01_", "02_", etc.
	parts := strings.Split(filename, "_")
	if len(parts) > 0 {
		numStr := parts[0]
		// Remove path prefix
		pathParts := strings.Split(numStr, "/")
		if len(pathParts) > 0 {
			numStr = pathParts[len(pathParts)-1]
		}

		// Try to parse as number
		var pageNum int
		for _, char := range numStr {
			if char >= '0' && char <= '9' {
				pageNum = pageNum*10 + int(char-'0')
			} else {
				break
			}
		}
		return pageNum
	}

	return 0
}
