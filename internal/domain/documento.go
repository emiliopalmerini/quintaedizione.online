package domain

type Documento struct {
	Pagina    int    `json:"pagina"    bson:"pagina"`
	Slug      Slug   `json:"slug"      bson:"slug"`
	Titolo    string `json:"titolo"    bson:"titolo"`
	Contenuto string `json:"contenuto" bson:"contenuto"`
}

func NewDocumento(pagina int, titolo, contenuto string) *Documento {
	slug, _ := NewSlug(titolo)

	return &Documento{
		Pagina:    pagina,
		Slug:      slug,
		Titolo:    titolo,
		Contenuto: contenuto,
	}
}

// EntityType implements ParsedEntity interface
func (d *Documento) EntityType() string {
	return "documento"
}
