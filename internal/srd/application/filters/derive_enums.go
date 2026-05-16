package filters

import (
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/collections"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/filters"
)

// EnumSource returns observed field values for a given collection. The map key
// is the value (as formatted by the aggregator) and the value is the count.
type EnumSource interface {
	Aggregate(collection, fieldPath string, match func(map[string]any) bool) map[string]int64
}

type entry struct {
	key     string
	display string
	count   int64
}

// DeriveEnumValues replaces each filter definition's EnumValues with the
// distinct values observed in the loaded data. Filters with pre-set EnumValues
// (e.g., the edition filter) are left untouched. Values are case-folded for
// display so that data noise (mixed case, masc/fem variants) collapses to a
// single canonical option.
func DeriveEnumValues(registry *InMemoryFilterRegistry, source EnumSource) {
	for i, def := range registry.filters {
		if len(def.EnumValues) > 0 {
			continue
		}
		values := collectValues(def, source)
		registry.filters[i].EnumValues = values
	}
}

func collectValues(def filters.FilterDefinition, source EnumSource) []string {
	colls := def.Collections
	if len(colls) == 0 {
		colls = collections.GetAllCollections()
	}

	// Case-fold bucket: lower-cased key → canonical display label (most common
	// case form wins; ties broken alphabetically).
	type variant struct {
		display string
		count   int64
	}
	bucket := make(map[string]variant)
	for _, c := range colls {
		raw := source.Aggregate(c.String(), def.FieldPath, nil)
		for k, n := range raw {
			if k == "" {
				continue
			}
			key := foldKey(k)
			v := bucket[key]
			if n > v.count || (n == v.count && k < v.display) {
				v.display = k
			}
			v.count += n
			bucket[key] = v
		}
	}
	if len(bucket) == 0 {
		return nil
	}

	entries := make([]entry, 0, len(bucket))
	for k, v := range bucket {
		entries = append(entries, entry{key: k, display: v.display, count: v.count})
	}

	if allNumeric(entries) {
		sort.Slice(entries, func(i, j int) bool {
			return numericLess(entries[i].key, entries[j].key)
		})
	} else {
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].count != entries[j].count {
				return entries[i].count > entries[j].count
			}
			return entries[i].display < entries[j].display
		})
	}

	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = titleDisplay(e.display)
	}
	return out
}

// titleDisplay returns a display-friendly version of a value: each word's
// first letter is upper-cased, the rest is left as-is. Punctuation, fractions,
// and parenthetical qualifiers pass through unchanged.
func titleDisplay(s string) string {
	if s == "" {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := true
	for _, r := range s {
		if prevSpace && unicode.IsLetter(r) {
			b.WriteRune(unicode.ToUpper(r))
		} else {
			b.WriteRune(r)
		}
		prevSpace = unicode.IsSpace(r)
	}
	return b.String()
}

func foldKey(s string) string {
	// Lowercase only; trimming or punctuation-folding would over-collapse
	// distinct values like "1/2" vs "1".
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func allNumeric(entries []entry) bool {
	for _, e := range entries {
		if _, err := strconv.ParseFloat(e.key, 64); err != nil {
			// Allow fractional CR notation (e.g. "1/4").
			if !isFraction(e.key) {
				return false
			}
		}
	}
	return true
}

func numericLess(a, b string) bool {
	return fracValue(a) < fracValue(b)
}

func isFraction(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			_, eN := strconv.Atoi(s[:i])
			_, eD := strconv.Atoi(s[i+1:])
			return eN == nil && eD == nil
		}
	}
	return false
}

func fracValue(s string) float64 {
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			n, _ := strconv.ParseFloat(s[:i], 64)
			d, _ := strconv.ParseFloat(s[i+1:], 64)
			if d == 0 {
				return n
			}
			return n / d
		}
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
