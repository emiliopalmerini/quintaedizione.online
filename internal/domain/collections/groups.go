package collections

// CollectionGroup defines a semantic grouping of collections for display.
type CollectionGroup struct {
	Label       string
	Collections []CollectionName
}

// GetGroups returns collection groups in display order.
func GetGroups() []CollectionGroup {
	return []CollectionGroup{
		{
			Label:       "Personaggi",
			Collections: []CollectionName{Classi, Specie, Backgrounds, Talenti},
		},
		{
			Label:       "Magia & Mostri",
			Collections: []CollectionName{Incantesimi, Mostri},
		},
		{
			Label:       "Equipaggiamento",
			Collections: []CollectionName{Equipaggiamenti, OggettiMagici},
		},
		{
			Label:       "Riferimento",
			Collections: []CollectionName{Regole, Servizi, Glossario},
		},
	}
}
