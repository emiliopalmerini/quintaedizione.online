package domain

// GroupInfo defines metadata for a generator group.
type GroupInfo struct {
	ID          string
	Label       string
	Description string
	Order       int
}

// GroupRegistry maps group IDs to their metadata, ordered by display priority.
var GroupRegistry = []GroupInfo{
	{
		ID:    "core-adventure",
		Label: "Generatore di Avventure",
		Description: "Le tabelle di questa sezione possono aiutarti a generare un'avventura fantasy basata " +
			"sul concetto tradizionale di essere ingaggiati da un patrono o un altro PNG per intraprendere " +
			"una missione in un luogo specifico. Spesso queste avventure si svolgono in piccoli insediamenti " +
			"circondati da rovine antiche e tane di mostri ai confini della civiltà.\n\n" +
			"Usa queste tabelle insieme per generare e ispirare avventure complete, oppure usa le singole " +
			"tabelle per riempire i dettagli di altre avventure che crei o giochi. Questo generatore " +
			"(e in particolare la tabella dei Mostri del Dungeon e quella dei Tesori) è pensato per " +
			"personaggi dal 1° al 4° livello, ma può essere facilmente modificato per avventure di livello superiore.",
		Order: 1,
	},
	{
		ID:    "npc-generator",
		Label: "Generatore di PNG",
		Description: "I PNG danno vita ai mondi di gioco. Puoi usare i generatori di questa sezione per " +
			"creare rapidamente PNG da inserire nella tua partita, tirando sulle tabelle seguenti per " +
			"generare le caratteristiche di base. Per dare davvero vita al PNG, puoi poi modellare la " +
			"sua personalità e interpretazione ispirandoti a personaggi dei tuoi libri, serie TV o film " +
			"preferiti, cambiando genere e altri tratti per renderli originali.",
		Order: 2,
	},
}

// GetGroupInfo returns the metadata for a given group ID.
func GetGroupInfo(id string) (GroupInfo, bool) {
	for _, g := range GroupRegistry {
		if g.ID == id {
			return g, true
		}
	}
	return GroupInfo{}, false
}
