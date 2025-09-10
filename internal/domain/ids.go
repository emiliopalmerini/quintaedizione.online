package domain

import (
	"regexp"
	"strings"
)

// ---------- ID Types for all domain entities ----------

type (
	// Core entities - using business identifiers (slugs)
	AnimaleSlug            Slug
	ArmaSlug               Slug
	ArmaturaSlug           Slug
	BackgroundSlug         Slug
	CavalcaturaVeicoloSlug Slug
	ClasseSlug             Slug
	DocumentoSlug          Slug
	EquipaggiamentoSlug    Slug
	IncantesimoSlug        Slug
	MostroSlug             Slug
	OggettoMagicoSlug      Slug
	RegolaSlug             Slug
	ServizioSlug           Slug
	SpecieSlug             Slug
	StrumentoSlug          Slug
	TalentoSlug            Slug

	// Related entities - using business identifiers
	ScuolaIncantesimoSlug Slug
	CaratteristicaSlug    Slug
	AbilitaSlug           Slug
	DannoSlug             Slug
)

// NormalizeID normalizes a string to be used as an ID
func NormalizeID(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces and special characters with underscores
	re := regexp.MustCompile(`[^\w\d]+`)
	s = re.ReplaceAllString(s, "_")

	// Remove leading/trailing underscores
	s = strings.Trim(s, "_")

	// Replace multiple underscores with single
	re = regexp.MustCompile(`_{2,}`)
	s = re.ReplaceAllString(s, "_")

	return s
}
