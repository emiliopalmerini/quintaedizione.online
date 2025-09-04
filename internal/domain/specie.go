package domain

import "github.com/google/uuid"

// ---- Value Objects ----

type OpzioneEquip Scelta

// ---- Entità ----

type SpecieID uuid.UUID

type Specie struct {
	ID                         SpecieID           `json:"id"        bson:"_id"`
	Slug                       Slug               `json:"slug"      bson:"slug"`
	Nome                       string             `json:"nome"      bson:"nome"`
	PunteggiCaratteristica     []CaratteristicaID `json:"punteggi_caratteristica" bson:"punteggi_caratteristica"`
	AbilitaCompetenze          []AbilitaID        `json:"abilità_competenze_ids"   bson:"abilità_competenze_ids"`
	StrumentiCompetenze        []StrumentoID      `json:"strumenti_competenze_ids" bson:"strumenti_competenze_ids"`
	Talento                    TalentoID          `json:"talento_id"               bson:"talento_id"`
	EquipaggiamentoInizialeOpt []OpzioneEquip     `json:"equipaggiamento_iniziale_opzioni" bson:"equipaggiamento_iniziale_opzioni"`
	Contenuto                  string             `json:"contenuto" bson:"contenuto"`
}

// ---- Costruttore tip-safe ----

func NewSpecie(
	id uuid.UUID,
	nome string,
	car []uuid.UUID,
	abi []uuid.UUID,
	str []uuid.UUID,
	tal uuid.UUID,
	equip []OpzioneEquip,
	cont string,
) (*Specie, error) {
	sg, err := NewSlug(nome)
	if err != nil {
		return nil, err
	}

	toCar := make([]CaratteristicaID, len(car))
	for i, v := range car {
		toCar[i] = CaratteristicaID(v)
	}
	toAbi := make([]AbilitaID, len(abi))
	for i, v := range abi {
		toAbi[i] = AbilitaID(v)
	}
	toStr := make([]StrumentoID, len(str))
	for i, v := range str {
		toStr[i] = StrumentoID(v)
	}

	return &Specie{
		ID:                         SpecieID(id),
		Slug:                       sg,
		Nome:                       nome,
		PunteggiCaratteristica:     toCar,
		AbilitaCompetenze:          toAbi,
		StrumentiCompetenze:        toStr,
		Talento:                    TalentoID(tal),
		EquipaggiamentoInizialeOpt: equip,
		Contenuto:                  cont,
	}, nil
}
