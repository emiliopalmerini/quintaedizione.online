package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// SpellSchool represents schools of magic
type SpellSchool string

const (
	SpellSchoolAbjuration    SpellSchool = "Abiurazione"
	SpellSchoolConjuration   SpellSchool = "Evocazione"
	SpellSchoolDivination    SpellSchool = "Divinazione"
	SpellSchoolEnchantment   SpellSchool = "Ammaliamento"
	SpellSchoolEvocation     SpellSchool = "Invocazione"
	SpellSchoolIllusion      SpellSchool = "Illusione"
	SpellSchoolNecromancy    SpellSchool = "Necromanzia"
	SpellSchoolTransmutation SpellSchool = "Trasmutazione"
)

// CastingTime represents spell casting times
type CastingTime string

const (
	CastingTimeAction      CastingTime = "1 azione"
	CastingTimeBonusAction CastingTime = "1 azione bonus"
	CastingTimeReaction    CastingTime = "1 reazione"
	CastingTimeRitual      CastingTime = "10 minuti (rituale)"
	CastingTimeMinute1     CastingTime = "1 minuto"
	CastingTimeMinute10    CastingTime = "10 minuti"
	CastingTimeHour1       CastingTime = "1 ora"
	CastingTimeHour8       CastingTime = "8 ore"
	CastingTimeHour24      CastingTime = "24 ore"
)

// SpellRange represents spell ranges
type SpellRange string

const (
	SpellRangeSelf       SpellRange = "Personale"
	SpellRangeTouch      SpellRange = "Contatto"
	SpellRangeFeet30     SpellRange = "9 metri"
	SpellRangeFeet60     SpellRange = "18 metri"
	SpellRangeFeet90     SpellRange = "27 metri"
	SpellRangeFeet120    SpellRange = "36 metri"
	SpellRangeFeet150    SpellRange = "45 metri"
	SpellRangeFeet300    SpellRange = "90 metri"
	SpellRangeFeet500    SpellRange = "150 metri"
	SpellRangeFeet1000   SpellRange = "300 metri"
	SpellRangeMile1      SpellRange = "1,5 chilometri"
	SpellRangeUnlimited  SpellRange = "Illimitata"
	SpellRangeSight      SpellRange = "A vista"
)

// SpellDuration represents spell durations
type SpellDuration string

const (
	SpellDurationInstantaneous SpellDuration = "Istantanea"
	SpellDurationRound1        SpellDuration = "1 round"
	SpellDurationMinute1       SpellDuration = "1 minuto"
	SpellDurationMinute10      SpellDuration = "10 minuti"
	SpellDurationHour1         SpellDuration = "1 ora"
	SpellDurationHour8         SpellDuration = "8 ore"
	SpellDurationHour24        SpellDuration = "24 ore"
	SpellDurationDay7          SpellDuration = "7 giorni"
	SpellDurationDay30         SpellDuration = "30 giorni"
	SpellDurationPermanent     SpellDuration = "Permanente"
	SpellDurationConcentration SpellDuration = "Concentrazione"
)

// SpellID represents a spell entity identifier
type SpellID struct {
	value string
}

// NewSpellID creates a new SpellID with validation
func NewSpellID(value string) (SpellID, error) {
	if strings.TrimSpace(value) == "" {
		return SpellID{}, errors.New("SpellID cannot be empty")
	}
	
	matched, err := regexp.MatchString(`^[a-z][a-z0-9-]*$`, value)
	if err != nil || !matched {
		return SpellID{}, fmt.Errorf("invalid spell ID format: %s", value)
	}
	
	return SpellID{value: value}, nil
}

// Value returns the string value of the SpellID
func (s SpellID) Value() string {
	return s.value
}

// SpellLevel represents a spell level (0-9)
type SpellLevel struct {
	value int
}

// NewSpellLevel creates a new SpellLevel with validation
func NewSpellLevel(value int) (SpellLevel, error) {
	if value < 0 || value > 9 {
		return SpellLevel{}, fmt.Errorf("spell level must be 0-9, got %d", value)
	}
	return SpellLevel{value: value}, nil
}

// Value returns the int value of the SpellLevel
func (s SpellLevel) Value() int {
	return s.value
}

// IsCantrip returns true if this is a cantrip (level 0)
func (s SpellLevel) IsCantrip() bool {
	return s.value == 0
}

// SpellComponent represents a spell casting component
type SpellComponent struct {
	Type        string  // "V", "S", "M"
	Description *string // Optional description
	CostGP      *int    // Optional gold cost
	Consumed    bool    // Whether component is consumed
}

// NewSpellComponent creates a new SpellComponent with validation
func NewSpellComponent(componentType string, description *string, costGP *int, consumed bool) (SpellComponent, error) {
	validTypes := map[string]bool{"V": true, "S": true, "M": true}
	if !validTypes[componentType] {
		return SpellComponent{}, fmt.Errorf("invalid component type: %s", componentType)
	}
	
	return SpellComponent{
		Type:        componentType,
		Description: description,
		CostGP:      costGP,
		Consumed:    consumed,
	}, nil
}

// SpellCasting represents spell casting information
type SpellCasting struct {
	Tempo          CastingTime
	Gittata        SpellRange
	Durata         SpellDuration
	GittataCustom  *string
	Componenti     []SpellComponent
	DurataCustom   *string
	Concentrazione bool
	Rituale        bool
}

// GetRangeText returns the range text (custom or enum)
func (s SpellCasting) GetRangeText() string {
	if s.GittataCustom != nil {
		return *s.GittataCustom
	}
	return string(s.Gittata)
}

// GetDurationText returns the duration text with concentration if applicable
func (s SpellCasting) GetDurationText() string {
	baseDuration := string(s.Durata)
	if s.DurataCustom != nil {
		baseDuration = *s.DurataCustom
	}
	
	if s.Concentrazione {
		return fmt.Sprintf("Concentrazione, fino a %s", strings.ToLower(baseDuration))
	}
	return baseDuration
}

// GetComponentsText returns formatted components text
func (s SpellCasting) GetComponentsText() string {
	var components []string
	for _, comp := range s.Componenti {
		if comp.Description != nil {
			components = append(components, fmt.Sprintf("%s (%s)", comp.Type, *comp.Description))
		} else {
			components = append(components, comp.Type)
		}
	}
	return strings.Join(components, ", ")
}

// Spell represents a D&D 5e spell entity
type Spell struct {
	ID                 SpellID
	Nome               string
	Livello            SpellLevel
	Scuola             SpellSchool
	Classi             []string
	Lancio             SpellCasting
	Descrizione        string
	ContenutoMarkdown  string
	
	// Optional fields
	Sottoscuole       []string
	LivelliSuperiori  *string
	Fonte             string // defaults to "SRD"
	Versione          string // defaults to "1.0"
}

// NewSpell creates a new Spell with validation
func NewSpell(
	id SpellID,
	nome string,
	livello SpellLevel,
	scuola SpellSchool,
	classi []string,
	lancio SpellCasting,
	descrizione string,
	contenutoMarkdown string,
) (Spell, error) {
	if strings.TrimSpace(nome) == "" {
		return Spell{}, errors.New("spell name cannot be empty")
	}
	if len(classi) == 0 {
		return Spell{}, errors.New("spell must have at least one class")
	}
	if strings.TrimSpace(descrizione) == "" {
		return Spell{}, errors.New("spell description cannot be empty")
	}
	
	return Spell{
		ID:                id,
		Nome:              nome,
		Livello:           livello,
		Scuola:            scuola,
		Classi:            classi,
		Lancio:            lancio,
		Descrizione:       descrizione,
		ContenutoMarkdown: contenutoMarkdown,
		Sottoscuole:       []string{},
		Fonte:             "SRD",
		Versione:          "1.0",
	}, nil
}

// IsCantrip returns true if this spell is a cantrip
func (s Spell) IsCantrip() bool {
	return s.Livello.IsCantrip()
}

// IsRitual returns true if this spell can be cast as a ritual
func (s Spell) IsRitual() bool {
	return s.Lancio.Rituale
}

// RequiresConcentration returns true if this spell requires concentration
func (s Spell) RequiresConcentration() bool {
	return s.Lancio.Concentrazione
}

// HasMaterialComponent returns true if this spell has material components
func (s Spell) HasMaterialComponent() bool {
	for _, comp := range s.Lancio.Componenti {
		if comp.Type == "M" {
			return true
		}
	}
	return false
}

// GetExpensiveComponents returns components with gold cost
func (s Spell) GetExpensiveComponents() []SpellComponent {
	var expensive []SpellComponent
	for _, comp := range s.Lancio.Componenti {
		if comp.Type == "M" && comp.CostGP != nil {
			expensive = append(expensive, comp)
		}
	}
	return expensive
}

// IsAvailableToClass checks if spell is available to specific class
func (s Spell) IsAvailableToClass(className string) bool {
	for _, class := range s.Classi {
		if class == className {
			return true
		}
	}
	return false
}

// ToDict converts to map for serialization
func (s Spell) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":                     s.ID.Value(),
		"nome":                   s.Nome,
		"livello":                s.Livello.Value(),
		"scuola":                 string(s.Scuola),
		"is_cantrip":             s.IsCantrip(),
		"is_ritual":              s.IsRitual(),
		"requires_concentration": s.RequiresConcentration(),
		"classi":                 s.Classi,
		"fonte":                  s.Fonte,
		"versione":               s.Versione,
	}
}