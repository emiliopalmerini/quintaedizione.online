package parsers

import "github.com/emiliopalmerini/due-draghi-5e-srd/internal/infrastructure"

// ParserFunc represents a function that parses markdown lines into documents
type ParserFunc func([]string) ([]map[string]any, error)

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

// LegacyParsingContext holds context information during parsing (deprecated)
type LegacyParsingContext struct {
	Filename   string
	Collection string
	Language   string
	Source     string
}

// NewIngestResult creates a new IngestResult
func NewIngestResult(collection, filename string) *IngestResult {
	return &IngestResult{
		Collection: collection,
		Filename:   filename,
		Parsed:     0,
		Written:    0,
	}
}

// NewLegacyParsingContext creates a new LegacyParsingContext
func NewLegacyParsingContext(filename, collection string) *LegacyParsingContext {
	return &LegacyParsingContext{
		Filename:   filename,
		Collection: collection,
		Language:   infrastructure.ExtractLanguageFromPath(filename),
	}
}

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
