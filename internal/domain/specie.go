package domain

// ---- Value Objects ----

type OpzioneEquip Scelta

// ---- Entità ----

type Specie struct {
	Slug                       Slug                 `json:"slug"      bson:"slug"`
	Nome                       string               `json:"nome"      bson:"nome"`
	PunteggiCaratteristica     []CaratteristicaSlug `json:"punteggi_caratteristica" bson:"punteggi_caratteristica"`
	AbilitaCompetenze          []AbilitaSlug        `json:"abilità_competenze_ids"   bson:"abilità_competenze_ids"`
	StrumentiCompetenze        []StrumentoSlug      `json:"strumenti_competenze_ids" bson:"strumenti_competenze_ids"`
	Talento                    TalentoSlug          `json:"talento_id"               bson:"talento_id"`
	EquipaggiamentoInizialeOpt []OpzioneEquip       `json:"equipaggiamento_iniziale_opzioni" bson:"equipaggiamento_iniziale_opzioni"`
	Contenuto                  string               `json:"contenuto" bson:"contenuto"`
}

// ---- Costruttore tip-safe ----

func NewSpecie(
	nome string,
	car []string,
	abi []string,
	str []string,
	tal string,
	equip []OpzioneEquip,
	cont string,
) (*Specie, error) {
	sg, err := NewSlug(nome)
	if err != nil {
		return nil, err
	}

	toCar := make([]CaratteristicaSlug, len(car))
	for i, v := range car {
		carSlug, _ := NewSlug(v)
		toCar[i] = CaratteristicaSlug(carSlug)
	}
	toAbi := make([]AbilitaSlug, len(abi))
	for i, v := range abi {
		abiSlug, _ := NewSlug(v)
		toAbi[i] = AbilitaSlug(abiSlug)
	}
	toStr := make([]StrumentoSlug, len(str))
	for i, v := range str {
		strSlug, _ := NewSlug(v)
		toStr[i] = StrumentoSlug(strSlug)
	}

	talSlug, _ := NewSlug(tal)

	return &Specie{
		Slug:                       sg,
		Nome:                       nome,
		PunteggiCaratteristica:     toCar,
		AbilitaCompetenze:          toAbi,
		StrumentiCompetenze:        toStr,
		Talento:                    TalentoSlug(talSlug),
		EquipaggiamentoInizialeOpt: equip,
		Contenuto:                  cont,
	}, nil
}

// EntityType returns the entity type identifier
func (s *Specie) EntityType() string {
	return "specie"
}
