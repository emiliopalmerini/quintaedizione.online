package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// ---------- ID ----------

type (
	ClasseDifficolta int
)

// ---------- Enum / VO di supporto ----------

const (
	CDMinima  ClasseDifficolta = 5
	CDMassima ClasseDifficolta = 30
)

// Validazione per ClasseDifficolta
func NewClasseDifficolta(cd int) (ClasseDifficolta, error) {
	if cd < int(CDMinima) || cd > int(CDMassima) {
		return 0, fmt.Errorf("classe difficoltà %d non valida: deve essere tra %d e %d", cd, CDMinima, CDMassima)
	}
	return ClasseDifficolta(cd), nil
}

// Utilizzo strumento
type UtilizzoStrumento struct {
	Descrizione string           `json:"descrizione" bson:"descrizione"`
	CD          ClasseDifficolta `json:"cd"          bson:"cd"` // Classe Difficoltà
}

// Costruttore per UtilizzoStrumento con validazione CD
func NewUtilizzoStrumento(descrizione string, cd int) (UtilizzoStrumento, error) {
	if descrizione == "" {
		return UtilizzoStrumento{}, fmt.Errorf("descrizione utilizzo strumento non può essere vuota")
	}

	classeDifficolta, err := NewClasseDifficolta(cd)
	if err != nil {
		return UtilizzoStrumento{}, fmt.Errorf("errore nella classe difficoltà: %w", err)
	}

	return UtilizzoStrumento{
		Descrizione: descrizione,
		CD:          classeDifficolta,
	}, nil
}

// ---------- Entità ----------

type Strumento struct {
	ID               StrumentoID         `json:"id"                bson:"_id"`
	Slug             Slug                `json:"slug"              bson:"slug"`
	Nome             string              `json:"nome"              bson:"nome"`
	Costo            Costo               `json:"costo"             bson:"costo"`
	Peso             Peso                `json:"peso"              bson:"peso"`
	AbilitaAssociata AbilitaID           `json:"abilita_associata" bson:"abilita_associata"`
	Utilizzi         []UtilizzoStrumento `json:"utilizzi"          bson:"utilizzi"`
	Creazioni        []string            `json:"creazioni"         bson:"creazioni"`
	Contenuto        string              `json:"contenuto"         bson:"contenuto"`
}

// ---------- Costruttore ----------

func NewStrumento(
	id uuid.UUID,
	nome string,
	costo Costo,
	peso Peso,
	abilitaAssociata uuid.UUID,
	utilizzi []UtilizzoStrumento,
	creazioni []string,
	contenuto string,
) *Strumento {
	slug, _ := NewSlug(nome)
	abilita := AbilitaID(abilitaAssociata)

	return &Strumento{
		ID:               StrumentoID(id),
		Slug:             slug,
		Nome:             nome,
		Costo:            costo,
		Peso:             peso,
		AbilitaAssociata: abilita,
		Utilizzi:         utilizzi,
		Creazioni:        creazioni,
		Contenuto:        contenuto,
	}
}
