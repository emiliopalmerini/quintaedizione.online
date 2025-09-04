package domain

import (
	"path/filepath"
	"time"
)

// WorkItem represents a parsing task
type WorkItem struct {
	Filename   string
	Collection string
	Parser     ParserFunc
}

// ParserFunc represents a parser function signature
type ParserFunc func(lines []string) ([]map[string]interface{}, error)

// IngestResult represents the result of an ingestion operation
type IngestResult struct {
	Collection string
	Filename   string
	Parsed     int
	Written    int
	Preview    *string
	Error      *string
}

// NewIngestResult creates a new IngestResult
func NewIngestResult(collection, filename string) *IngestResult {
	return &IngestResult{
		Collection: collection,
		Filename:   filepath.Base(filename),
		Parsed:     0,
		Written:    0,
	}
}

// SetError sets an error on the result
func (r *IngestResult) SetError(err error) {
	errStr := err.Error()
	r.Error = &errStr
}

// SetPreview sets a preview on the result
func (r *IngestResult) SetPreview(preview string) {
	r.Preview = &preview
}

// Repository interface for parser operations
type ParserRepository interface {
	UpsertMany(collection string, uniqueFields []string, docs []map[string]interface{}) (int, error)
}

// IngestRunner handles the ingestion process
type IngestRunner struct {
	baseDir    string
	repository ParserRepository
}

// NewIngestRunner creates a new ingest runner
func NewIngestRunner(baseDir string, repository ParserRepository) *IngestRunner {
	return &IngestRunner{
		baseDir:    baseDir,
		repository: repository,
	}
}

// RunIngest executes the ingestion process
func (r *IngestRunner) RunIngest(workItems []WorkItem, dryRun bool) ([]*IngestResult, error) {
	var results []*IngestResult

	for _, item := range workItems {
		result := r.processWorkItem(item, dryRun)
		results = append(results, result)
	}

	return results, nil
}

// processWorkItem processes a single work item
func (r *IngestRunner) processWorkItem(item WorkItem, dryRun bool) *IngestResult {
	result := NewIngestResult(item.Collection, item.Filename)

	// Read file
	fullPath := filepath.Join(r.baseDir, item.Filename)
	lines, err := ReadLines(fullPath)
	if err != nil {
		result.SetError(err)
		return result
	}

	// Parse content
	docs, err := item.Parser(lines)
	if err != nil {
		result.SetError(err)
		return result
	}

	result.Parsed = len(docs)

	if dryRun || r.repository == nil {
		// Generate preview for dry run
		preview := r.generatePreview(docs)
		result.SetPreview(preview)
		result.Written = 0
	} else {
		// Write to repository
		uniqueFields := GetUniqueFieldsForCollection(item.Collection)
		written, err := r.repository.UpsertMany(item.Collection, uniqueFields, docs)
		if err != nil {
			result.SetError(err)
			return result
		}
		result.Written = written
	}

	return result
}

// generatePreview generates a preview of the parsed documents
func (r *IngestRunner) generatePreview(docs []map[string]interface{}) string {
	previewKeys := []string{"nome", "titolo", "livello", "rarita", "tipo", "scuola"}
	
	var preview []map[string]interface{}
	maxPreview := 5
	if len(docs) < maxPreview {
		maxPreview = len(docs)
	}

	for i := 0; i < maxPreview; i++ {
		doc := docs[i]
		previewDoc := make(map[string]interface{})
		
		for _, key := range previewKeys {
			if value, exists := doc[key]; exists {
				previewDoc[key] = value
			}
		}
		
		if len(previewDoc) > 0 {
			preview = append(preview, previewDoc)
		}
	}

	// Convert to JSON-like string (simplified)
	return "Preview generated" // TODO: implement proper JSON marshaling
}

// GetUniqueFieldsForCollection returns unique fields for a collection
// Uses "slug" field for compatibility with existing Python schema
func GetUniqueFieldsForCollection(collection string) []string {
	// All collections use "slug" as the primary unique field for compatibility
	return []string{"slug"}
}

// ParsedDocument represents a parsed document with metadata
type ParsedDocument struct {
	ID          string                 `json:"id"`
	Collection  string                 `json:"collection"`
	Source      string                 `json:"source"`
	Language    string                 `json:"language"`
	ParsedAt    time.Time              `json:"parsed_at"`
	Content     map[string]interface{} `json:"content"`
}

// NewParsedDocument creates a new parsed document
func NewParsedDocument(id, collection, source, language string, content map[string]interface{}) *ParsedDocument {
	return &ParsedDocument{
		ID:         id,
		Collection: collection,
		Source:     source,
		Language:   language,
		ParsedAt:   time.Now(),
		Content:    content,
	}
}

// ParsingContext holds context information during parsing
type ParsingContext struct {
	Filename   string
	Collection string
	Language   string
	Source     string
	DryRun     bool
}

// NewParsingContext creates a new parsing context
func NewParsingContext(filename, collection, language, source string, dryRun bool) *ParsingContext {
	return &ParsingContext{
		Filename:   filename,
		Collection: collection,
		Language:   language,
		Source:     source,
		DryRun:     dryRun,
	}
}