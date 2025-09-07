package mongodb

// GetUniqueFieldsForCollection returns unique fields for MongoDB upsert operations
func GetUniqueFieldsForCollection(collection string) []string {
	switch collection {
	case "documenti":
		return []string{"slug"}
	case "incantesimi":
		return []string{"nome", "slug"}
	case "mostri":
		return []string{"nome", "slug"}
	case "classi":
		return []string{"nome", "slug"}
	case "backgrounds":
		return []string{"nome", "slug"}
	case "armi":
		return []string{"nome", "slug"}
	case "armature":
		return []string{"nome", "slug"}
	case "strumenti":
		return []string{"nome", "slug"}
	case "servizi":
		return []string{"nome", "slug"}
	case "equipaggiamento":
		return []string{"nome", "slug"}
	case "oggetti_magici":
		return []string{"nome", "slug"}
	case "talenti":
		return []string{"nome", "slug"}
	case "animali":
		return []string{"nome", "slug"}
	default:
		return []string{"slug"}
	}
}
