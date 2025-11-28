package mappers

func GetString(m map[string]any, key string, defaultValue string) string {
	if m == nil {
		return defaultValue
	}
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultValue
}

func GetInt64(m map[string]any, key string, defaultValue int64) int64 {
	if m == nil {
		return defaultValue
	}
	if val, ok := m[key].(int64); ok {
		return val
	}
	return defaultValue
}

func GetBool(m map[string]any, key string, defaultValue bool) bool {
	if m == nil {
		return defaultValue
	}
	if val, ok := m[key].(bool); ok {
		return val
	}
	return defaultValue
}

func GetSlice(m map[string]any, key string, defaultValue []any) []any {
	if m == nil {
		return defaultValue
	}
	if val, ok := m[key].([]any); ok {
		return val
	}
	return defaultValue
}

func GetMap(m map[string]any, key string, defaultValue map[string]any) map[string]any {
	if m == nil {
		return defaultValue
	}
	if val, ok := m[key].(map[string]any); ok {
		return val
	}
	return defaultValue
}
