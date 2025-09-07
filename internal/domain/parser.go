package domain

// ParserRepository interface for parser data operations
type ParserRepository interface {
	UpsertMany(collection string, uniqueFields []string, docs []map[string]any) (int, error)
	Count(collection string) (int64, error)
}
