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
			Slug:        "magia-mostri",
			Label:       "Magia & Mostri",
			Description: "Incantesimi e bestiario per dare vita alle tue sessioni.",
			Collections: []CollectionName{Incantesimi, Mostri},
		},
		{
			Slug:        "equipaggiamento",
			Label:       "Equipaggiamento",
			Description: "Armi, armature, oggetti magici e tesori per gli avventurieri.",
			Collections: []CollectionName{Equipaggiamenti, OggettiMagici},
		},
		{
			Slug:        "riferimento",
			Label:       "Riferimento",
			Description: "Regole, servizi e glossario per consultazioni rapide.",
			Collections: []CollectionName{Regole, Servizi, Glossario},
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
