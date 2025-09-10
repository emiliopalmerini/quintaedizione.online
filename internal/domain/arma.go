package domain

// ---------- Enum / VO di supporto ----------

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

// Gittata (per armi da lancio e distanza)
type GittataArma struct {
	Normale string `json:"normale" bson:"normale"`
	Lunga   string `json:"lunga"   bson:"lunga"`
}

// ---------- Entità ----------

type Arma struct {
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

// EntityType implements ParsedEntity interface
func (a *Arma) EntityType() string {
	return "arma"
}
