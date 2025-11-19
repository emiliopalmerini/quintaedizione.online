package mappers

// GetString extracts a string value from a map with a default fallback
func GetString(m map[string]any, key string, defaultValue string) string {
	if m == nil {
		return defaultValue
	}
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultValue
}

// GetInt64 extracts an int64 value from a map with a default fallback
func GetInt64(m map[string]any, key string, defaultValue int64) int64 {
	if m == nil {
		return defaultValue
	}
	if val, ok := m[key].(int64); ok {
		return val
	}
	return defaultValue
}

// GetBool extracts a bool value from a map with a default fallback
func GetBool(m map[string]any, key string, defaultValue bool) bool {
	if m == nil {
		return defaultValue
	}
	if val, ok := m[key].(bool); ok {
		return val
	}
	return defaultValue
}

// GetSlice extracts a slice of any type from a map with a default fallback
func GetSlice(m map[string]any, key string, defaultValue []any) []any {
	if m == nil {
		return defaultValue
	}
	if val, ok := m[key].([]any); ok {
		return val
	}
	return defaultValue
}

// GetMap extracts a nested map from a map with a default fallback
func GetMap(m map[string]any, key string, defaultValue map[string]any) map[string]any {
	if m == nil {
		return defaultValue
	}
	if val, ok := m[key].(map[string]any); ok {
		return val
	}
	return defaultValue
}
