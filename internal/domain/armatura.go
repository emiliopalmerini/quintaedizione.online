package domain

import "github.com/google/uuid"

// ---------- Enum / VO di supporto ----------

// Categoria armature
type CategoriaArmatura string

const (
	CategoriaArmaturaLeggera CategoriaArmatura = "Leggera"
	CategoriaArmaturaMedia   CategoriaArmatura = "Media"
	CategoriaArmaturaPesante CategoriaArmatura = "Pesante"
	CategoriaScudo           CategoriaArmatura = "Scudo"
)

// Classe Armatura
type ClasseArmatura struct {
	Base            int  `json:"base"             bson:"base"`
	ModificatoreDes bool `json:"modificatore_des" bson:"modificatore_des"`
	LimiteDes       int  `json:"limite_des"       bson:"limite_des,omitempty"` // null se non c'è limite
}

// ---------- Entità ----------

type Armatura struct {
	ID                  ArmaturaID        `json:"id"                    bson:"_id"`
	Slug                Slug              `json:"slug"                  bson:"slug"`
	Nome                string            `json:"nome"                  bson:"nome"`
	Costo               Costo             `json:"costo"                 bson:"costo"`
	Peso                Peso              `json:"peso"                  bson:"peso"`
	Categoria           CategoriaArmatura `json:"categoria"             bson:"categoria"`
	ClasseArmatura      ClasseArmatura    `json:"classe_armatura"       bson:"classe_armatura"`
	ForzaRichiesta      int               `json:"forza_richiesta"       bson:"forza_richiesta,omitempty"` // null se non richiesta
	SvantaggioFurtivita bool              `json:"svantaggio_furtivita"  bson:"svantaggio_furtivita"`
	Contenuto           string            `json:"contenuto"             bson:"contenuto"`
}

// ---------- Costruttore ----------

func NewArmatura(
	id uuid.UUID,
	nome string,
	costo Costo,
	peso Peso,
	categoria CategoriaArmatura,
	classeArmatura ClasseArmatura,
	forzaRichiesta int,
	svantaggioFurtivita bool,
	contenuto string,
) *Armatura {
	slug, _ := NewSlug(nome)

	return &Armatura{
		ID:                  ArmaturaID(id),
		Slug:                slug,
		Nome:                nome,
		Costo:               costo,
		Peso:                peso,
		Categoria:           categoria,
		ClasseArmatura:      classeArmatura,
		ForzaRichiesta:      forzaRichiesta,
		SvantaggioFurtivita: svantaggioFurtivita,
		Contenuto:           contenuto,
	}
}
