package domain

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type Slug string

func NewSlug(value string) (Slug, error) {
	if strings.TrimSpace(value) == "" {
		return "", NewDocumentError("create_slug", "", ErrInvalidDocumentTitle, "slug source cannot be empty")
	}
	return slugify(value), nil
}

var (
	nonAlnum  = regexp.MustCompile(`[^a-z0-9\-]+`)
	multiDash = regexp.MustCompile(`-+`)
)

func slugify(value string) Slug {
	s := strings.ToLower(strings.TrimSpace(value))

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
