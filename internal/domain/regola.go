package domain

type Regola struct {
	Slug      Slug
	Nome      string `json:"nome" bson:"nome"`
	Contenuto string `json:"contenuto" bson:"contenuto"`
}

func NewRegola(
	nome string,
	contenuto string,
) *Regola {
	slug, _ := NewSlug(nome)

	return &Regola{
		Slug:      slug,
		Nome:      nome,
		Contenuto: contenuto,
	}
}

func (r *Regola) EntityType() string {
	return "regola"
}
