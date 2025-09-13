package parsers

// CreateDefaultWorkItems creates the default Italian work items configuration for the new strategy pattern
func CreateDefaultWorkItems() []WorkItem {
	return []WorkItem{
		{
			Filename:   "armature.md",
			Collection: "armature",
			Language:   Italian,
		},
		{
			Filename:   "armi.md",
			Collection: "armi",
			Language:   Italian,
		},
		{
			Filename:   "animali.md",
			Collection: "animali",
			Language:   Italian,
		},
		{
			Filename:   "backgrounds.md",
			Collection: "backgrounds",
			Language:   Italian,
		},
		{
			Filename:   "cavalcature_veicoli_items.md",
			Collection: "cavalcature_veicoli",
			Language:   Italian,
		},
		{
			Filename:   "classi.md",
			Collection: "classi",
			Language:   Italian,
		},
		{
			Filename:   "equipaggiamenti.md",
			Collection: "equipaggiamenti",
			Language:   Italian,
		},
		{
			Filename:   "incantesimi.md",
			Collection: "incantesimi",
			Language:   Italian,
		},
		{
			Filename:   "mostri.md",
			Collection: "mostri",
			Language:   Italian,
		},
		{
			Filename:   "oggetti_magici.md",
			Collection: "oggetti_magici",
			Language:   Italian,
		},
		{
			Filename:   "regole.md",
			Collection: "regole",
			Language:   Italian,
		},
		{
			Filename:   "servizi.md",
			Collection: "servizi",
			Language:   Italian,
		},
		{
			Filename:   "strumenti.md",
			Collection: "strumenti",
			Language:   Italian,
		},
		{
			Filename:   "talenti.md",
			Collection: "talenti",
			Language:   Italian,
		},
	}
}

// CreateDocumentWorkItems creates work items for documentation files

// GetWorkItemsForCollection returns work items filtered by collection name
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

// GetAllWorkItems returns all available work items (lists + documents)
func GetAllWorkItems() []WorkItem {
	items := CreateDefaultWorkItems()
	return items
}

// ValidateWorkItem checks if a work item has valid fields
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
