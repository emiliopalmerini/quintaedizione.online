package domain

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type Slug string

func NewSlug(value string) (Slug, error) {
	if value == "" {
		return "", fmt.Errorf("slug cannot be empty")
	}
	return slugify(value), nil
}

var (
	nonAlnum  = regexp.MustCompile(`[^a-z0-9\-]+`)
	multiDash = regexp.MustCompile(`-+`)
)

// slugify converte una stringa in slug ASCII [a-z0-9-].
// - minuscole
// - accenti rimossi
// - spazi → trattini
// - caratteri invalidi eliminati
// - trattini multipli compressi
// - tratto iniziale/finale eliminato
// - fallback "n-a" se vuoto
func slugify(value string) Slug {
	s := strings.ToLower(strings.TrimSpace(value))

	// NFD → remove diacritics → NFC
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	if res, _, err := transform.String(t, s); err == nil {
		s = res
	}

	s = strings.ReplaceAll(s, " ", "-")
	s = nonAlnum.ReplaceAllString(s, "")
	s = multiDash.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")

	if s == "" {
		s = "n-a"
	}
	return Slug(s)
}
