package domain

import "github.com/google/uuid"

// ---------- Enum / VO di supporto ----------

// Rarità oggetti magici
type Rarita string

const (
	RaritaComune      Rarita = "Comune"
	RaritaNonComune   Rarita = "Non Comune"
	RaritaRara        Rarita = "Raro"
	RaritaMoltoRara   Rarita = "Molto Raro"
	RaritaLeggendaria Rarita = "Leggendario"
	RaritaArtefatto   Rarita = "Artefatto"
)

// Tipo oggetto magico (categoria generale)
type TipoOggettoMagico string

const (
	TipoArmatura      TipoOggettoMagico = "Armatura"
	TipoArma          TipoOggettoMagico = "Arma"
	TipoBacchetta     TipoOggettoMagico = "Bacchetta"
	TipoVerga         TipoOggettoMagico = "Verga"
	TipoBastoneMagico TipoOggettoMagico = "BastoneMagico"
	TipoAnello        TipoOggettoMagico = "Anello"
	TipoAmuleto       TipoOggettoMagico = "Amuleto"
	TipoPozione       TipoOggettoMagico = "Pozione"
	TipoRotolo        TipoOggettoMagico = "Rotolo"
	TipoMeraviglioso  TipoOggettoMagico = "Meraviglioso"
)

// ---------- Entità ----------

type OggettoMagico struct {
	ID              OggettoMagicoID `json:"id"             bson:"_id"`
	Slug            Slug            `json:"slug"           bson:"slug"`
	Nome            string          `json:"nome"           bson:"nome"`
	Tipo            string          `json:"tipo"           bson:"tipo"` // descrizione dettagliata del tipo
	Rarita          Rarita          `json:"rarita"         bson:"rarita"`
	Sintonizzazione bool            `json:"sintonizzazione" bson:"sintonizzazione"` // richiede attunement
	Contenuto       string          `json:"contenuto"      bson:"contenuto"`
}

// ---------- Costruttore ----------

func NewOggettoMagico(
	id uuid.UUID,
	nome string,
	tipo string,
	rarita Rarita,
	sintonizzazione bool,
	contenuto string,
) *OggettoMagico {
	slug, _ := NewSlug(nome)

	return &OggettoMagico{
		ID:              OggettoMagicoID(id),
		Slug:            slug,
		Nome:            nome,
		Tipo:            tipo,
		Rarita:          rarita,
		Sintonizzazione: sintonizzazione,
		Contenuto:       contenuto,
	}
}

// EntityType implements ParsedEntity interface
func (o *OggettoMagico) EntityType() string {
	return "oggetto_magico"
}
