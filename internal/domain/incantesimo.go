package domain

// ---------- Enum / VO di supporto ----------

type ComponenteIncantesimo string

const (
	CompV ComponenteIncantesimo = "V"
	CompS ComponenteIncantesimo = "S"
	CompM ComponenteIncantesimo = "M"
)

// Tempo di lancio
type TempoTipo string

const (
	TempoAzione      TempoTipo = "Azione"
	TempoAzioneBonus TempoTipo = "AzioneBonus"
	TempoReazione    TempoTipo = "Reazione"
	TempoMinuti      TempoTipo = "Minuti"
	TempoOre         TempoTipo = "Ore"
	TempoSpeciale    TempoTipo = "Speciale"
)

type TempoLancio struct {
	Tipo   TempoTipo `json:"tipo"   bson:"tipo"`
	Valore int       `json:"valore" bson:"valore"` // minuti/ore se applicabile, 0 per Azione/Reazione/AzioneBonus/Speciale
	Nota   string    `json:"nota"   bson:"nota"`   // trigger di Reazione o testo libero
}

// Distanza
type Distanza struct {
	Valore float64 `json:"valore" bson:"valore"`
	Unita  string  `json:"unita"  bson:"unita"` // "ft","m","km","mile"
}

// Gittata
type GittataTipo string

const (
	GittataDistanza GittataTipo = "Distanza"
	GittataContatto GittataTipo = "Contatto"
	GittataSe       GittataTipo = "Se"
	GittataVista    GittataTipo = "Vista"
	GittataSpeciale GittataTipo = "Speciale"
)

type GittataIncantesimo struct {
	Tipo     GittataTipo `json:"tipo"     bson:"tipo"`
	Distanza *Distanza   `json:"distanza" bson:"distanza,omitempty"` // presente solo se Tipo==Distanza
	Nota     string      `json:"nota"     bson:"nota"`               // testo libero quando Speciale
}

// Durata
type DurataTipo string

const (
	DurataIstantanea     DurataTipo = "Istantanea"
	DurataConcentrazione DurataTipo = "Concentrazione"
	DurataTempo          DurataTipo = "Tempo"
	DurataSpeciale       DurataTipo = "Speciale"
)

type Durata struct {
	Tipo           DurataTipo `json:"tipo"           bson:"tipo"`
	Minuti         int        `json:"minuti"         bson:"minuti"`         // per Tempo o Concentrazione
	Concentrazione bool       `json:"concentrazione" bson:"concentrazione"` // true se richiede concentrazione
	Nota           string     `json:"nota"           bson:"nota"`           // testo libero quando Speciale
}

// Componenti
type Componenti struct {
	V         bool   `json:"v"        bson:"v"`
	S         bool   `json:"s"        bson:"s"`
	M         bool   `json:"m"        bson:"m"`
	Materiali string `json:"materiali" bson:"materiali"` // descrizione materiali se M==true
}

// Lancio aggrega i VO di lancio
type Lancio struct {
	Tempo      TempoLancio        `json:"tempo"      bson:"tempo"`
	Gittata    GittataIncantesimo `json:"gittata"    bson:"gittata"`
	Componenti Componenti         `json:"componenti" bson:"componenti"`
	Durata     Durata             `json:"durata"     bson:"durata"`
}

// ---------- Entità ----------

type Incantesimo struct {
	Slug      Slug                  `json:"slug"     bson:"slug"`
	Nome      string                `json:"nome"     bson:"nome"`
	Livello   uint8                 `json:"livello"  bson:"livello"`
	Scuola    ScuolaIncantesimoSlug `json:"scuola"   bson:"scuola"`
	Classi    []ClasseSlug          `json:"classi"   bson:"classi"` // riferimenti per nome/slug business identifiers
	Lancio    Lancio                `json:"lancio"   bson:"lancio"`
	Contenuto string                `json:"contenuto" bson:"contenuto"`
}

// ---------- Costruttore ----------

func NewIncantesimo(
	nome string,
	livello uint8,
	scuola string,
	classi []string,
	lancio Lancio,
	contenuto string,
) *Incantesimo {
	sg, _ := NewSlug(nome)
	classiSlugs := make([]ClasseSlug, len(classi))
	for i, c := range classi {
		classeSlug, _ := NewSlug(c)
		classiSlugs[i] = ClasseSlug(classeSlug)
	}

	scuolaSlug, _ := NewSlug(scuola)

	return &Incantesimo{
		Slug:      sg,
		Nome:      nome,
		Livello:   livello,
		Scuola:    ScuolaIncantesimoSlug(scuolaSlug),
		Classi:    classiSlugs,
		Lancio:    lancio,
		Contenuto: contenuto,
	}
}

// EntityType implements ParsedEntity interface
func (i *Incantesimo) EntityType() string {
	return "incantesimo"
}
