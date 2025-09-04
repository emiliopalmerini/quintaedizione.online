package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// WeaponCategory represents weapon categories
type WeaponCategory string

const (
	WeaponCategorySimpleMelee  WeaponCategory = "Semplice da Mischia"
	WeaponCategorySimpleRanged WeaponCategory = "Semplice a Distanza"
	WeaponCategoryMartialMelee WeaponCategory = "Da Guerra da Mischia"
	WeaponCategoryMartialRanged WeaponCategory = "Da Guerra a Distanza"
)

// WeaponProperty represents weapon properties
type WeaponProperty string

const (
	WeaponPropertyAccurate   WeaponProperty = "Accurata"
	WeaponPropertyAmmunition WeaponProperty = "Munizioni"
	WeaponPropertyFinesse    WeaponProperty = "Elegante"
	WeaponPropertyHeavy      WeaponProperty = "Pesante"
	WeaponPropertyLight      WeaponProperty = "Leggera"
	WeaponPropertyLoading    WeaponProperty = "Ricarica"
	WeaponPropertyRange      WeaponProperty = "Gittata"
	WeaponPropertyReach      WeaponProperty = "Portata"
	WeaponPropertySpecial    WeaponProperty = "Speciale"
	WeaponPropertyThrown     WeaponProperty = "Da Lancio"
	WeaponPropertyTwoHanded  WeaponProperty = "A Due Mani"
	WeaponPropertyVersatile  WeaponProperty = "Versatile"
)

// ArmorCategory represents armor categories
type ArmorCategory string

const (
	ArmorCategoryLight  ArmorCategory = "Leggera"
	ArmorCategoryMedium ArmorCategory = "Media"
	ArmorCategoryHeavy  ArmorCategory = "Pesante"
	ArmorCategoryShield ArmorCategory = "Scudo"
)

// MagicItemRarity represents magic item rarity
type MagicItemRarity string

const (
	MagicItemRarityCommon    MagicItemRarity = "Comune"
	MagicItemRarityUncommon  MagicItemRarity = "Non Comune"
	MagicItemRarityRare      MagicItemRarity = "Raro"
	MagicItemRarityVeryRare  MagicItemRarity = "Molto Raro"
	MagicItemRarityLegendary MagicItemRarity = "Leggendario"
	MagicItemRarityArtifact  MagicItemRarity = "Artefatto"
)

// EquipmentID represents equipment entity identifier
type EquipmentID struct {
	value string
}

// NewEquipmentID creates a new EquipmentID with validation
func NewEquipmentID(value string) (EquipmentID, error) {
	if strings.TrimSpace(value) == "" {
		return EquipmentID{}, errors.New("EquipmentID cannot be empty")
	}
	
	matched, err := regexp.MatchString(`^[a-z][a-z0-9-]*$`, value)
	if err != nil || !matched {
		return EquipmentID{}, fmt.Errorf("invalid equipment ID format: %s", value)
	}
	
	return EquipmentID{value: value}, nil
}

// Value returns the string value of the EquipmentID
func (e EquipmentID) Value() string {
	return e.value
}

// Currency represents currency amount in gold pieces
type Currency struct {
	goldPieces interface{} // Can be int, float64, or string
}

// NewCurrency creates a new Currency
func NewCurrency(goldPieces interface{}) (Currency, error) {
	switch v := goldPieces.(type) {
	case string:
		// Validate string format for cases like "2,000 mo" or "1/2 mo"
		matched, err := regexp.MatchString(`^[\d,./\s]+\s*(mo|ma|me|mr|mp)?$`, strings.ToLower(v))
		if err != nil || !matched {
			return Currency{}, fmt.Errorf("invalid currency format: %s", v)
		}
	case int, float64:
		// Valid numeric types
	default:
		return Currency{}, fmt.Errorf("currency must be int, float64, or string, got %T", goldPieces)
	}
	
	return Currency{goldPieces: goldPieces}, nil
}

// ToGold converts to gold pieces as float64
func (c Currency) ToGold() float64 {
	switch v := c.goldPieces.(type) {
	case int:
		return float64(v)
	case float64:
		return v
	case string:
		// Parse string format
		text := strings.ToLower(strings.TrimSpace(v))
		// Extract numeric part
		numericPart := regexp.MustCompile(`[^\d,./]`).ReplaceAllString(text, "")
		
		if strings.Contains(numericPart, "/") {
			parts := strings.Split(numericPart, "/")
			if len(parts) == 2 {
				num, _ := strconv.ParseFloat(parts[0], 64)
				den, _ := strconv.ParseFloat(parts[1], 64)
				if den != 0 {
					return num / den
				}
			}
		}
		
		numericPart = strings.ReplaceAll(numericPart, ",", "")
		result, _ := strconv.ParseFloat(numericPart, 64)
		return result
	}
	return 0
}

// GetText returns display text
func (c Currency) GetText() string {
	if str, ok := c.goldPieces.(string); ok {
		return str
	}
	return fmt.Sprintf("%v mo", c.goldPieces)
}

// Weight represents item weight in kilograms
type Weight struct {
	kilograms interface{} // Can be float64 or string
}

// NewWeight creates a new Weight
func NewWeight(kilograms interface{}) (Weight, error) {
	switch v := kilograms.(type) {
	case string:
		matched, err := regexp.MatchString(`^[\d,./\s]+\s*kg?$`, strings.ToLower(v))
		if err != nil || !matched {
			return Weight{}, fmt.Errorf("invalid weight format: %s", v)
		}
	case float64, int:
		// Valid numeric types
	default:
		return Weight{}, fmt.Errorf("weight must be float64, int, or string, got %T", kilograms)
	}
	
	return Weight{kilograms: kilograms}, nil
}

// ToKg converts to kilograms as float64
func (w Weight) ToKg() float64 {
	switch v := w.kilograms.(type) {
	case int:
		return float64(v)
	case float64:
		return v
	case string:
		// Parse string format
		numericPart := regexp.MustCompile(`[^\d,./]`).ReplaceAllString(v, "")
		if strings.Contains(numericPart, "/") {
			parts := strings.Split(numericPart, "/")
			if len(parts) == 2 {
				num, _ := strconv.ParseFloat(parts[0], 64)
				den, _ := strconv.ParseFloat(parts[1], 64)
				if den != 0 {
					return num / den
				}
			}
		}
		
		numericPart = strings.ReplaceAll(numericPart, ",", "")
		result, _ := strconv.ParseFloat(numericPart, 64)
		return result
	}
	return 0
}

// GetText returns display text
func (w Weight) GetText() string {
	if str, ok := w.kilograms.(string); ok {
		return str
	}
	return fmt.Sprintf("%v kg", w.kilograms)
}

// DamageInfo represents weapon damage information
type DamageInfo struct {
	Dice          string  // e.g., "1d8"
	DamageType    string  // e.g., "Perforante"
	VersatileDice *string // e.g., "1d10" for versatile
}

// NewDamageInfo creates a new DamageInfo with validation
func NewDamageInfo(dice, damageType string, versatileDice *string) (DamageInfo, error) {
	if strings.TrimSpace(dice) == "" {
		return DamageInfo{}, errors.New("damage dice cannot be empty")
	}
	if strings.TrimSpace(damageType) == "" {
		return DamageInfo{}, errors.New("damage type cannot be empty")
	}
	
	return DamageInfo{
		Dice:          dice,
		DamageType:    damageType,
		VersatileDice: versatileDice,
	}, nil
}

// GetDamageText returns damage display text
func (d DamageInfo) GetDamageText() string {
	baseDamage := fmt.Sprintf("%s %s", d.Dice, d.DamageType)
	if d.VersatileDice != nil {
		return fmt.Sprintf("%s (%s versatile)", baseDamage, *d.VersatileDice)
	}
	return baseDamage
}

// WeaponRange represents weapon range information
type WeaponRange struct {
	Normale *string // e.g., "6 m"
	Lunga   *string // e.g., "18 m"
}

// GetRangeText returns range display text
func (w WeaponRange) GetRangeText() string {
	if w.Normale != nil && w.Lunga != nil {
		return fmt.Sprintf("%s/%s", *w.Normale, *w.Lunga)
	} else if w.Normale != nil {
		return *w.Normale
	}
	return "—"
}

// Weapon represents a D&D 5e weapon entity
type Weapon struct {
	ID                EquipmentID
	Nome              string
	Costo             Currency
	Peso              Weight
	Danno             DamageInfo
	Categoria         WeaponCategory
	Proprieta         []WeaponProperty
	Maestria          string
	ContenutoMarkdown string
	
	// Optional fields
	Gittata     *WeaponRange
	Descrizione *string
	Fonte       string // defaults to "SRD"
	Versione    string // defaults to "1.0"
}

// NewWeapon creates a new Weapon with validation
func NewWeapon(
	id EquipmentID,
	nome string,
	costo Currency,
	peso Weight,
	danno DamageInfo,
	categoria WeaponCategory,
	proprieta []WeaponProperty,
	maestria string,
	contenutoMarkdown string,
) (Weapon, error) {
	if strings.TrimSpace(nome) == "" {
		return Weapon{}, errors.New("weapon name cannot be empty")
	}
	if strings.TrimSpace(maestria) == "" {
		return Weapon{}, errors.New("weapon mastery cannot be empty")
	}
	
	return Weapon{
		ID:                id,
		Nome:              nome,
		Costo:             costo,
		Peso:              peso,
		Danno:             danno,
		Categoria:         categoria,
		Proprieta:         proprieta,
		Maestria:          maestria,
		ContenutoMarkdown: contenutoMarkdown,
		Fonte:             "SRD",
		Versione:          "1.0",
	}, nil
}

// IsMelee checks if weapon is melee
func (w Weapon) IsMelee() bool {
	return w.Categoria == WeaponCategorySimpleMelee || w.Categoria == WeaponCategoryMartialMelee
}

// IsRanged checks if weapon is ranged
func (w Weapon) IsRanged() bool {
	return w.Categoria == WeaponCategorySimpleRanged || w.Categoria == WeaponCategoryMartialRanged
}

// IsMartial checks if weapon is martial
func (w Weapon) IsMartial() bool {
	return w.Categoria == WeaponCategoryMartialMelee || w.Categoria == WeaponCategoryMartialRanged
}

// HasProperty checks if weapon has specific property
func (w Weapon) HasProperty(prop WeaponProperty) bool {
	for _, p := range w.Proprieta {
		if p == prop {
			return true
		}
	}
	return false
}

// IsVersatile checks if weapon is versatile
func (w Weapon) IsVersatile() bool {
	return w.HasProperty(WeaponPropertyVersatile)
}

// ToDict converts to map for serialization
func (w Weapon) ToDict() map[string]interface{} {
	properties := make([]string, len(w.Proprieta))
	for i, prop := range w.Proprieta {
		properties[i] = string(prop)
	}
	
	return map[string]interface{}{
		"id":          w.ID.Value(),
		"nome":        w.Nome,
		"categoria":   string(w.Categoria),
		"is_melee":    w.IsMelee(),
		"is_ranged":   w.IsRanged(),
		"is_martial":  w.IsMartial(),
		"is_versatile": w.IsVersatile(),
		"proprieta":   properties,
		"fonte":       w.Fonte,
		"versione":    w.Versione,
	}
}

// Armor represents a D&D 5e armor entity
type Armor struct {
	ID                EquipmentID
	Nome              string
	Costo             Currency
	Peso              Weight
	ClasseArmatura    string // e.g., "11 + mod Des"
	Categoria         ArmorCategory
	ContenutoMarkdown string
	
	// Optional fields
	ForzaRichiesta        *int
	SvantaggioFurtivita   bool
	Descrizione           *string
	Fonte                 string // defaults to "SRD"
	Versione              string // defaults to "1.0"
}

// NewArmor creates a new Armor with validation
func NewArmor(
	id EquipmentID,
	nome string,
	costo Currency,
	peso Weight,
	classeArmatura string,
	categoria ArmorCategory,
	contenutoMarkdown string,
) (Armor, error) {
	if strings.TrimSpace(nome) == "" {
		return Armor{}, errors.New("armor name cannot be empty")
	}
	if strings.TrimSpace(classeArmatura) == "" {
		return Armor{}, errors.New("AC cannot be empty")
	}
	
	return Armor{
		ID:                id,
		Nome:              nome,
		Costo:             costo,
		Peso:              peso,
		ClasseArmatura:    classeArmatura,
		Categoria:         categoria,
		ContenutoMarkdown: contenutoMarkdown,
		Fonte:             "SRD",
		Versione:          "1.0",
	}, nil
}

// IsShield checks if this is a shield
func (a Armor) IsShield() bool {
	return a.Categoria == ArmorCategoryShield
}

// RequiresStrength checks if armor has strength requirement
func (a Armor) RequiresStrength() bool {
	return a.ForzaRichiesta != nil
}

// ImposesSteathDisadvantage checks if armor imposes stealth disadvantage
func (a Armor) ImposesSteathDisadvantage() bool {
	return a.SvantaggioFurtivita
}

// ToDict converts to map for serialization
func (a Armor) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":                             a.ID.Value(),
		"nome":                           a.Nome,
		"categoria":                      string(a.Categoria),
		"is_shield":                      a.IsShield(),
		"requires_strength":              a.RequiresStrength(),
		"imposes_stealth_disadvantage":   a.ImposesSteathDisadvantage(),
		"fonte":                          a.Fonte,
		"versione":                       a.Versione,
	}
}

// MagicItem represents a D&D 5e magic item entity
type MagicItem struct {
	ID                EquipmentID
	Nome              string
	Tipo              string // e.g., "Armor (Any Medium or Heavy, Except Hide Armor)"
	Rarita            MagicItemRarity
	Sintonizzazione   bool
	ContenutoMarkdown string
	
	// Optional fields
	Descrizione         *string
	CostoStimato        *Currency
	SlotEquipaggiamento *string // e.g., "Torso", "Mani"
	ScuolaMagica        *string
	Fonte               string // defaults to "SRD"
	Versione            string // defaults to "1.0"
}

// NewMagicItem creates a new MagicItem with validation
func NewMagicItem(
	id EquipmentID,
	nome string,
	tipo string,
	rarita MagicItemRarity,
	sintonizzazione bool,
	contenutoMarkdown string,
) (MagicItem, error) {
	if strings.TrimSpace(nome) == "" {
		return MagicItem{}, errors.New("magic item name cannot be empty")
	}
	if strings.TrimSpace(tipo) == "" {
		return MagicItem{}, errors.New("magic item type cannot be empty")
	}
	
	return MagicItem{
		ID:                id,
		Nome:              nome,
		Tipo:              tipo,
		Rarita:            rarita,
		Sintonizzazione:   sintonizzazione,
		ContenutoMarkdown: contenutoMarkdown,
		Fonte:             "SRD",
		Versione:          "1.0",
	}, nil
}

// RequiresAttunement checks if item requires attunement
func (m MagicItem) RequiresAttunement() bool {
	return m.Sintonizzazione
}

// IsConsumable checks if item is consumable (rough heuristic)
func (m MagicItem) IsConsumable() bool {
	consumableTypes := []string{"potion", "scroll", "pozione", "pergamena"}
	lowerType := strings.ToLower(m.Tipo)
	for _, ctype := range consumableTypes {
		if strings.Contains(lowerType, ctype) {
			return true
		}
	}
	return false
}

// GetRarityTier returns rarity as numeric tier (1-6)
func (m MagicItem) GetRarityTier() int {
	rarityTiers := map[MagicItemRarity]int{
		MagicItemRarityCommon:    1,
		MagicItemRarityUncommon:  2,
		MagicItemRarityRare:      3,
		MagicItemRarityVeryRare:  4,
		MagicItemRarityLegendary: 5,
		MagicItemRarityArtifact:  6,
	}
	if tier, ok := rarityTiers[m.Rarita]; ok {
		return tier
	}
	return 1
}

// ToDict converts to map for serialization
func (m MagicItem) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":                 m.ID.Value(),
		"nome":               m.Nome,
		"tipo":               m.Tipo,
		"rarita":             string(m.Rarita),
		"rarity_tier":        m.GetRarityTier(),
		"requires_attunement": m.RequiresAttunement(),
		"is_consumable":      m.IsConsumable(),
		"fonte":              m.Fonte,
		"versione":           m.Versione,
	}
}