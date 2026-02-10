package collections

import "sort"

type CollectionName string

const (
	Classi          CollectionName = "classi"
	Backgrounds     CollectionName = "backgrounds"
	Incantesimi     CollectionName = "incantesimi"
	Talenti         CollectionName = "talenti"
	Equipaggiamenti CollectionName = "equipaggiamenti"
	Servizi         CollectionName = "servizi"
	Regole          CollectionName = "regole"
	OggettiMagici   CollectionName = "oggetti_magici"
	Mostri          CollectionName = "mostri"
	Glossario       CollectionName = "glossario"
	Specie          CollectionName = "specie"
)

type CollectionInfo struct {
	Name           CollectionName
	Title          string
	Description    string
	HasNestedValue bool
}

var Registry = map[CollectionName]CollectionInfo{
	Classi:          {Name: Classi, Title: "Classi", Description: "Le classi dei personaggi giocanti con caratteristiche, privilegi e sottoclassi.", HasNestedValue: true},
	Backgrounds:     {Name: Backgrounds, Title: "Background", Description: "I background dei personaggi con abilità, talenti e tratti caratteristici.", HasNestedValue: true},
	Incantesimi:     {Name: Incantesimi, Title: "Incantesimi", Description: "Tutti gli incantesimi dal trucchetto al 9° livello con descrizione e componenti.", HasNestedValue: false},
	Talenti:         {Name: Talenti, Title: "Talenti", Description: "I talenti disponibili per personalizzare e potenziare il tuo personaggio.", HasNestedValue: true},
	Equipaggiamenti: {Name: Equipaggiamenti, Title: "Equipaggiamento", Description: "Armi, armature, strumenti e altro equipaggiamento per gli avventurieri.", HasNestedValue: true},
	Servizi:         {Name: Servizi, Title: "Servizi", Description: "Servizi, cavalcature, veicoli e spese di sostentamento.", HasNestedValue: true},
	Regole:          {Name: Regole, Title: "Regole", Description: "Le regole base del gioco: combattimento, esplorazione e interazione.", HasNestedValue: false},
	OggettiMagici:   {Name: OggettiMagici, Title: "Oggetti Magici", Description: "Oggetti magici di ogni rarità: armi, armature, pozioni e oggetti meravigliosi.", HasNestedValue: true},
	Mostri:          {Name: Mostri, Title: "Mostri", Description: "Il bestiario completo con statistiche, abilità e gradi sfida.", HasNestedValue: false},
	Glossario:       {Name: Glossario, Title: "Glossario", Description: "Definizioni dei termini chiave e delle regole di gioco.", HasNestedValue: false},
	Specie:          {Name: Specie, Title: "Specie", Description: "Le specie giocabili con tratti, abilità speciali e caratteristiche.", HasNestedValue: true},
}

func (c CollectionName) String() string {
	return string(c)
}

func GetInfo(name string) (CollectionInfo, bool) {
	info, exists := Registry[CollectionName(name)]
	return info, exists
}

func GetTitle(name string) string {
	if info, exists := GetInfo(name); exists {
		return info.Title
	}
	return name
}

func HasNestedValue(name string) bool {
	if info, exists := GetInfo(name); exists {
		return info.HasNestedValue
	}
	return false
}

func IsValid(name string) bool {
	_, exists := Registry[CollectionName(name)]
	return exists
}

func GetAllCollections() []CollectionName {
	collections := make([]CollectionName, 0, len(Registry))
	for name := range Registry {
		collections = append(collections, name)
	}
	return collections
}

// GetAllCollectionNames returns collection names as strings in alphabetical order.
// This is useful for middleware, repositories, and other components that work with string collection names.
func GetAllCollectionNames() []string {
	names := make([]string, 0, len(Registry))
	for name := range Registry {
		names = append(names, name.String())
	}
	sort.Strings(names)
	return names
}

// FromString converts a string to CollectionName if it's valid.
// Returns the CollectionName and a boolean indicating if the conversion was successful.
func FromString(name string) (CollectionName, bool) {
	cn := CollectionName(name)
	_, exists := Registry[cn]
	return cn, exists
}

func GetAllWithInfo() map[CollectionName]CollectionInfo {
	return Registry
}
