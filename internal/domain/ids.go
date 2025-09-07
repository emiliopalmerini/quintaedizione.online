package domain

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// ---------- ID Types for all domain entities ----------

type (
	// Core entities
	AnimaleID            uuid.UUID
	ArmaID               uuid.UUID
	ArmaturaID           uuid.UUID
	BackgroundID         uuid.UUID
	CavalcaturaVeicoloID uuid.UUID
	ClasseID             uuid.UUID
	DocumentoID          uuid.UUID
	EquipaggiamentoID    uuid.UUID
	IncantesimoID        uuid.UUID
	MostroID             uuid.UUID
	OggettoMagicoID      uuid.UUID
	ServizioID           uuid.UUID
	SpecieID             uuid.UUID
	StrumentoID          uuid.UUID
	TalentoID            uuid.UUID

	// Related entities
	ScuolaIncantesimoID uuid.UUID
	CaratteristicaID    uuid.UUID
	AbilitaID           uuid.UUID
	DannoID             uuid.UUID
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
