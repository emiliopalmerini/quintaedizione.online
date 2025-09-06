package domain

// TODO: da controllare

import "github.com/google/uuid"

// ---------- Enum / VO di supporto ----------

// Competenze armature
type CompetenzaArmatura string

const (
	CompetenzaArmatureLeggere CompetenzaArmatura = "Leggere"
	CompetenzaArmatureMedie   CompetenzaArmatura = "Medie"
	CompetenzaArmaturePesanti CompetenzaArmatura = "Pesanti"
	CompetenzaScudi           CompetenzaArmatura = "Scudi"
)

// Tipo preparazione incantesimi
type TipoPreparazione string

const (
	PreparazionePreparato  TipoPreparazione = "Preparato"
	PreparazioneConosciuto TipoPreparazione = "Conosciuto"
	PreparazioneNone       TipoPreparazione = "Nessuna"
)

type NomeAbilita string

const (
	Acrobazia         NomeAbilita = "Acrobazia"
	AddestrareAnimali NomeAbilita = "Addestrare animali"
	Arcano            NomeAbilita = "Arcano"
	Atletica          NomeAbilita = "Atletica"
	Indagare          NomeAbilita = "Indagare"
	Ingannare         NomeAbilita = "Ingannare"
	Intimidirire      NomeAbilita = "Intimidire"
	Intuizione        NomeAbilita = "Intuizione"
	Storia            NomeAbilita = "Storia"
	Medicina          NomeAbilita = "Medicina"
	Natura            NomeAbilita = "Natura"
	Percezione        NomeAbilita = "Percezione"
	Intrattenere      NomeAbilita = "Intrattenere"
	Persuasione       NomeAbilita = "Persuasione"
	Religione         NomeAbilita = "Religione"
	RapiditaDiMano    NomeAbilita = "Rapidità di mano"
	Furtivita         NomeAbilita = "Furtività"
	Sopravvivenza     NomeAbilita = "Sopravvivenza"
)

type Abilita struct {
	ID                     AbilitaID          `json:"id"   bson:"id"`
	Nome                   NomeAbilita        `json:"nome" bson:"nome"`
	CaratteristicaRelativa NomeCaratteristica `json:"caratteristica_relativa" bson:"caratteristica_relativa"`
	Competenza             bool               `json:"competenza" bson:"competenza"`
}

// Equipaggiamento iniziale opzione
type EquipaggiamentoOpzione struct {
	Etichetta string              `json:"etichetta" bson:"etichetta"`
	Oggetti   []EquipaggiamentoID `json:"oggetti"   bson:"oggetti"`
}

// Multiclasse
type Multiclasse struct {
	Prerequisiti    []string `json:"prerequisiti"     bson:"prerequisiti"`
	TrattiAcquisiti []string `json:"tratti_acquisiti" bson:"tratti_acquisiti"`
	Note            string   `json:"note"             bson:"note"`
}

// Progressione livelli (mappa livello -> valore)
type ProgressioneLivelli map[int]any

// Risorsa di classe
type RisorsaClasse struct {
	Chiave  string              `json:"chiave"   bson:"chiave"`
	Livelli ProgressioneLivelli `json:"livelli"  bson:"livelli"`
}

// Progressioni
type Progressioni struct {
	MaestriaArmi       ProgressioneLivelli `json:"maestria_armi"          bson:"maestria_armi"`
	StiliCombattimento struct {
		Livelli ProgressioneLivelli `json:"livelli" bson:"livelli"`
		Scelte  []string            `json:"scelte"  bson:"scelte"`
	} `json:"stili_combattimento"    bson:"stili_combattimento"`
	AttacchiExtra         ProgressioneLivelli `json:"attacchi_extra"         bson:"attacchi_extra"`
	Risorse               []RisorsaClasse     `json:"risorse"                bson:"risorse"`
	AumentiCaratteristica []int               `json:"aumenti_caratteristica" bson:"aumenti_caratteristica"`
	DonoEpico             int                 `json:"dono_epico"             bson:"dono_epico"`
}

// Patto Warlock
type SlotIncantesimo struct {
	Slot        ProgressioneLivelli `json:"slot"        bson:"slot"`
	LivelloSlot ProgressioneLivelli `json:"livello_slot" bson:"livello_slot"`
}

// Magia
type Magia struct {
	HaIncantesimi             bool                `json:"ha_incantesimi"                 bson:"ha_incantesimi"`
	ListaRiferimento          ListaIncantesimi    `json:"lista_riferimento"              bson:"lista_riferimento"`
	CaratteristicaIncantatore *NomeCaratteristica `json:"caratteristica_incantatore"     bson:"caratteristica_incantatore,omitempty"`
	Preparazione              TipoPreparazione    `json:"preparazione"                   bson:"preparazione"`
	Focus                     string              `json:"focus"                          bson:"focus"`
	Trucchetti                ListaIncantesimi    `json:"trucchetti"                     bson:"trucchetti"`
	Incantesimi               ListaIncantesimi    `json:"incantesimi_preparati_o_noti"   bson:"incantesimi_preparati_o_noti"`
}

// Entry tabella livelli
type TabellaLivello struct {
	Livello              int            `json:"livello"                 bson:"livello"`
	BonusCompetenza      int            `json:"bonus_competenza"        bson:"bonus_competenza"`
	PrivilegiDiClasse    []string       `json:"privilegi_di_classe"     bson:"privilegi_di_classe"`
	Risorse              map[string]int `json:"risorse"                 bson:"risorse"`
	Trucchetti           int            `json:"trucchetti"              bson:"trucchetti"`
	IncantesimiPreparati int            `json:"incantesimi_preparati"   bson:"incantesimi_preparati"`
	Slot                 []int          `json:"slot"                    bson:"slot"`
	Note                 string         `json:"note"                    bson:"note"`
}

type Privilegio struct {
	Nome        string `json:"nome"        bson:"nome"`
	Livello     int    `json:"livello"     bson:"livello"`
	Descrizione string `json:"descrizione" bson:"descrizione"`
}

// Sottoclasse
type Sottoclasse struct {
	Slug                       Slug                       `json:"slug"                         bson:"slug"`
	Nome                       string                     `json:"nome"                         bson:"nome"`
	Descrizione                string                     `json:"descrizione"                  bson:"descrizione"`
	PrivilegiSottoclasse       []Privilegio               `json:"privilegi_sottoclasse"        bson:"privilegi_sottoclasse"`
	IncantesimiSemprePreparati map[string][]IncantesimoID `json:"incantesimi_sempre_preparati" bson:"incantesimi_sempre_preparati"`
}

// Liste incantesimi per livello
type ListaIncantesimi map[int][]IncantesimoID

// Raccomandazioni
type Raccomandazioni struct {
	TruccanettiCons         []IncantesimoID     `json:"trucchetti_cons"         bson:"trucchetti_cons"`
	IncantesimiInizialiCons []IncantesimoID     `json:"incantesimi_iniziali_cons" bson:"incantesimi_iniziali_cons"`
	EquipInizialecons       []EquipaggiamentoID `json:"equip_iniziale_cons"     bson:"equip_iniziale_cons"`
	TalentiCons             []TalentoID         `json:"talenti_cons"            bson:"talenti_cons"`
	DonoEpicoCons           []TalentoID         `json:"dono_epico_cons"         bson:"dono_epico_cons"`
}

// ---------- Entit? ----------

type Classe struct {
	ID                             ClasseID                 `json:"id"                               bson:"_id"`
	Slug                           Slug                     `json:"slug"                             bson:"slug"`
	Nome                           string                   `json:"nome"                             bson:"nome"`
	Sottotitolo                    string                   `json:"sottotitolo"                      bson:"sottotitolo"`
	Markdown                       string                   `json:"markdown"                         bson:"markdown"`
	DadoVita                       Dadi                     `json:"dado_vita"                        bson:"dado_vita"`
	CaratteristicaPrimaria         []Caratteristica         `json:"caratteristica_primaria"          bson:"caratteristica_primaria"`
	SalvezzeCompetenze             []NomeCaratteristica     `json:"salvezze_competenze"              bson:"salvezze_competenze"`
	AbilitaCompetenzeOpzioni       Scelta                   `json:"abilita_competenze_opzioni"       bson:"abilita_competenze_opzioni"`
	ArmiCompetenze                 []string                 `json:"armi_competenze"                  bson:"armi_competenze"`
	ArmatureCompetenze             []CompetenzaArmatura     `json:"armature_competenze"              bson:"armature_competenze"`
	StrumentiCompetenze            []StrumentoID            `json:"strumenti_competenze"             bson:"strumenti_competenze"`
	EquipaggiamentoInizialeOpzioni []EquipaggiamentoOpzione `json:"equipaggiamento_iniziale_opzioni" bson:"equipaggiamento_iniziale_opzioni"`
	Multiclasse                    Multiclasse              `json:"multiclasse"                      bson:"multiclasse"`
	Progressioni                   Progressioni             `json:"progressioni"                     bson:"progressioni"`
	Magia                          Magia                    `json:"magia"                            bson:"magia"`
	PrivilegiDiClasse              []Privilegio             `json:"privilegi_di_classe"              bson:"privilegi_di_classe"`
	Sottoclassi                    []Sottoclasse            `json:"sottoclassi"                      bson:"sottoclassi"`
	Raccomandazioni                Raccomandazioni          `json:"raccomandazioni"                  bson:"raccomandazioni"`
	Contenuto                      string                   `json:"contenuto"                        bson:"contenuto"`
}

// ---------- Costruttore ----------

func NewClasse(
	id uuid.UUID,
	nome string,
	sottotitolo string,
	markdown string,
	dadoVita Dadi,
	caratteristicaPrimaria []Caratteristica,
	salvezzeCompetenze []NomeCaratteristica,
	abilitaCompetenzeOpzioni Scelta,
	armiCompetenze []string,
	armatureCompetenze []CompetenzaArmatura,
	strumentiCompetenze []StrumentoID,
	equipaggiamentoInizialeOpzioni []EquipaggiamentoOpzione,
	multiclasse Multiclasse,
	progressioni Progressioni,
	magia Magia,
	privilegiDiClasse []Privilegio,
	sottoclassi []Sottoclasse,
	listeIncantesimi ListaIncantesimi,
	raccomandazioni Raccomandazioni,
	contenuto string,
) *Classe {
	slug, _ := NewSlug(nome)

	return &Classe{
		ID:                             ClasseID(id),
		Slug:                           slug,
		Nome:                           nome,
		Sottotitolo:                    sottotitolo,
		Markdown:                       markdown,
		DadoVita:                       dadoVita,
		CaratteristicaPrimaria:         caratteristicaPrimaria,
		SalvezzeCompetenze:             salvezzeCompetenze,
		AbilitaCompetenzeOpzioni:       abilitaCompetenzeOpzioni,
		ArmiCompetenze:                 armiCompetenze,
		ArmatureCompetenze:             armatureCompetenze,
		StrumentiCompetenze:            strumentiCompetenze,
		EquipaggiamentoInizialeOpzioni: equipaggiamentoInizialeOpzioni,
		Multiclasse:                    multiclasse,
		Progressioni:                   progressioni,
		Magia:                          magia,
		PrivilegiDiClasse:              privilegiDiClasse,
		Sottoclassi:                    sottoclassi,
		Raccomandazioni:                raccomandazioni,
		Contenuto:                      contenuto,
	}
}
