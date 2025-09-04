package domain

import "github.com/google/uuid"

type DocumentoID uuid.UUID

type Documento struct {
	ID        DocumentoID `json:"id"        bson:"_id"`
	Pagina    int         `json:"pagina"    bson:"pagina"`
	Slug      Slug        `json:"slug"      bson:"slug"`
	Titolo    string      `json:"titolo"    bson:"titolo"`
	Contenuto string      `json:"contenuto" bson:"contenuto"`
}

func NewDocumento(id uuid.UUID, pagina int, titolo, contenuto string) *Documento {
	slug, _ := NewSlug(titolo)

	return &Documento{
		ID:        DocumentoID(id),
		Pagina:    pagina,
		Slug:      slug,
		Titolo:    titolo,
		Contenuto: contenuto,
	}
}
