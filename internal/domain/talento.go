package domain

import "github.com/google/uuid"

// ---------- Enum / VO di supporto ----------

// Categoria talento
type CategoriaTalento string

const (
	CategoriaTalentoOrigine  CategoriaTalento = "Talento di Origine"
	CategoriaTalentoGenerale CategoriaTalento = "Generale"
	CategoriaTalentoCombat   CategoriaTalento = "Combattimento"
	CategoriaTalentoMagia    CategoriaTalento = "Magia"
	CategoriaTalentoAbilita  CategoriaTalento = "Abilità"
	CategoriaTalentoRazziale CategoriaTalento = "Razziale"
)

// ---------- Entità ----------

type Talento struct {
	ID           TalentoID        `json:"id"           bson:"_id"`
	Slug         Slug             `json:"slug"         bson:"slug"`
	Nome         string           `json:"nome"         bson:"nome"`
	Categoria    CategoriaTalento `json:"categoria"    bson:"categoria"`
	Prerequisiti string           `json:"prerequisiti" bson:"prerequisiti"`
	Benefici     []string         `json:"benefici"     bson:"benefici"`
	Contenuto    string           `json:"contenuto"    bson:"contenuto"`
}

// ---------- Costruttore ----------

func NewTalento(
	id uuid.UUID,
	nome string,
	categoria CategoriaTalento,
	prerequisiti string,
	benefici []string,
	contenuto string,
) *Talento {
	slug, _ := NewSlug(nome)

	return &Talento{
		ID:           TalentoID(id),
		Slug:         slug,
		Nome:         nome,
		Categoria:    categoria,
		Prerequisiti: prerequisiti,
		Benefici:     benefici,
		Contenuto:    contenuto,
	}
}
