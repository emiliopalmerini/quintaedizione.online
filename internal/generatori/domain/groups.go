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
		ID:          "core-adventure",
		Label:       "Generatore di Avventure",
		Description: "Genera avventure complete per livelli 1°–4°",
		Order:       1,
	},
	{
		ID:          "npc-generator",
		Label:       "Generatore di PNG",
		Description: "Crea rapidamente PNG con personalità uniche",
		Order:       2,
	},
	{
		ID:          "treasure-generator",
		Label:       "Generatore di Tesori",
		Description: "Monete, gemme e reliquie per i tuoi dungeon",
		Order:       3,
	},
	{
		ID:          "random-traps",
		Label:       "Generatore di Trappole",
		Description: "Trappole semplici o complesse con innesco e variante",
		Order:       4,
	},
	{
		ID:          "random-chambers",
		Label:       "Stanze Casuali",
		Description: "Stanze per quindici ambientazioni di dungeon comuni",
		Order:       5,
	},
	{
		ID:          "random-items",
		Label:       "Oggetti Casuali",
		Description: "Reliquie, armi magiche e oggetti mondani unici",
		Order:       6,
	},
	{
		ID:          "town-events",
		Label:       "Eventi in Città",
		Description: "Cosa succede in città quando arrivano gli avventurieri",
		Order:       7,
	},
	{
		ID:          "dungeon-monsters",
		Label:       "Mostri Casuali del Dungeon",
		Description: "Mostri casuali organizzati per livello del dungeon",
		Order:       8,
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
