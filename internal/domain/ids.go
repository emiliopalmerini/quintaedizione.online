package domain

import "github.com/google/uuid"

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
