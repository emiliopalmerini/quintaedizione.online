package domain

// ---------- Enum / VO di supporto ----------

// Tipo animale (semplificato rispetto ai mostri)
type TipoAnimale string

const (
	TipoAnimaleAnimale TipoAnimale = "Animale"
	TipoAnimaleBestia  TipoAnimale = "Bestia"
)

// Grado di Sfida
type GradoSfida struct {
	Valore          string `json:"valore"             bson:"valore"` // "0", "1/8", "1/4", "1/2", "1", "2", ecc.
	PuntiEsperienza int    `json:"punti_esperienza"   bson:"punti_esperienza"`
	BonusCompetenza int    `json:"bonus_competenza"   bson:"bonus_competenza"`
}

// Velocità multiple per animali
type VelocitaMultipla struct {
	Base     Velocita            `json:"base"    bson:"base"`
	Speciali map[string]Velocita `json:"speciali" bson:"speciali"` // "Scalare", "Nuoto", "Volo", ecc.
}

// Sensi speciali
type Sensi struct {
	SensiCiechi       int    `json:"sensi_ciechi"       bson:"sensi_ciechi,omitempty"` // metri
	Scurovisione      int    `json:"scurovisione"       bson:"scurovisione,omitempty"` // metri
	Tremosensore      int    `json:"tremosensore"       bson:"tremosensore,omitempty"` // metri
	PercezionePassiva int    `json:"percezione_passiva" bson:"percezione_passiva"`
	Altri             string `json:"altri"              bson:"altri,omitempty"` // testo libero per sensi speciali
}

// Abilità animale (diversa da quelle delle classi)
type AbilitaAnimale struct {
	Nome  string `json:"nome"  bson:"nome"`  // es. "Perception", "Athletics"
	Bonus int    `json:"bonus" bson:"bonus"` // bonus totale
}

// ---------- Entità ----------

type Animale struct {
	Slug             Slug             `json:"slug"            bson:"slug"`
	Nome             string           `json:"nome"            bson:"nome"`
	Taglia           Taglia           `json:"taglia"          bson:"taglia"`
	Tipo             TipoAnimale      `json:"tipo"            bson:"tipo"`
	Sottotipo        string           `json:"sottotipo"       bson:"sottotipo,omitempty"` // es. "Dinosauro"
	Allineamento     Allineamento     `json:"allineamento"    bson:"allineamento"`
	ClasseArmatura   ClasseArmatura   `json:"ca"              bson:"ca"`
	PuntiFerita      PuntiFerita      `json:"pf"              bson:"pf"`
	VelocitaMultipla VelocitaMultipla `json:"velocita_multipla" bson:"velocita_multipla"`
	Caratteristiche  []Caratteristica `json:"caratteristiche" bson:"caratteristiche"`
	Abilita          []AbilitaAnimale `json:"abilita"         bson:"abilita"`
	Sensi            Sensi            `json:"sensi"           bson:"sensi"`
	GradoSfida       GradoSfida       `json:"grado_sfida"     bson:"grado_sfida"`
	Tratti           []Tratto         `json:"tratti"          bson:"tratti"`
	Azioni           []Azione         `json:"azioni"          bson:"azioni"`
	Contenuto        string           `json:"contenuto"       bson:"contenuto"`
}

// ---------- Costruttore ----------

func NewAnimale(
	nome string,
	taglia Taglia,
	tipo TipoAnimale,
	sottotipo string,
	allineamento Allineamento,
	ca ClasseArmatura,
	pf PuntiFerita,
	velocitaMultipla VelocitaMultipla,
	caratteristiche []Caratteristica,
	abilita []AbilitaAnimale,
	sensi Sensi,
	gradoSfida GradoSfida,
	tratti []Tratto,
	azioni []Azione,
	contenuto string,
) *Animale {
	slug, _ := NewSlug(nome)

	return &Animale{
		Slug:             slug,
		Nome:             nome,
		Taglia:           taglia,
		Tipo:             tipo,
		Sottotipo:        sottotipo,
		Allineamento:     allineamento,
		ClasseArmatura:   ca,
		PuntiFerita:      pf,
		VelocitaMultipla: velocitaMultipla,
		Caratteristiche:  caratteristiche,
		Abilita:          abilita,
		Sensi:            sensi,
		GradoSfida:       gradoSfida,
		Tratti:           tratti,
		Azioni:           azioni,
		Contenuto:        contenuto,
	}
}

// EntityType implements ParsedEntity interface
func (a *Animale) EntityType() string {
	return "animale"
}
