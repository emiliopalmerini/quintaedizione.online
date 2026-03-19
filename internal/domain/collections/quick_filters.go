package collections

// QuickFilterChip represents a single tappable chip in the quick filter bar.
// A chip may map to one or more filter values (e.g., CR range "0-¼" maps to "0", "1/8", "1/4").
type QuickFilterChip struct {
	Label  string
	Values []string
}

// QuickFilter defines the quick filter configuration for a collection.
type QuickFilter struct {
	FilterName string
	Chips      []QuickFilterChip
}

var quickFilters = map[CollectionName]QuickFilter{
	Incantesimi: {
		FilterName: "livello",
		Chips: []QuickFilterChip{
			{Label: "Trucchetto", Values: []string{"0"}},
			{Label: "1°", Values: []string{"1"}},
			{Label: "2°", Values: []string{"2"}},
			{Label: "3°", Values: []string{"3"}},
			{Label: "4°", Values: []string{"4"}},
			{Label: "5°", Values: []string{"5"}},
			{Label: "6°", Values: []string{"6"}},
			{Label: "7°", Values: []string{"7"}},
			{Label: "8°", Values: []string{"8"}},
			{Label: "9°", Values: []string{"9"}},
		},
	},
	Mostri: {
		FilterName: "grado_sfida",
		Chips: []QuickFilterChip{
			{Label: "0-¼", Values: []string{"0", "1/8", "1/4"}},
			{Label: "½-1", Values: []string{"1/2", "1"}},
			{Label: "2-4", Values: []string{"2", "3", "4"}},
			{Label: "5-8", Values: []string{"5", "6", "7", "8"}},
			{Label: "9-12", Values: []string{"9", "10", "11", "12"}},
			{Label: "13-16", Values: []string{"13", "14", "15", "16"}},
			{Label: "17-20", Values: []string{"17", "19", "20"}},
			{Label: "21+", Values: []string{"21", "22", "23", "24", "30"}},
		},
	},
	Equipaggiamenti: {
		FilterName: "categoria",
		Chips: []QuickFilterChip{
			{Label: "Armi semplici", Values: []string{"Armi da mischia semplici", "Armi a distanza semplici"}},
			{Label: "Armi da guerra", Values: []string{"Armi da mischia da guerra", "Armi a distanza da guerra"}},
			{Label: "Armature", Values: []string{"Armatura leggera", "Armatura media", "Armatura pesante", "Scudo"}},
			{Label: "Equipaggiamento", Values: []string{"Equipaggiamento d'avventura"}},
			{Label: "Strumenti", Values: []string{"Strumenti da artigiano", "Altri strumenti"}},
			{Label: "Veicoli", Values: []string{"Cavalcature e altri animali", "Finimenti e veicoli da tiro", "Veicoli aerei e imbarcazioni"}},
		},
	},
	OggettiMagici: {
		FilterName: "rarita",
		Chips: []QuickFilterChip{
			{Label: "Comune", Values: []string{"Comune"}},
			{Label: "Non Comune", Values: []string{"Non Comune"}},
			{Label: "Raro", Values: []string{"Raro"}},
			{Label: "Molto Raro", Values: []string{"Molto Raro"}},
			{Label: "Leggendario", Values: []string{"Leggendario"}},
		},
	},
}

// GetQuickFilter returns the quick filter configuration for a collection, if defined.
func GetQuickFilter(name CollectionName) (QuickFilter, bool) {
	qf, ok := quickFilters[name]
	return qf, ok
}
