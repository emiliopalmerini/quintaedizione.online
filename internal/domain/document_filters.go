package domain

// DocumentFilters contains metadata for querying and filtering documents
// Common filter keys:
// - "collection": collection name (animali, armi, mostri, etc.)
// - "type": document type (Bestia, Arma Semplice, etc.)
// - "rarity": for magic items (Comune, Non Comune, Raro, etc.)
// - "level": for spells (0-9)
// - "cr": challenge rating for monsters
// - "category": category/subcategory
// - "source_file": original markdown file
// - "locale": content locale (always "ita")
type DocumentFilters map[string]any

// NewDocumentFilters creates a new DocumentFilters
func NewDocumentFilters() DocumentFilters {
	return make(DocumentFilters)
}

// Set adds or updates a filter
func (f DocumentFilters) Set(key string, value any) {
	f[key] = value
}

// Get retrieves a filter value
func (f DocumentFilters) Get(key string) (any, bool) {
	val, ok := f[key]
	return val, ok
}

// GetString retrieves a string filter value
func (f DocumentFilters) GetString(key string) string {
	if val, ok := f[key].(string); ok {
		return val
	}
	return ""
}

// GetInt retrieves an int filter value
func (f DocumentFilters) GetInt(key string) int {
	if val, ok := f[key].(int); ok {
		return val
	}
	return 0
}

// GetFloat retrieves a float64 filter value
func (f DocumentFilters) GetFloat(key string) float64 {
	if val, ok := f[key].(float64); ok {
		return val
	}
	return 0.0
}

// GetBool retrieves a bool filter value
func (f DocumentFilters) GetBool(key string) bool {
	if val, ok := f[key].(bool); ok {
		return val
	}
	return false
}

// Has checks if a filter key exists
func (f DocumentFilters) Has(key string) bool {
	_, ok := f[key]
	return ok
}

// Delete removes a filter
func (f DocumentFilters) Delete(key string) {
	delete(f, key)
}
