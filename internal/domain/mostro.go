package domain

import "github.com/google/uuid"

// ---------- Enum / VO di supporto ----------

// Tipo mostro
type TipoMostro string

const (
	TipoAberrazione TipoMostro = "Aberrazione"
	TipoBestia      TipoMostro = "Bestia"
	TipoCostrutto   TipoMostro = "Costrutto"
	TipoDrago       TipoMostro = "Drago"
	TipoElementale  TipoMostro = "Elementale"
	TipoFata        TipoMostro = "Fata"
	TipoFolletto    TipoMostro = "Folletto"
	TipoGigante     TipoMostro = "Gigante"
	TipoUmanoide    TipoMostro = "Umanoide"
	TipoMelma       TipoMostro = "Melma"
	TipoMostrosoide TipoMostro = "Mostrosoide"
	TipoNonMorto    TipoMostro = "Non Morto"
	TipoPianta      TipoMostro = "Pianta"
)

// Allineamento
type Allineamento string

const (
	AllineamentoLegaleBuono      Allineamento = "Legale Buono"
	AllineamentoNeutraleBuono    Allineamento = "Neutrale Buono"
	AllineamentoCaoticoBuono     Allineamento = "Caotico Buono"
	AllineamentoLegaleNeutrale   Allineamento = "Legale Neutrale"
	AllineamentoNeutrale         Allineamento = "Neutrale"
	AllineamentoCaoticoNeutrale  Allineamento = "Caotico Neutrale"
	AllineamentoLegaleMalvagio   Allineamento = "Legale Malvagio"
	AllineamentoNeutraleMalvagio Allineamento = "Neutrale Malvagio"
	AllineamentoCaoticoMalvagio  Allineamento = "Caotico Malvagio"
)

// Punti Esperienza
type PuntiEsperienza struct {
	Base int `json:"base" bson:"base"`
	Tana int `json:"tana" bson:"tana,omitempty"` // PE aggiuntivi se in tana
}

// Sensibilità (vulnerabilità, resistenze)
type Sensibilita struct {
	Vulnerabilita []string `json:"vulnerabilita" bson:"vulnerabilita"`
	Resistenze    []string `json:"resistenze"    bson:"resistenze"`
}

// Tiri salvezza
type TiriSalvezza map[TipoCaratteristica]int // es. "Destrezza": +5, "Saggezza": +3

// Abilità
type AbilitaMostro map[AbilitaID]int // es. "percezione": +8, "furtivita": +4

// Immunità
type Immunita struct {
	Danni      []DannoID `json:"danni"      bson:"danni"`
	Condizioni []string  `json:"condizioni" bson:"condizioni"`
}

// Reazione del mostro
type ReazioneMostro struct {
	Nome        string `json:"nome"        bson:"nome"`
	Descrizione string `json:"descrizione" bson:"descrizione"`
}

// Azione leggendaria
type AzioneLeggendaria struct {
	Nome        string `json:"nome"        bson:"nome"`
	Costo       int    `json:"costo"       bson:"costo"` // costo in punti azione leggendaria
	Descrizione string `json:"descrizione" bson:"descrizione"`
}

// Incantesimi del mostro
type IncantesimiMostro struct {
	CD      *int            `json:"cd"       bson:"cd,omitempty"`      // Classe Difficoltà (opzionale)
	Attacco *int            `json:"attacco"  bson:"attacco,omitempty"` // Bonus attacco incantesimi (opzionale)
	Lista   []IncantesimoID `json:"lista"    bson:"lista"`             // lista incantesimi per ID
}

// ---------- Entità ----------

type Mostro struct {
	ID                MostroID            `json:"id"                  bson:"_id"`
	Slug              Slug                `json:"slug"                bson:"slug"`
	Nome              string              `json:"nome"                bson:"nome"`
	Taglia            Taglia              `json:"taglia"              bson:"taglia"`
	Tipo              TipoMostro          `json:"tipo"                bson:"tipo"`
	Allineamento      Allineamento        `json:"allineamento"        bson:"allineamento"`
	GradoSfida        int                 `json:"gs"                  bson:"gs"`
	PuntiEsperienza   PuntiEsperienza     `json:"pe"                  bson:"pe"`
	ClasseArmatura    ClasseArmatura      `json:"ac"                  bson:"ac"`
	PuntiFerita       PuntiFerita         `json:"hp"                  bson:"hp"`
	Velocita          Velocita            `json:"velocita"            bson:"velocita"`
	Caratteristiche   []Caratteristica    `json:"caratteristiche"     bson:"caratteristiche"`
	Sensibilita       Sensibilita         `json:"sensibilita"         bson:"sensibilita"`
	TiriSalvezza      TiriSalvezza        `json:"tiri_salvezza"       bson:"tiri_salvezza"`
	Abilita           AbilitaMostro       `json:"abilita"             bson:"abilita"`
	Immunita          Immunita            `json:"immunita"            bson:"immunita"`
	Azioni            []Azione            `json:"azioni"              bson:"azioni"`
	Tratti            []Tratto            `json:"tratti"              bson:"tratti"`
	Reazioni          []ReazioneMostro    `json:"reazioni"            bson:"reazioni"`
	AzioniLeggendarie []AzioneLeggendaria `json:"azioni_leggendarie"  bson:"azioni_leggendarie"`
	Incantesimi       IncantesimiMostro   `json:"incantesimi"         bson:"incantesimi"`
	Contenuto         string              `json:"contenuto"           bson:"contenuto"`
}

// ---------- Costruttore ----------

func NewMostro(
	id uuid.UUID,
	nome string,
	taglia Taglia,
	tipo TipoMostro,
	allineamento Allineamento,
	gradoSfida int,
	puntiEsperienza PuntiEsperienza,
	classeArmatura ClasseArmatura,
	puntiFerita PuntiFerita,
	velocita Velocita,
	caratteristiche []Caratteristica,
	sensibilita Sensibilita,
	tiriSalvezza TiriSalvezza,
	abilita AbilitaMostro,
	immunita Immunita,
	azioni []Azione,
	tratti []Tratto,
	reazioni []ReazioneMostro,
	azioniLeggendarie []AzioneLeggendaria,
	incantesimi IncantesimiMostro,
	contenuto string,
) *Mostro {
	slug, _ := NewSlug(nome)

	return &Mostro{
		ID:                MostroID(id),
		Slug:              slug,
		Nome:              nome,
		Taglia:            taglia,
		Tipo:              tipo,
		Allineamento:      allineamento,
		GradoSfida:        gradoSfida,
		PuntiEsperienza:   puntiEsperienza,
		ClasseArmatura:    classeArmatura,
		PuntiFerita:       puntiFerita,
		Velocita:          velocita,
		Caratteristiche:   caratteristiche,
		Sensibilita:       sensibilita,
		TiriSalvezza:      tiriSalvezza,
		Abilita:           abilita,
		Immunita:          immunita,
		Azioni:            azioni,
		Tratti:            tratti,
		Reazioni:          reazioni,
		AzioniLeggendarie: azioniLeggendarie,
		Incantesimi:       incantesimi,
		Contenuto:         contenuto,
	}
}

// EntityType implements ParsedEntity interface
func (m *Mostro) EntityType() string {
	return "mostro"
}
