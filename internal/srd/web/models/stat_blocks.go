package models

// SpellStatBlock holds the view model for rendering a spell stat block.
type SpellStatBlock struct {
	Level              int
	School             string
	CastingTime        string
	Range              string
	Components         string
	Duration           string
	Ritual             bool
	DescriptionHTML    string
	AtHigherLevelsHTML string
	Classes            string
}

// AbilityScore holds a single ability score row for a monster stat block.
type AbilityScore struct {
	Label string // FOR, DES, COS, INT, SAG, CAR
	Score int
	Mod   int
	Save  string
}

// Feature holds a named feature with pre-rendered HTML description.
type Feature struct {
	Name            string
	DescriptionHTML string
}

// FeatureSection holds a named section of features (e.g., "Azioni", "Reazioni").
type FeatureSection struct {
	Heading  string
	Features []Feature
}

// MonsterStatBlock holds the view model for rendering a monster stat block.
type MonsterStatBlock struct {
	Subtitle                string // e.g. "Aberrazione Grande, legale malvagio"
	AC                      string
	Initiative              string
	HP                      string
	Speed                   string
	AbilityScores           []AbilityScore
	Skills                  string
	ResistancesHTML         string
	DamageImmunitiesHTML    string
	ConditionImmunitiesHTML string
	Senses                  string
	Languages               string
	CR                      string
	Equipment               string
	Traits                  []Feature
	Sections                []FeatureSection // Actions, Bonus Actions, Reactions, Legendary Actions
}

// ClassFeature holds a class feature with its level.
type ClassFeature struct {
	Name            string
	Level           int
	DescriptionHTML string
}

// Subclass holds a subclass with its features.
type Subclass struct {
	Name            string
	DescriptionHTML string
	Features        []ClassFeature
}

// ClassStatBlock holds the view model for rendering a class stat block.
type ClassStatBlock struct {
	DescriptionHTML   string
	ProficienciesHTML string
	Features          []ClassFeature
	Subclasses        []Subclass
}

// SpeciesStatBlock holds the view model for rendering a species stat block.
type SpeciesStatBlock struct {
	Subtitle        string // e.g. "Umanoide Medio, 9 m"
	DescriptionHTML string
}
