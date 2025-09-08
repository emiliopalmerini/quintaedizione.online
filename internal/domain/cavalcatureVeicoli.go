package domain

import "github.com/google/uuid"

// ---------- Enum / VO di supporto ----------

// Tipo cavalcatura/veicolo
type TipoCavalcaturaVeicolo string

const (
	TipoCavalcatura TipoCavalcaturaVeicolo = "cavalcatura"
	TipoNave        TipoCavalcaturaVeicolo = "nave"
	TipoVeicolo     TipoCavalcaturaVeicolo = "veicolo"
	TipoAltro       TipoCavalcaturaVeicolo = "altro"
)

// Velocità per cavalcature/veicoli (può essere null)
type VelocitaVeicolo struct {
	Valore *int          `json:"valore" bson:"valore,omitempty"`
	Unita  UnitaVelocita `json:"unita"  bson:"unita"`
}

// ---------- Entità ----------

type CavalcaturaVeicolo struct {
	ID             CavalcaturaVeicoloID   `json:"id"              bson:"_id"`
	Slug           Slug                   `json:"slug"            bson:"slug"`
	Nome           string                 `json:"nome"            bson:"nome"`
	Tipo           TipoCavalcaturaVeicolo `json:"tipo"            bson:"tipo"`
	Costo          Costo                  `json:"costo"           bson:"costo"`
	Velocita       VelocitaVeicolo        `json:"velocita"        bson:"velocita"`
	CapacitaCarico Peso                   `json:"capacita_carico" bson:"capacita_carico"`
	Equipaggio     *int                   `json:"equipaggio"      bson:"equipaggio,omitempty"`
	Passeggeri     *int                   `json:"passeggeri"      bson:"passeggeri,omitempty"`
	CA             *int                   `json:"ca"              bson:"ca,omitempty"`
	PF             *int                   `json:"pf"              bson:"pf,omitempty"`
	SogliaDanni    *int                   `json:"soglia_danni"    bson:"soglia_danni,omitempty"`
	Descrizione    string                 `json:"descrizione"     bson:"descrizione"`
	Contenuto      string                 `json:"contenuto"       bson:"contenuto"`
}

// ---------- Costruttore ----------

func NewCavalcaturaVeicolo(
	id uuid.UUID,
	nome string,
	tipo TipoCavalcaturaVeicolo,
	costo Costo,
	velocita VelocitaVeicolo,
	capacitaCarico Peso,
	equipaggio *int,
	passeggeri *int,
	ca *int,
	pf *int,
	sogliaDanni *int,
	descrizione string,
	contenuto string,
) *CavalcaturaVeicolo {
	slug, _ := NewSlug(nome)

	return &CavalcaturaVeicolo{
		ID:             CavalcaturaVeicoloID(id),
		Slug:           slug,
		Nome:           nome,
		Tipo:           tipo,
		Costo:          costo,
		Velocita:       velocita,
		CapacitaCarico: capacitaCarico,
		Equipaggio:     equipaggio,
		Passeggeri:     passeggeri,
		CA:             ca,
		PF:             pf,
		SogliaDanni:    sogliaDanni,
		Descrizione:    descrizione,
		Contenuto:      contenuto,
	}
}

// EntityType returns the entity type identifier
func (c *CavalcaturaVeicolo) EntityType() string {
	return "cavalcatura_veicolo"
}
