package parsers

import "errors"

// Parsing errors
var (
	ErrEmptyContent        = errors.New("content is empty")
	ErrInvalidContentType  = errors.New("invalid content type")
	ErrParserAlreadyExists = errors.New("parser already exists")
	ErrParserNotFound      = errors.New("parser not found")
	ErrMissingSectionTitle = errors.New("section missing title")
	ErrEmptySectionContent = errors.New("section has no content")
	ErrInvalidContext      = errors.New("invalid parsing context")
	ErrBuilderNotReady     = errors.New("document builder not properly initialized")
)
