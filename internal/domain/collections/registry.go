package collections

type CollectionName string

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

type CollectionInfo struct {
	Name           CollectionName
	Title          string
	HasNestedValue bool
}

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

func GetAllWithInfo() map[CollectionName]CollectionInfo {
	return Registry
}
