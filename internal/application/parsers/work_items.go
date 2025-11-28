package parsers

func CreateDefaultWorkItems() []WorkItem {
	return []WorkItem{
		{
			Filename:   "ita/lists/armature.md",
			Collection: "armature",
			Language:   Italian,
		},
		{
			Filename:   "ita/lists/armi.md",
			Collection: "armi",
			Language:   Italian,
		},
		{
			Filename:   "ita/lists/animali.md",
			Collection: "animali",
			Language:   Italian,
		},
		{
			Filename:   "ita/lists/backgrounds.md",
			Collection: "backgrounds",
			Language:   Italian,
		},
		{
			Filename:   "ita/lists/cavalcature_veicoli_items.md",
			Collection: "cavalcature_veicoli",
			Language:   Italian,
		},
		{
			Filename:   "ita/lists/classi.md",
			Collection: "classi",
			Language:   Italian,
		},
		{
			Filename:   "ita/lists/equipaggiamenti.md",
			Collection: "equipaggiamenti",
			Language:   Italian,
		},
		{
			Filename:   "ita/lists/incantesimi.md",
			Collection: "incantesimi",
			Language:   Italian,
		},
		{
			Filename:   "ita/lists/mostri.md",
			Collection: "mostri",
			Language:   Italian,
		},
		{
			Filename:   "ita/lists/oggetti_magici.md",
			Collection: "oggetti_magici",
			Language:   Italian,
		},
		{
			Filename:   "ita/lists/regole.md",
			Collection: "regole",
			Language:   Italian,
		},
		{
			Filename:   "ita/lists/servizi.md",
			Collection: "servizi",
			Language:   Italian,
		},
		{
			Filename:   "ita/lists/strumenti.md",
			Collection: "strumenti",
			Language:   Italian,
		},
		{
			Filename:   "ita/lists/talenti.md",
			Collection: "talenti",
			Language:   Italian,
		},
	}
}

func GetWorkItemsForCollection(collections []string) []WorkItem {
	if len(collections) == 0 {
		return CreateDefaultWorkItems()
	}

	collectionMap := make(map[string]bool)
	for _, collection := range collections {
		collectionMap[collection] = true
	}

	var filtered []WorkItem
	allItems := CreateDefaultWorkItems()

	for _, item := range allItems {
		if collectionMap[item.Collection] {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

func GetAllWorkItems() []WorkItem {
	items := CreateDefaultWorkItems()
	return items
}

func ValidateWorkItem(item WorkItem) error {
	if item.Filename == "" {
		return ErrInvalidContext
	}
	if item.Collection == "" {
		return ErrInvalidContext
	}
	if item.Language == "" {
		return ErrInvalidContext
	}
	return nil
}
