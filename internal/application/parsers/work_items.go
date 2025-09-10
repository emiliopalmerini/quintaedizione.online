package parsers

// CreateDefaultWorkItems creates the default Italian work items configuration for the new strategy pattern
func CreateDefaultWorkItems() []WorkItem {
	return []WorkItem{
		{
			Filename:   "ita/lists/armature.md",
			Collection: "armature",
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

