package domain

type DocumentFilters map[string]any

func NewDocumentFilters() DocumentFilters {
	return make(DocumentFilters)
}

func (f DocumentFilters) Set(key string, value any) {
	f[key] = value
}

func (f DocumentFilters) Get(key string) (any, bool) {
	val, ok := f[key]
	return val, ok
}

func (f DocumentFilters) GetString(key string) string {
	if val, ok := f[key].(string); ok {
		return val
	}
	return ""
}

func (f DocumentFilters) GetInt(key string) int {
	if val, ok := f[key].(int); ok {
		return val
	}
	return 0
}

func (f DocumentFilters) GetFloat(key string) float64 {
	if val, ok := f[key].(float64); ok {
		return val
	}
	return 0.0
}

func (f DocumentFilters) GetBool(key string) bool {
	if val, ok := f[key].(bool); ok {
		return val
	}
	return false
}

func (f DocumentFilters) Has(key string) bool {
	_, ok := f[key]
	return ok
}

func (f DocumentFilters) Delete(key string) {
	delete(f, key)
}
