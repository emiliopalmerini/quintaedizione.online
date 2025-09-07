package domain

import "github.com/google/uuid"

// ---------- Enum / VO di supporto ----------

// Unità di capacità
type UnitaCapacita string

const (
	UnitaLitri      UnitaCapacita = "l"
	UnitaGalloni    UnitaCapacita = "gal"
	UnitaMillilitri UnitaCapacita = "ml"
)

// Capacità
type Capacita struct {
	Valore float64       `json:"valore" bson:"valore"`
	Unita  UnitaCapacita `json:"unita"  bson:"unita"`
}

// ---------- Entità ----------

type Equipaggiamento struct {
	ID        EquipaggiamentoID `json:"id"        bson:"_id"`
	Slug      Slug              `json:"slug"      bson:"slug"`
	Nome      string            `json:"nome"      bson:"nome"`
	Costo     Costo             `json:"costo"     bson:"costo"`
	Peso      Peso              `json:"peso"      bson:"peso"`
	Capacita  *Capacita         `json:"capacita"  bson:"capacita,omitempty"` // opzionale per oggetti senza capacità
	Note      string            `json:"note"      bson:"note"`
	Contenuto string            `json:"contenuto" bson:"contenuto"`
}

// ---------- Costruttore ----------

func NewEquipaggiamento(
	id uuid.UUID,
	nome string,
	costo Costo,
	peso Peso,
	capacita *Capacita,
	note string,
	contenuto string,
) *Equipaggiamento {
	slug, _ := NewSlug(nome)

	return &Equipaggiamento{
		ID:        EquipaggiamentoID(id),
		Slug:      slug,
		Nome:      nome,
		Costo:     costo,
		Peso:      peso,
		Capacita:  capacita,
		Note:      note,
		Contenuto: contenuto,
	}
}

// EntityType implements ParsedEntity interface
func (e *Equipaggiamento) EntityType() string {
	return "equipaggiamento"
}
