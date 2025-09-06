package domain

import "github.com/google/uuid"

// ---------- Enum / VO di supporto ----------

// Tipo animale (semplificato rispetto ai mostri)
type TipoAnimale string

const (
	TipoA TipoAnimale = "Animale"
	TipoB TipoAnimale = "Bestia"
)

// ---------- Entità ----------

type Animale struct {
	ID              AnimaleID        `json:"id"              bson:"_id"`
	Slug            Slug             `json:"slug"            bson:"slug"`
	Nome            string           `json:"nome"            bson:"nome"`
	Taglia          Taglia           `json:"taglia"          bson:"taglia"`
	Tipo            TipoAnimale      `json:"tipo"            bson:"tipo"`
	ClasseArmatura  ClasseArmatura   `json:"ca"              bson:"ca"`
	PuntiFerita     PuntiFerita      `json:"pf"              bson:"pf"`
	Velocita        Velocita         `json:"velocita"        bson:"velocita"`
	Caratteristiche []Caratteristica `json:"caratteristiche" bson:"caratteristiche"`
	Tratti          []Tratto         `json:"tratti"          bson:"tratti"`
	Azioni          []Azione         `json:"azioni"          bson:"azioni"`
	Contenuto       string           `json:"contenuto"       bson:"contenuto"`
	// TODO: add to the parser
	BonusCompetenza BonusCompetenza `json:"bonus_competenza" bson:"bonus_competenza"`
}

// ---------- Costruttore ----------

func NewAnimale(
	id uuid.UUID,
	nome string,
	taglia Taglia,
	tipo TipoAnimale,
	ca ClasseArmatura,
	pf PuntiFerita,
	velocita Velocita,
	caratteristiche []Caratteristica,
	tratti []Tratto,
	azioni []Azione,
	contenuto string,
	bc BonusCompetenza,
) *Animale {
	slug, _ := NewSlug(nome)

	return &Animale{
		ID:              AnimaleID(id),
		Slug:            slug,
		Nome:            nome,
		Taglia:          taglia,
		Tipo:            tipo,
		ClasseArmatura:  ca,
		PuntiFerita:     pf,
		Velocita:        velocita,
		Caratteristiche: caratteristiche,
		Tratti:          tratti,
		Azioni:          azioni,
		Contenuto:       contenuto,
		BonusCompetenza: bc,
	}
}
