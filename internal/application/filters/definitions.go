package filters

import (
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/collections"
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/filters"
)

// RegisterDefaultFilters populates the filter registry with predefined filter
// definitions for each collection that supports faceted filtering.
func RegisterDefaultFilters(registry *InMemoryFilterRegistry) {
	// Incantesimi
	registry.AddFilter(filters.FilterDefinition{
		Name:        "scuola",
		FieldPath:   "scuola",
		DataType:    filters.EnumFilter,
		Operator:    filters.ExactMatch,
		Collections: []collections.CollectionName{collections.Incantesimi},
		EnumValues:  []string{"Abiurazione", "Ammaliamento", "Divinazione", "Evocazione", "Illusione", "Invocazione", "Necromanzia", "Trasmutazione"},
		Description: "Scuola di magia",
	})
	registry.AddFilter(filters.FilterDefinition{
		Name:        "livello",
		FieldPath:   "livello",
		DataType:    filters.NumberFilter,
		Operator:    filters.ExactMatch,
		Collections: []collections.CollectionName{collections.Incantesimi},
		EnumValues:  []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"},
		Description: "Livello dell'incantesimo",
	})
	registry.AddFilter(filters.FilterDefinition{
		Name:        "classe",
		FieldPath:   "classe",
		DataType:    filters.StringFilter,
		Operator:    filters.RegexMatch,
		Collections: []collections.CollectionName{collections.Incantesimi},
		EnumValues:  []string{"Bardo", "Chierico", "Druido", "Guerriero (Cavaliere Mistico)", "Mago", "Monaco", "Paladino", "Ranger", "Stregone", "Warlock"},
		Description: "Classe che può lanciare l'incantesimo",
	})

	// Mostri
	registry.AddFilter(filters.FilterDefinition{
		Name:        "tipo",
		FieldPath:   "tipo",
		DataType:    filters.EnumFilter,
		Operator:    filters.ExactMatch,
		Collections: []collections.CollectionName{collections.Mostri},
		EnumValues:  []string{"Aberrazione", "Bestia", "Celestiale", "Costrutto", "Drago", "Elementale", "Folletto", "Gigante", "Immondo", "Melma", "Mostruosità", "Non morto", "Sciame", "Umanoide", "Vegetale"},
		Description: "Tipo di creatura",
	})
	registry.AddFilter(filters.FilterDefinition{
		Name:        "taglia",
		FieldPath:   "taglia",
		DataType:    filters.EnumFilter,
		Operator:    filters.ExactMatch,
		Collections: []collections.CollectionName{collections.Mostri},
		EnumValues:  []string{"Minuscola", "Piccola", "Media", "Grande", "Enorme", "Mastodontica"},
		Description: "Taglia della creatura",
	})
	registry.AddFilter(filters.FilterDefinition{
		Name:        "grado_sfida",
		FieldPath:   "grado_sfida",
		DataType:    filters.StringFilter,
		Operator:    filters.ExactMatch,
		Collections: []collections.CollectionName{collections.Mostri},
		EnumValues:  []string{"0", "1/8", "1/4", "1/2", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "19", "20", "21", "22", "23", "24", "30"},
		Description: "Grado di sfida",
	})

	// Oggetti Magici
	registry.AddFilter(filters.FilterDefinition{
		Name:        "rarita",
		FieldPath:   "rarita",
		DataType:    filters.EnumFilter,
		Operator:    filters.ExactMatch,
		Collections: []collections.CollectionName{collections.OggettiMagici},
		EnumValues:  []string{"Comune", "Non Comune", "Raro", "Molto Raro", "Leggendario"},
		Description: "Rarità dell'oggetto magico",
	})
	registry.AddFilter(filters.FilterDefinition{
		Name:        "tipo_oggetto",
		FieldPath:   "tipo",
		DataType:    filters.EnumFilter,
		Operator:    filters.ExactMatch,
		Collections: []collections.CollectionName{collections.OggettiMagici},
		EnumValues:  []string{"Armatura", "Arma", "Bacchetta", "Bastone", "Oggetto Meraviglioso", "Pozione", "Pergamena", "Anello", "Verga"},
		Description: "Tipo di oggetto magico",
	})

	// Equipaggiamenti
	registry.AddFilter(filters.FilterDefinition{
		Name:        "categoria",
		FieldPath:   "categoria",
		DataType:    filters.EnumFilter,
		Operator:    filters.ExactMatch,
		Collections: []collections.CollectionName{collections.Equipaggiamenti},
		EnumValues: []string{
			"Armi da mischia semplici", "Armi a distanza semplici", "Armi da mischia da guerra", "Armi a distanza da guerra",
			"Armatura leggera", "Armatura media", "Armatura pesante", "Scudo",
			"Equipaggiamento d'avventura",
			"Strumenti da artigiano", "Altri strumenti",
			"Cavalcature e altri animali", "Finimenti e veicoli da tiro", "Veicoli aerei e imbarcazioni",
		},
		Description: "Categoria dell'equipaggiamento",
	})

	// Glossario
	registry.AddFilter(filters.FilterDefinition{
		Name:        "categoria",
		FieldPath:   "categoria",
		DataType:    filters.EnumFilter,
		Operator:    filters.ExactMatch,
		Collections: []collections.CollectionName{collections.Glossario},
		EnumValues:  []string{"condizione", "azione", "atteggiamento"},
		Description: "Categoria del termine",
	})

	// Specie
	registry.AddFilter(filters.FilterDefinition{
		Name:        "tipo_creatura",
		FieldPath:   "tipo_creatura",
		DataType:    filters.EnumFilter,
		Operator:    filters.ExactMatch,
		Collections: []collections.CollectionName{collections.Specie},
		EnumValues:  []string{"umanoide", "fatato", "costrutto"},
		Description: "Tipo di creatura",
	})
}
