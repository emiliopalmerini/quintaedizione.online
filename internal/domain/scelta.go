package domain

type Scelta struct {
	Numero  uint8 `json:"numero" bson:"numero"`
	Opzioni []any `json:"opzioni" bson:"opzioni"`
}

func NewScelta(numero uint8, opzioni []any) Scelta {
	return Scelta{
		numero,
		opzioni,
	}
}
