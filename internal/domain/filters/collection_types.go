package filters

type CollectionType string

const (
	IncantesimiCollection        CollectionType = "incantesimi"
	MostriCollection             CollectionType = "mostri"
	ClassiCollection             CollectionType = "classi"
	BackgroundsCollection        CollectionType = "backgrounds"
	EquipaggiamentiCollection    CollectionType = "equipaggiamenti"
	OggettiMagiciCollection      CollectionType = "oggetti_magici"
	ArmiCollection               CollectionType = "armi"
	ArmatureCollection           CollectionType = "armature"
	TalentiCollection            CollectionType = "talenti"
	ServiziCollection            CollectionType = "servizi"
	StrumentiCollection          CollectionType = "strumenti"
	AnimaliCollection            CollectionType = "animali"
	RegoleCollection             CollectionType = "regole"
	CavalcatureVeicoliCollection CollectionType = "cavalcature_veicoli"
)

func (c CollectionType) String() string {
	return string(c)
}

func (c CollectionType) IsValid() bool {
	validCollections := []CollectionType{
		IncantesimiCollection,
		MostriCollection,
		ClassiCollection,
		BackgroundsCollection,
		EquipaggiamentiCollection,
		OggettiMagiciCollection,
		ArmiCollection,
		ArmatureCollection,
		TalentiCollection,
		ServiziCollection,
		StrumentiCollection,
		AnimaliCollection,
		RegoleCollection,
		CavalcatureVeicoliCollection,
	}

	for _, valid := range validCollections {
		if c == valid {
			return true
		}
	}
	return false
}

func (c CollectionType) GetDisplayName() string {
	displayNames := map[CollectionType]string{
		IncantesimiCollection:        "Incantesimi",
		MostriCollection:             "Mostri",
		ClassiCollection:             "Classi",
		BackgroundsCollection:        "Background",
		EquipaggiamentiCollection:    "Equipaggiamento",
		OggettiMagiciCollection:      "Oggetti Magici",
		ArmiCollection:               "Armi",
		ArmatureCollection:           "Armature",
		TalentiCollection:            "Talenti",
		ServiziCollection:            "Servizi",
		StrumentiCollection:          "Strumenti",
		AnimaliCollection:            "Animali",
		RegoleCollection:             "Regole",
		CavalcatureVeicoliCollection: "Cavalcature e Veicoli",
	}

	if name, exists := displayNames[c]; exists {
		return name
	}
	return string(c)
}
