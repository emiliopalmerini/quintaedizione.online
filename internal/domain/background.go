package domain

import "github.com/google/uuid"

type Background struct {
	ID                      BackgroundID       `json:"id"                         bson:"_id"`
	Slug                    Slug               `json:"slug"                      bson:"slug"`
	Nome                    string             `json:"nome"                      bson:"nome"`
	Caratteristiche         []CaratteristicaID `json:"caratteristiche"           bson:"caratteristiche"`
	CompetenzeAbilita       []AbilitaID        `json:"competenze_abilita_ids"    bson:"competenze_abilita_ids"`
	CompetenzeStrumenti     []StrumentoID      `json:"competenze_strumenti_ids"  bson:"competenze_strumenti_ids"`
	Talento                 TalentoID          `json:"talento_id"                bson:"talento_id"`
	EquipaggiamentoIniziale Scelta             `json:"equipaggiamento_iniziale"  bson:"equipaggiamento_iniziale"`
	Contenuto               string             `json:"contenuto"                 bson:"contenuto"`
}

func NewBackground(
	id uuid.UUID,
	nome string,
	car []uuid.UUID,
	abi []uuid.UUID,
	str []uuid.UUID,
	tal uuid.UUID,
	equip Scelta,
	cont string,
) *Background {
	slug, _ := NewSlug(nome)

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

	return &Background{
		ID:                      BackgroundID(id),
		Slug:                    slug,
		Nome:                    nome,
		Caratteristiche:         toCar,
		CompetenzeAbilita:       toAbi,
		CompetenzeStrumenti:     toStr,
		Talento:                 TalentoID(tal),
		EquipaggiamentoIniziale: equip,
		Contenuto:               cont,
	}
}

// EntityType implements ParsedEntity interface
func (b *Background) EntityType() string {
	return "background"
}
