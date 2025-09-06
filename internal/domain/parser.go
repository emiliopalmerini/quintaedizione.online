package domain

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ---------- Parser Types ----------

// ParserFunc represents a function that parses markdown lines into documents
type ParserFunc func([]string) ([]map[string]interface{}, error)

// WorkItem represents a parsing task
type WorkItem struct {
	Filename   string     `json:"filename"`
	Collection string     `json:"collection"`
	Parser     ParserFunc `json:"-"`
}

// IngestResult represents the result of processing a single work item
type IngestResult struct {
	Collection string `json:"collection"`
	Filename   string `json:"filename"`
	Parsed     int    `json:"parsed"`
	Written    int    `json:"written"`
	Error      string `json:"error,omitempty"`
	Preview    string `json:"preview,omitempty"`
}

// ParsingContext holds context information during parsing
type ParsingContext struct {
	Filename   string
	Collection string
	Language   string
	Source     string
}

// ParserRepository interface for parser data operations
type ParserRepository interface {
	UpsertMany(collection string, uniqueFields []string, docs []map[string]interface{}) (int, error)
	Count(collection string) (int64, error)
}

// ---------- Constructor Functions ----------

// NewIngestResult creates a new IngestResult
func NewIngestResult(collection, filename string) *IngestResult {
	return &IngestResult{
		Collection: collection,
		Filename:   filename,
		Parsed:     0,
		Written:    0,
	}
}

// NewParsingContext creates a new ParsingContext
func NewParsingContext(filename, collection string) *ParsingContext {
	return &ParsingContext{
		Filename:   filename,
		Collection: collection,
		Language:   ExtractLanguageFromPath(filename),
	}
}

// ---------- IngestResult Methods ----------

// SetError sets an error on the result
func (r *IngestResult) SetError(err error) {
	if err != nil {
		r.Error = err.Error()
	}
}

// SetPreview sets the preview content
func (r *IngestResult) SetPreview(preview string) {
	r.Preview = preview
}

// HasError returns true if the result has an error
func (r *IngestResult) HasError() bool {
	return r.Error != ""
}

// ---------- Utility Functions ----------

// FileExists checks if a file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// ReadLines reads all lines from a file
func ReadLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// ExtractLanguageFromPath extracts language code from file path
func ExtractLanguageFromPath(path string) string {
	dir := filepath.Dir(path)
	parts := strings.Split(dir, "/")

	for _, part := range parts {
		if part == "ita" || part == "eng" {
			return part
		}
	}

	// Default to Italian
	return "ita"
}

// NormalizeID normalizes a string to be used as an ID
func NormalizeID(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces and special characters with underscores
	re := regexp.MustCompile(`[^\w\d]+`)
	s = re.ReplaceAllString(s, "_")

	// Remove leading/trailing underscores
	s = strings.Trim(s, "_")

	// Replace multiple underscores with single
	re = regexp.MustCompile(`_{2,}`)
	s = re.ReplaceAllString(s, "_")

	return s
}

// RemoveMarkdownHeaders removes markdown headers from content
func RemoveMarkdownHeaders(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		// Skip lines that start with # (headers)
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// GetUniqueFieldsForCollection returns unique fields for MongoDB upsert operations
func GetUniqueFieldsForCollection(collection string) []string {
	switch collection {
	case "documenti":
		return []string{"slug"}
	case "incantesimi":
		return []string{"nome", "slug"}
	case "mostri":
		return []string{"nome", "slug"}
	case "classi":
		return []string{"nome", "slug"}
	case "backgrounds":
		return []string{"nome", "slug"}
	case "armi":
		return []string{"nome", "slug"}
	case "armature":
		return []string{"nome", "slug"}
	case "strumenti":
		return []string{"nome", "slug"}
	case "servizi":
		return []string{"nome", "slug"}
	case "equipaggiamento":
		return []string{"nome", "slug"}
	case "oggetti_magici":
		return []string{"nome", "slug"}
	case "talenti":
		return []string{"nome", "slug"}
	case "animali":
		return []string{"nome", "slug"}
	default:
		return []string{"slug"}
	}
}
