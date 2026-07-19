package collections

// CollectionGroup defines a semantic grouping of collections for display.
type CollectionGroup struct {
	Slug        string
	Label       string
	Description string
	Collections []CollectionName
}

// GetGroups returns collection groups in display order.
func GetGroups() []CollectionGroup {
	return []CollectionGroup{
		{
			Slug:        "personaggi",
			Label:       "Personaggi",
			Description: "Classi, specie, background e talenti per costruire il tuo personaggio.",
			Collections: []CollectionName{Classi, Specie, Backgrounds, Talenti},
		},
		{
			Slug:        "regole",
			Label:       "Regole",
			Description: "Regole di gioco e definizioni per una consultazione rapida.",
			Collections: []CollectionName{Regole, Glossario},
		},
		{
			Slug:        "magia",
			Label:       "Magia",
			Description: "Incantesimi dal trucchetto al 9° livello.",
			Collections: []CollectionName{Incantesimi},
		},
		{
			Slug:        "equipaggiamento",
			Label:       "Equipaggiamento",
			Description: "Equipaggiamento, oggetti magici e servizi per gli avventurieri.",
			Collections: []CollectionName{Equipaggiamenti, OggettiMagici, Servizi},
		},
		{
			Slug:        "bestiario",
			Label:       "Bestiario",
			Description: "Mostri e creature per popolare incontri e avventure.",
			Collections: []CollectionName{Mostri},
		},
	}
}

// GetGroup returns the group matching the given slug.
// The boolean is false if no group matches.
func GetGroup(slug string) (CollectionGroup, bool) {
	for _, g := range GetGroups() {
		if g.Slug == slug {
			return g, true
		}
	}
	return CollectionGroup{}, false
}
