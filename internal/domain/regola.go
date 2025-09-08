package domain

import "github.com/google/uuid"

type Regola struct {
	ID        RegolaID `json:"id" bson:"_id"`
	Slug      Slug
	Nome      string `json:"nome" bson:"nome"`
	Contenuto string `json:"contenuto" bson:"contenuto"`
}

func NewRegola(
	id uuid.UUID,
	nome string,
	contenuto string,
) *Regola {
	slug, _ := NewSlug(nome)

	return &Regola{
		ID:        RegolaID(id),
		Slug:      slug,
		Nome:      nome,
		Contenuto: contenuto,
	}
}

func (r *Regola) EntityType() string {
	return "regola"
}
