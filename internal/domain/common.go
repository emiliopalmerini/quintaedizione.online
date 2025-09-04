package domain

// ---------- Value Objects comuni ----------

// Unità di velocità
type UnitaVelocita string

const (
	UnitaMetriRound UnitaVelocita = "m/round"
	UnitaMetriTurno UnitaVelocita = "m/turno"
	UnitaKmOra      UnitaVelocita = "km/h"
	UnitaNodi       UnitaVelocita = "nodi"
	UnitaMetri      UnitaVelocita = "m"
)

// Velocità unificata
type Velocita struct {
	Valore int           `json:"valore" bson:"valore"`
	Unita  UnitaVelocita `json:"unita"  bson:"unita"`
}

// Costruttore per Velocita
func NewVelocita(valore int, unita UnitaVelocita) Velocita {
	return Velocita{
		Valore: valore,
		Unita:  unita,
	}
}

// Punti Ferita unificati
type PuntiFerita int

