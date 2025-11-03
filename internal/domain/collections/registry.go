package collections

// CollectionName represents a typed collection name for type safety
type CollectionName string

// Collection name constants - single source of truth for all collection names
const (
	Armature           CollectionName = "armature"
	Classi             CollectionName = "classi"
	Armi               CollectionName = "armi"
	Animali            CollectionName = "animali"
	Backgrounds        CollectionName = "backgrounds"
	Incantesimi        CollectionName = "incantesimi"
	Talenti            CollectionName = "talenti"
	Equipaggiamenti    CollectionName = "equipaggiamenti"
	Servizi            CollectionName = "servizi"
	Strumenti          CollectionName = "strumenti"
	Regole             CollectionName = "regole"
	CavalcatureVeicoli CollectionName = "cavalcature_veicoli"
	OggettiMagici      CollectionName = "oggetti_magici"
	Mostri             CollectionName = "mostri"
)

// CollectionInfo contains metadata about a collection
type CollectionInfo struct {
	Name           CollectionName
	Title          string
	HasNestedValue bool // True if documents use {"value": {...}} structure
}

// Registry provides centralized collection metadata
var Registry = map[CollectionName]CollectionInfo{
	Armature:           {Name: Armature, Title: "Armature", HasNestedValue: true},
	Classi:             {Name: Classi, Title: "Classi", HasNestedValue: true},
	Armi:               {Name: Armi, Title: "Armi", HasNestedValue: true},
	Animali:            {Name: Animali, Title: "Animali", HasNestedValue: false},
	Backgrounds:        {Name: Backgrounds, Title: "Background", HasNestedValue: true},
	Incantesimi:        {Name: Incantesimi, Title: "Incantesimi", HasNestedValue: false},
	Talenti:            {Name: Talenti, Title: "Talenti", HasNestedValue: true},
	Equipaggiamenti:    {Name: Equipaggiamenti, Title: "Equipaggiamento", HasNestedValue: false},
	Servizi:            {Name: Servizi, Title: "Servizi", HasNestedValue: true},
	Strumenti:          {Name: Strumenti, Title: "Strumenti", HasNestedValue: true},
	Regole:             {Name: Regole, Title: "Regole", HasNestedValue: false},
	CavalcatureVeicoli: {Name: CavalcatureVeicoli, Title: "Cavalcature e Veicoli", HasNestedValue: false},
	OggettiMagici:      {Name: OggettiMagici, Title: "Oggetti Magici", HasNestedValue: true},
	Mostri:             {Name: Mostri, Title: "Mostri", HasNestedValue: false},
}

// String returns the string representation of the collection name
func (c CollectionName) String() string {
	return string(c)
}

// GetInfo returns the collection info for a given collection name
func GetInfo(name string) (CollectionInfo, bool) {
	info, exists := Registry[CollectionName(name)]
	return info, exists
}

// GetTitle returns the display title for a collection
func GetTitle(name string) string {
	if info, exists := GetInfo(name); exists {
		return info.Title
	}
	return name
}

// HasNestedValue returns whether a collection uses nested value structure
func HasNestedValue(name string) bool {
	if info, exists := GetInfo(name); exists {
		return info.HasNestedValue
	}
	return false
}

// IsValid checks if a collection name is valid
func IsValid(name string) bool {
	_, exists := Registry[CollectionName(name)]
	return exists
}

// GetAllCollections returns all collection names
func GetAllCollections() []CollectionName {
	collections := make([]CollectionName, 0, len(Registry))
	for name := range Registry {
		collections = append(collections, name)
	}
	return collections
}

// GetAllWithInfo returns all collections with their info
func GetAllWithInfo() map[CollectionName]CollectionInfo {
	return Registry
}
