package domain

import "github.com/google/uuid"

// ---------- Enum / VO di supporto ----------

// Categoria servizi
type CategoriaServizio string

const (
	CategoriaTenorevita     CategoriaServizio = "Tenore di vita"
	CategoriaAlloggio       CategoriaServizio = "Alloggio"
	CategoriaTrasporto      CategoriaServizio = "Trasporto"
	CategoriaServizioMagico CategoriaServizio = "Servizio Magico"
)

// Costo esteso per servizi (include "gratuito")
type CostoServizio struct {
	Valore int    `json:"valore" bson:"valore"`
	Valuta string `json:"valuta" bson:"valuta"` // può essere "mr", "ma", "me", "mo", "mp" o "gratuito"
}

// ---------- Entità ----------

type Servizio struct {
	ID          ServizioID        `json:"id"          bson:"_id"`
	Slug        Slug              `json:"slug"        bson:"slug"`
	Nome        string            `json:"nome"        bson:"nome"`
	Costo       CostoServizio     `json:"costo"       bson:"costo"`
	Categoria   CategoriaServizio `json:"categoria"   bson:"categoria"`
	Descrizione string            `json:"descrizione" bson:"descrizione"`
	Contenuto   string            `json:"contenuto"   bson:"contenuto"`
}

// ---------- Costruttore ----------

func NewServizio(
	id uuid.UUID,
	nome string,
	costo CostoServizio,
	categoria CategoriaServizio,
	descrizione string,
	contenuto string,
) *Servizio {
	slug, _ := NewSlug(nome)

	return &Servizio{
		ID:          ServizioID(id),
		Slug:        slug,
		Nome:        nome,
		Costo:       costo,
		Categoria:   categoria,
		Descrizione: descrizione,
		Contenuto:   contenuto,
	}
}

