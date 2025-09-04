package domain

// ---------- Value Objects comuni ----------
type NomeCaratteristica string

const (
	Forza        NomeCaratteristica = "Forza"
	Destrezza    NomeCaratteristica = "Destrezza"
	Carisma      NomeCaratteristica = "Carisma"
	Costituzione NomeCaratteristica = "Costituzione"
	Intelligenza NomeCaratteristica = "Intelligenza"
	Saggezza     NomeCaratteristica = "Saggezza"
)

type AbbreviazioneCaratteristica string

const (
	FOR AbbreviazioneCaratteristica = "FOR"
	DES AbbreviazioneCaratteristica = "DES"
	CAR AbbreviazioneCaratteristica = "CAR"
	COS AbbreviazioneCaratteristica = "COS"
	INT AbbreviazioneCaratteristica = "INT"
	SAG AbbreviazioneCaratteristica = "SAG"
)

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

// Costo unificato
type Costo struct {
	Valore int    `json:"valore" bson:"valore"`
	Valuta Valuta `json:"valuta" bson:"valuta"`
}

// Peso unificato
type Peso struct {
	Valore float64   `json:"valore" bson:"valore"`
	Unita  UnitaPeso `json:"unita"  bson:"unita"`
}

// Taglia (comune a mostri e animali)
type Taglia string

const (
	TagliaMinuscola Taglia = "Minuscola"
	TagliaPiccola   Taglia = "Piccola"
	TagliaMedia     Taglia = "Media"
	TagliaGrande    Taglia = "Grande"
	TagliaEnorme    Taglia = "Enorme"
	TagliaColossale Taglia = "Colossale"
)

// Unità di velocità
type UnitaVelocita string

const (
	UnitaMetriRound UnitaVelocita = "m/round"
	UnitaMetriTurno UnitaVelocita = "m/turno"
	UnitaKmOra      UnitaVelocita = "km/h"
	UnitaNodi       UnitaVelocita = "nodi"
	UnitaMetri      UnitaVelocita = "m"
)

// Velocità unificata
type Velocita struct {
	Valore int           `json:"valore" bson:"valore"`
	Unita  UnitaVelocita `json:"unita"  bson:"unita"`
}

// Punti Ferita unificati
type PuntiFerita struct {
	Totali     int `json:"totali"     bson:"totali"`     // punti ferita massimi
	Attuali    int `json:"attuali"    bson:"attuali"`    // punti ferita attuali
	Temporanei int `json:"temporanei" bson:"temporanei"` // punti ferita temporanei
}

// Tipo di caratteristica unificato
type TipoCaratteristica string

const (
	CaratteristicaForza        TipoCaratteristica = "Forza"
	CaratteristicaDestrezza    TipoCaratteristica = "Destrezza"
	CaratteristicaCostituzione TipoCaratteristica = "Costituzione"
	CaratteristicaIntelligenza TipoCaratteristica = "Intelligenza"
	CaratteristicaSaggezza     TipoCaratteristica = "Saggezza"
	CaratteristicaCarisma      TipoCaratteristica = "Carisma"
)

// Caratteristica unificata
type Caratteristica struct {
	Tipo   TipoCaratteristica `json:"tipo"   bson:"tipo"`
	Valore int                `json:"valore" bson:"valore"`
}

// Dado per rappresentare dadi (es. "20d10")
type Dado struct {
	Numero int `json:"numero" bson:"numero"` // numero di dadi
	Facce  int `json:"facce"  bson:"facce"`  // facce del dado (d4, d6, d8, d10, d12, d20)
	Bonus  int `json:"bonus"  bson:"bonus"`  // modificatore fisso
}

// Azione generica (usata in mostri, animali, ecc.)
type Azione struct {
	Nome        string `json:"nome"        bson:"nome"`
	Descrizione string `json:"descrizione" bson:"descrizione"`
}

// Tratto generico (usato in mostri, animali, ecc.)
type Tratto struct {
	Nome        string `json:"nome"        bson:"nome"`
	Descrizione string `json:"descrizione" bson:"descrizione"`
}

// ---------- Costruttori ----------

func NewCosto(valore int, valuta Valuta) Costo {
	return Costo{Valore: valore, Valuta: valuta}
}

func NewPeso(valore float64, unita UnitaPeso) Peso {
	return Peso{Valore: valore, Unita: unita}
}

func NewVelocita(valore int, unita UnitaVelocita) Velocita {
	return Velocita{Valore: valore, Unita: unita}
}

func NewPuntiFerita(totali int) PuntiFerita {
	return PuntiFerita{Totali: totali, Attuali: totali, Temporanei: 0}
}

func NewPuntiFeriteCustom(totali, attuali, temporanei int) PuntiFerita {
	return PuntiFerita{Totali: totali, Attuali: attuali, Temporanei: temporanei}
}

func NewCaratteristica(tipo TipoCaratteristica, valore int) Caratteristica {
	return Caratteristica{Tipo: tipo, Valore: valore}
}

func NewDado(numero, facce, bonus int) Dado {
	return Dado{Numero: numero, Facce: facce, Bonus: bonus}
}

func NewAzione(nome, descrizione string) Azione {
	return Azione{Nome: nome, Descrizione: descrizione}
}

func NewTraito(nome, descrizione string) Tratto {
	return Tratto{Nome: nome, Descrizione: descrizione}
}

// ---------- Metodi ----------

func (c Caratteristica) Modificatore() int {
	return (c.Valore - 10) / 2
}

func (pf PuntiFerita) Effettivi() int {
	return pf.Attuali + pf.Temporanei
}

