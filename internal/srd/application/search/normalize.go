package search

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// normalizer strips combining marks (e.g. accents) so that "perché" and
// "perche" compare equal. Italian search frequently sees both spellings.
var normalizer = transform.Chain(
	norm.NFD,
	runes.Remove(runes.In(unicode.Mn)),
	norm.NFC,
)

// normalize lowercases and removes diacritics. Empty strings pass through.
func normalize(s string) string {
	if s == "" {
		return ""
	}
	out, _, err := transform.String(normalizer, s)
	if err != nil {
		return strings.ToLower(s)
	}
	return strings.ToLower(out)
}

// tokenize splits a normalized query into whitespace-separated tokens,
// dropping empty entries. Caller should normalize first.
func tokenize(s string) []string {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || r == '\'' || r == '’'
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}
