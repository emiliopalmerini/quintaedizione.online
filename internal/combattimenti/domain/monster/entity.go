package monster

// AbilityScores holds the six core ability values.
type AbilityScores struct {
	Strength     int
	Dexterity    int
	Constitution int
	Intelligence int
	Wisdom       int
	Charisma     int
}

// SavingThrows holds saving throw modifiers as formatted strings (e.g. "+5").
type SavingThrows struct {
	Strength     string
	Dexterity    string
	Constitution string
	Intelligence string
	Wisdom       string
	Charisma     string
}

// NamedDescription pairs a name with a description (traits, actions, etc.).
type NamedDescription struct {
	Name        string
	Description string
}

// Monster represents a D&D creature with combat-relevant stats.
type Monster struct {
	ID   string
	Name string
	Type string
	Size string
	CR   string
	XP   int
	AC   string
	HP   string

	// Detail fields
	Group               string
	Subtype             string
	Alignment           string
	Initiative          string
	Speed               string
	AbilityScores       AbilityScores
	AbilityMods         AbilityScores
	SavingThrows        SavingThrows
	Skills              string
	Senses              string
	Languages           string
	Resistances         string
	DamageImmunities    string
	ConditionImmunities string
	Equipment           string
	CRDetail            string
	Traits              []NamedDescription
	Actions             []NamedDescription
	BonusActions        []NamedDescription
	Reactions           []NamedDescription
	LegendaryActions    []NamedDescription
}

// SearchFilters holds all possible filter criteria for monster search.
type SearchFilters struct {
	Query string
	MaxXP int
	Type  string
	Size  string
	CRMin string
	CRMax string
}

// Repository defines the interface for accessing monster data.
type Repository interface {
	FindByMaxXP(maxXP int) []Monster
	Search(query string, maxXP int) []Monster
	SearchWithFilters(filters SearchFilters) []Monster
	AvailableTypes() []string
	AvailableSizes() []string
	AvailableCRs() []string
}
