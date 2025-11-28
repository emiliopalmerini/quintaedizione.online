package parsers

import (
	"regexp"

	"github.com/emiliopalmerini/quintaedizione.online/internal/infrastructure"
)

type LanguageCode string

const (
	Italian LanguageCode = "ita"
	English LanguageCode = "eng"
)

type LanguageConfig struct {
	DataPath         string                    `yaml:"data_path"`
	SectionDelimiter string                    `yaml:"section_delimiter"`
	FieldMappings    map[string]string         `yaml:"field_mappings"`
	Patterns         map[string]*regexp.Regexp `yaml:"-"`
	PatternStrings   map[string]string         `yaml:"patterns"`
	RequiredFields   map[string][]string       `yaml:"required_fields"`
}

type ParserFunc func([]string) ([]map[string]any, error)

type WorkItem struct {
	Filename   string       `json:"filename"`
	Collection string       `json:"collection"`
	Language   LanguageCode `json:"language"`
}

type IngestResult struct {
	Collection string `json:"collection"`
	Filename   string `json:"filename"`
	Parsed     int    `json:"parsed"`
	Written    int    `json:"written"`
	Error      string `json:"error,omitempty"`
	Preview    string `json:"preview,omitempty"`
}

type LegacyParsingContext struct {
	Filename   string
	Collection string
	Language   string
	Source     string
}

func NewIngestResult(collection, filename string) *IngestResult {
	return &IngestResult{
		Collection: collection,
		Filename:   filename,
		Parsed:     0,
		Written:    0,
	}
}

func NewLegacyParsingContext(filename, collection string) *LegacyParsingContext {
	return &LegacyParsingContext{
		Filename:   filename,
		Collection: collection,
		Language:   infrastructure.ExtractLanguageFromPath(filename),
	}
}

func (r *IngestResult) SetError(err error) {
	if err != nil {
		r.Error = err.Error()
	}
}

func (r *IngestResult) SetPreview(preview string) {
	r.Preview = preview
}

func (r *IngestResult) HasError() bool {
	return r.Error != ""
}
