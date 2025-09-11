package mongodb

// GetUniqueFieldsForCollection returns unique fields for MongoDB upsert operations
// All entity data is now nested under "value" object, so we use nested field paths
func GetUniqueFieldsForCollection(collection string) []string {
	switch collection {
	case "documenti":
		return []string{"value.slug"}
	case "incantesimi":
		return []string{"value.nome", "value.slug"}
	case "mostri":
		return []string{"value.nome", "value.slug"}
	case "classi":
		return []string{"value.nome", "value.slug"}
	case "backgrounds":
		return []string{"value.nome", "value.slug"}
	case "armi":
		return []string{"value.nome", "value.slug"}
	case "armature":
		return []string{"value.nome", "value.slug"}
	case "strumenti":
		return []string{"value.nome", "value.slug"}
	case "servizi":
		return []string{"value.nome", "value.slug"}
	case "equipaggiamento":
		return []string{"value.nome", "value.slug"}
	case "oggetti_magici":
		return []string{"value.nome", "value.slug"}
	case "talenti":
		return []string{"value.nome", "value.slug"}
	case "animali":
		return []string{"value.nome", "value.slug"}
	default:
		return []string{"value.slug"}
	}
}
