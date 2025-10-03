package parsers

type ParsingContext struct {
	Filename string
	Language string
	Metadata map[string]any
	Logger   Logger
}

func NewParsingContext(filename, language string) *ParsingContext {
	return &ParsingContext{
		Filename: filename,
		Language: language,
		Metadata: make(map[string]any),
		Logger:   &NoOpLogger{},
	}
}

func (pc *ParsingContext) WithLogger(logger Logger) *ParsingContext {
	pc.Logger = logger
	return pc
}

func (pc *ParsingContext) WithMetadata(key string, value interface{}) *ParsingContext {
	pc.Metadata[key] = value
	return pc
}

func (pc *ParsingContext) Validate() error {
	if pc.Filename == "" {
		return ErrInvalidContext
	}
	if pc.Language == "" {
		return ErrInvalidContext
	}
	return nil
}
