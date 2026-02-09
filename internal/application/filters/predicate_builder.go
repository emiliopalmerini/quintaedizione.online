package filters

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/filters"
)

// PredicateBuilder builds in-memory document predicates from filter definitions.
type PredicateBuilder struct{}

// NewPredicateBuilder creates a new PredicateBuilder.
func NewPredicateBuilder() *PredicateBuilder {
	return &PredicateBuilder{}
}

// BuildPredicate converts a FilterSet into a DocumentPredicate that tests documents in-memory.
func (b *PredicateBuilder) BuildPredicate(filterSet *filters.FilterSet) (filters.DocumentPredicate, error) {
	if !filterSet.HasFilters() {
		return nil, nil
	}

	var predicates []filters.DocumentPredicate

	for _, fv := range filterSet.Filters {
		p, err := b.buildSinglePredicate(fv)
		if err != nil {
			return nil, fmt.Errorf("failed to build predicate for %s: %w", fv.Definition.Name, err)
		}
		if p != nil {
			predicates = append(predicates, p)
		}
	}

	if len(predicates) == 0 {
		return nil, nil
	}
	if len(predicates) == 1 {
		return predicates[0], nil
	}

	// All predicates must match (AND)
	return func(doc map[string]any) bool {
		for _, p := range predicates {
			if !p(doc) {
				return false
			}
		}
		return true
	}, nil
}

func (b *PredicateBuilder) buildSinglePredicate(fv filters.FilterValue) (filters.DocumentPredicate, error) {
	def := fv.Definition
	value := fv.Value
	if value == "" {
		return nil, nil
	}

	switch def.Operator {
	case filters.ExactMatch:
		return b.buildExactMatch(def.FieldPath, value, def.DataType)
	case filters.RegexMatch:
		return b.buildRegexMatch(def.FieldPath, value)
	case filters.RangeMatch:
		return b.buildRangeMatch(def.FieldPath, value, def.DataType)
	case filters.InMatch:
		return b.buildInMatch(def.FieldPath, value)
	default:
		return nil, fmt.Errorf("unsupported operator: %d", def.Operator)
	}
}

func (b *PredicateBuilder) buildExactMatch(fieldPath, value string, dataType filters.FilterDataType) (filters.DocumentPredicate, error) {
	converted, err := filters.ConvertValue(value, dataType)
	if err != nil {
		return nil, err
	}
	return func(doc map[string]any) bool {
		return fmt.Sprintf("%v", getField(doc, fieldPath)) == fmt.Sprintf("%v", converted)
	}, nil
}

func (b *PredicateBuilder) buildRegexMatch(fieldPath, value string) (filters.DocumentPredicate, error) {
	// Multi-value: comma-separated values become OR patterns
	parts := strings.Split(value, ",")
	var patterns []*regexp.Regexp
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		compiled, err := regexp.Compile("(?i)" + regexp.QuoteMeta(p))
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %w", err)
		}
		patterns = append(patterns, compiled)
	}
	if len(patterns) == 0 {
		return nil, nil
	}
	return func(doc map[string]any) bool {
		field := getField(doc, fieldPath)
		if field == nil {
			return false
		}
		fieldStr := fmt.Sprintf("%v", field)
		for _, pattern := range patterns {
			if pattern.MatchString(fieldStr) {
				return true
			}
		}
		return false
	}, nil
}

func (b *PredicateBuilder) buildRangeMatch(fieldPath, value string, dataType filters.FilterDataType) (filters.DocumentPredicate, error) {
	if dataType != filters.NumberFilter {
		return nil, fmt.Errorf("range match only supported for number filters")
	}

	value = strings.TrimSpace(value)

	type numPred func(float64) bool

	var pred numPred

	if strings.HasPrefix(value, ">=") {
		n, err := strconv.ParseFloat(strings.TrimSpace(value[2:]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid range value: %s", value)
		}
		pred = func(v float64) bool { return v >= n }
	} else if strings.HasPrefix(value, "<=") {
		n, err := strconv.ParseFloat(strings.TrimSpace(value[2:]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid range value: %s", value)
		}
		pred = func(v float64) bool { return v <= n }
	} else if strings.HasPrefix(value, ">") {
		n, err := strconv.ParseFloat(strings.TrimSpace(value[1:]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid range value: %s", value)
		}
		pred = func(v float64) bool { return v > n }
	} else if strings.HasPrefix(value, "<") {
		n, err := strconv.ParseFloat(strings.TrimSpace(value[1:]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid range value: %s", value)
		}
		pred = func(v float64) bool { return v < n }
	} else if strings.Contains(value, "-") {
		parts := strings.Split(value, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid range format: %s", value)
		}
		min, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid range min value: %s", parts[0])
		}
		max, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid range max value: %s", parts[1])
		}
		pred = func(v float64) bool { return v >= min && v <= max }
	} else {
		n, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number value: %s", value)
		}
		pred = func(v float64) bool { return v == n }
	}

	return func(doc map[string]any) bool {
		field := getField(doc, fieldPath)
		if field == nil {
			return false
		}
		f, ok := toFloat64(field)
		if !ok {
			return false
		}
		return pred(f)
	}, nil
}

func (b *PredicateBuilder) buildInMatch(fieldPath, value string) (filters.DocumentPredicate, error) {
	values := strings.Split(value, ",")
	trimmed := make([]string, 0, len(values))
	for _, v := range values {
		if t := strings.TrimSpace(v); t != "" {
			trimmed = append(trimmed, t)
		}
	}
	if len(trimmed) == 0 {
		return nil, nil
	}

	return func(doc map[string]any) bool {
		field := getField(doc, fieldPath)
		if field == nil {
			return false
		}
		fieldStr := fmt.Sprintf("%v", field)
		for _, v := range trimmed {
			if fieldStr == v {
				return true
			}
		}
		return false
	}, nil
}

// BuildSearchPredicate returns a predicate that does case-insensitive substring match on title.
func (b *PredicateBuilder) BuildSearchPredicate(searchTerm string) filters.DocumentPredicate {
	if searchTerm == "" {
		return nil
	}
	lower := strings.ToLower(strings.TrimSpace(searchTerm))
	return func(doc map[string]any) bool {
		title, _ := doc["title"].(string)
		rawContent, _ := doc["raw_content"].(string)
		return strings.Contains(strings.ToLower(title), lower) ||
			strings.Contains(strings.ToLower(rawContent), lower)
	}
}

// getField retrieves a potentially nested field from a document.
// Supports dot-separated paths like "filters.scuola".
func getField(doc map[string]any, path string) any {
	parts := strings.Split(path, ".")
	var current any = doc
	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = m[part]
	}
	return current
}

// toFloat64 converts various numeric types to float64.
func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	default:
		return 0, false
	}
}
