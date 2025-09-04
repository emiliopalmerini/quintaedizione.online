package domain

import "github.com/google/uuid"

// ---------- ID ----------

type ArmaID uuid.UUID

// ---------- Enum / VO di supporto ----------

// Valuta per costi
type Valuta string

const (
	ValutaRame    Valuta = "mr"
	ValutaArgento Valuta = "ma"
	ValutaElettro Valuta = "me"
	ValutaOro     Valuta = "mo"
	ValutaPlatino Valuta = "mp"
)

// Unità di peso
type UnitaPeso string

const (
	UnitaKg     UnitaPeso = "kg"
	UnitaLibbre UnitaPeso = "lb"
)

// Categoria armi
type CategoriaArma string

const (
	CategoriaArmaSimpliceMischia  CategoriaArma = "Semplice da Mischia"
	CategoriaArmaSempliceDistanza CategoriaArma = "Semplice da Distanza"
	CategoriaArmaMarzialeMischia  CategoriaArma = "Marziale da Mischia"
	CategoriaArmaMarzialeDistanza CategoriaArma = "Marziale da Distanza"
)

// Proprietà armi
type ProprietaArma string

const (
	ProprietaAccurata        ProprietaArma = "Accurata"
	ProprietaAfferrare       ProprietaArma = "Afferrare"
	ProprietaArmaturaPesante ProprietaArma = "Armatura Pesante"
	ProprietaCaricare        ProprietaArma = "Caricare"
	ProprietaDaLancio        ProprietaArma = "Da Lancio"
	ProprietaDueManiBast     ProprietaArma = "A Due Mani (Bastone)"
	ProprietaDueManiLance    ProprietaArma = "A Due Mani (Lance)"
	ProprietaLeggera         ProprietaArma = "Leggera"
	ProprietaMunizioni       ProprietaArma = "Munizioni"
	ProprietaPortata         ProprietaArma = "Portata"
	ProprietaSpeciale        ProprietaArma = "Speciale"
	ProprietaVersatile       ProprietaArma = "Versatile"
)

// Costo
type Costo struct {
	Valore int    `json:"valore" bson:"valore"`
	Valuta Valuta `json:"valuta" bson:"valuta"`
}

// Peso
type Peso struct {
	Valore float64   `json:"valore" bson:"valore"`
	Unita  UnitaPeso `json:"unita"  bson:"unita"`
}

// Gittata (per armi da lancio e distanza)
type GittataArma struct {
	Normale string `json:"normale" bson:"normale"`
	Lunga   string `json:"lunga"   bson:"lunga"`
}

// ---------- Entità ----------

type Arma struct {
	ID        ArmaID          `json:"id"         bson:"_id"`
	Slug      Slug            `json:"slug"       bson:"slug"`
	Nome      string          `json:"nome"       bson:"nome"`
	Costo     Costo           `json:"costo"      bson:"costo"`
	Peso      Peso            `json:"peso"       bson:"peso"`
	Danno     string          `json:"danno"      bson:"danno"`
	Categoria CategoriaArma   `json:"categoria"  bson:"categoria"`
	Proprieta []ProprietaArma `json:"proprieta"  bson:"proprieta"`
	Maestria  string          `json:"maestria"   bson:"maestria,omitempty"`
	Gittata   *GittataArma    `json:"gittata"    bson:"gittata,omitempty"`
	Contenuto string          `json:"contenuto"  bson:"contenuto"`
}

// ---------- Costruttore ----------

func NewArma(
	id uuid.UUID,
	nome string,
	costo Costo,
	peso Peso,
	danno string,
	categoria CategoriaArma,
	proprieta []ProprietaArma,
	maestria string,
	gittata *GittataArma,
	contenuto string,
) *Arma {
	slug, _ := NewSlug(nome)

	return &Arma{
		ID:        ArmaID(id),
		Slug:      slug,
		Nome:      nome,
		Costo:     costo,
		Peso:      peso,
		Danno:     danno,
		Categoria: categoria,
		Proprieta: proprieta,
		Maestria:  maestria,
		Gittata:   gittata,
		Contenuto: contenuto,
	}
}
