package domain

type Background struct {
	Slug                    Slug                 `json:"slug"                      bson:"slug"`
	Nome                    string               `json:"nome"                      bson:"nome"`
	Caratteristiche         []CaratteristicaSlug `json:"caratteristiche"           bson:"caratteristiche"`
	CompetenzeAbilita       []AbilitaSlug        `json:"competenze_abilita_ids"    bson:"competenze_abilita_ids"`
	CompetenzeStrumenti     []StrumentoSlug      `json:"competenze_strumenti_ids"  bson:"competenze_strumenti_ids"`
	Talento                 TalentoSlug          `json:"talento_id"                bson:"talento_id"`
	EquipaggiamentoIniziale Scelta               `json:"equipaggiamento_iniziale"  bson:"equipaggiamento_iniziale"`
	Contenuto               string               `json:"contenuto"                 bson:"contenuto"`
}

func NewBackground(
	nome string,
	car []string,
	abi []string,
	str []string,
	tal string,
	equip Scelta,
	cont string,
) *Background {
	slug, _ := NewSlug(nome)

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

	return &Background{
		Slug:                    slug,
		Nome:                    nome,
		Caratteristiche:         toCar,
		CompetenzeAbilita:       toAbi,
		CompetenzeStrumenti:     toStr,
		Talento:                 TalentoSlug(talSlug),
		EquipaggiamentoIniziale: equip,
		Contenuto:               cont,
	}
}

// EntityType implements ParsedEntity interface
func (b *Background) EntityType() string {
	return "background"
}
