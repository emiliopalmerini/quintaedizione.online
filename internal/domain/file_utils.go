package domain

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ReadLines reads all lines from a file
func ReadLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filename, err)
	}

	return lines, nil
}

// WriteLines writes lines to a file
func WriteLines(filename string, lines []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("error writing to file %s: %w", filename, err)
		}
	}

	return nil
}

// IsMarkdownFile checks if a file has markdown extension
func IsMarkdownFile(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), ".md")
}

// ExtractLanguageFromPath extracts language code from file path
func ExtractLanguageFromPath(path string) string {
	if strings.Contains(path, "/ita/") || strings.Contains(path, "\\ita\\") {
		return "it"
	}
	if strings.Contains(path, "/eng/") || strings.Contains(path, "\\eng\\") {
		return "en"
	}
	return "unknown"
}

// NormalizeID creates a normalized ID from a name
func NormalizeID(name string) string {
	// Convert to lowercase
	id := strings.ToLower(name)
	
	// Replace spaces and special characters with hyphens
	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ReplaceAll(id, "'", "")
	id = strings.ReplaceAll(id, "\"", "")
	id = strings.ReplaceAll(id, ".", "")
	id = strings.ReplaceAll(id, ",", "")
	id = strings.ReplaceAll(id, "(", "")
	id = strings.ReplaceAll(id, ")", "")
	id = strings.ReplaceAll(id, "[", "")
	id = strings.ReplaceAll(id, "]", "")
	id = strings.ReplaceAll(id, "/", "-")
	id = strings.ReplaceAll(id, "\\", "-")
	
	// Replace multiple hyphens with single hyphen
	for strings.Contains(id, "--") {
		id = strings.ReplaceAll(id, "--", "-")
	}
	
	// Trim hyphens from start and end
	id = strings.Trim(id, "-")
	
	return id
}

// CleanText removes extra whitespace and normalizes text
func CleanText(text string) string {
	// Replace multiple whitespace with single space
	lines := strings.Split(text, "\n")
	var cleanLines []string
	
	for _, line := range lines {
		// Trim each line
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}
	
	return strings.Join(cleanLines, "\n")
}

// ExtractFirstLine extracts the first non-empty line from text
func ExtractFirstLine(text string) string {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

// RemoveMarkdownHeaders removes markdown header syntax from text
func RemoveMarkdownHeaders(text string) string {
	text = strings.TrimSpace(text)
	
	// Remove ### ## # headers
	for strings.HasPrefix(text, "#") {
		text = strings.TrimPrefix(text, "#")
		text = strings.TrimSpace(text)
	}
	
	return text
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// EnsureDir ensures a directory exists
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}