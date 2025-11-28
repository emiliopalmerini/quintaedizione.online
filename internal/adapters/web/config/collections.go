package config

var CollectionTitles = map[string]string{
	"incantesimi":         "Incantesimi",
	"mostri":              "Mostri",
	"classi":              "Classi",
	"backgrounds":         "Background",
	"equipaggiamenti":     "Equipaggiamento",
	"armi":                "Armi",
	"armature":            "Armature",
	"oggetti_magici":      "Oggetti Magici",
	"talenti":             "Talenti",
	"servizi":             "Servizi",
	"strumenti":           "Strumenti",
	"animali":             "Animali",
	"regole":              "Regole",
	"cavalcature_veicoli": "Cavalcature e Veicoli",
}

func GetCollectionTitle(collection string) string {
	if title, exists := CollectionTitles[collection]; exists {
		return title
	}
	return collection
}
