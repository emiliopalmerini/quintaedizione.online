package filters

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/filters"
)

// PredicateBuilder builds in-memory document predicates from filter definitions.
//
// All filters share the same semantics: the user-supplied value is split on
// commas into one or more needles; a document matches when any field value
// equals any needle (case-insensitive). The document field may be a scalar or
// a slice; for slices any element counts as a candidate.
type PredicateBuilder struct{}

func NewPredicateBuilder() *PredicateBuilder { return &PredicateBuilder{} }

// BuildPredicate converts a FilterSet into a DocumentPredicate. All filters
// must match (AND); within a single filter, comma-separated values are ORed.
func (b *PredicateBuilder) BuildPredicate(filterSet *filters.FilterSet) (filters.DocumentPredicate, error) {
	if !filterSet.HasFilters() {
		return nil, nil
	}

	var predicates []filters.DocumentPredicate
	for _, fv := range filterSet.Filters {
		if p := buildFieldPredicate(fv.Definition.FieldPath, fv.Value); p != nil {
			predicates = append(predicates, p)
		}
	}

	switch len(predicates) {
	case 0:
		return nil, nil
	case 1:
		return predicates[0], nil
	}
	return func(doc map[string]any) bool {
		for _, p := range predicates {
			if !p(doc) {
				return false
			}
		}
		return true
	}, nil
}

// BuildSearchPredicate returns a predicate that does case-insensitive substring
// match on title.
func (b *PredicateBuilder) BuildSearchPredicate(searchTerm string) filters.DocumentPredicate {
	if searchTerm == "" {
		return nil
	}
	needle := strings.ToLower(strings.TrimSpace(searchTerm))
	return func(doc map[string]any) bool {
		title, _ := doc["title"].(string)
		return strings.Contains(strings.ToLower(title), needle)
	}
}

func buildFieldPredicate(fieldPath, rawValue string) filters.DocumentPredicate {
	needles := splitAndLower(rawValue)
	if len(needles) == 0 {
		return nil
	}
	return func(doc map[string]any) bool {
		for _, v := range fieldValues(doc, fieldPath) {
			lv := strings.ToLower(v)
			for _, n := range needles {
				if lv == n {
					return true
				}
			}
		}
		return false
	}
}

// splitAndLower splits a comma-separated value into lowercased non-empty
// needles.
func splitAndLower(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.ToLower(strings.TrimSpace(p)); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// fieldValues resolves a dot-separated path to one or more string values. A
// missing field yields nothing; a slice yields each element formatted as a
// string.
func fieldValues(doc map[string]any, path string) []string {
	cur := any(doc)
	for _, part := range strings.Split(path, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = m[part]
	}
	return toStrings(cur)
}

func toStrings(v any) []string {
	switch x := v.(type) {
	case nil:
		return nil
	case []any:
		out := make([]string, 0, len(x))
		for _, e := range x {
			if s := fmt.Sprintf("%v", e); s != "" {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return x
	default:
		s := fmt.Sprintf("%v", v)
		if s == "" {
			return nil
		}
		return []string{s}
	}
}
