package parsers

import ()

// ParseDocument wraps DocumentsStrategy for document parsing (backward compatibility)
func ParseDocument(filename string) func([]string) ([]map[string]any, error) {
	return func(lines []string) ([]map[string]any, error) {
		strategy := NewDocumentsStrategy()
		entities, err := strategy.Parse(lines, &ParsingContext{})
		if err != nil {
			return nil, err
		}
		
		var result []map[string]any
		for _, entity := range entities {
			entityMap := map[string]any{
				"entity_type": entity.EntityType(),
				"entity":      entity,
				"filename":    filename, // preserve filename info
			}
			result = append(result, entityMap)
		}
		return result, nil
	}
}

// Temporary wrapper functions to maintain compatibility with old WorkItem system
// TODO: Migrate work.go to use strategy pattern directly

// ParseClasses wraps ClassesStrategy for backward compatibility
func ParseClasses(lines []string) ([]map[string]any, error) {
	strategy := NewClassesStrategy()
	entities, err := strategy.Parse(lines, &ParsingContext{})
	if err != nil {
		return nil, err
	}
	
	// Convert ParsedEntity back to map[string]any for compatibility
	var result []map[string]any
	for _, entity := range entities {
		// This is a temporary solution - ideally work.go should use strategies directly
		entityMap := map[string]any{
			"entity_type": entity.EntityType(),
			"entity":      entity,
		}
		result = append(result, entityMap)
	}
	return result, nil
}

// ParseBackgrounds wraps BackgroundsStrategy for backward compatibility  
func ParseBackgrounds(lines []string) ([]map[string]any, error) {
	strategy := NewBackgroundsStrategy()
	entities, err := strategy.Parse(lines, &ParsingContext{})
	if err != nil {
		return nil, err
	}
	
	var result []map[string]any
	for _, entity := range entities {
		entityMap := map[string]any{
			"entity_type": entity.EntityType(),
			"entity":      entity,
		}
		result = append(result, entityMap)
	}
	return result, nil
}

// ParseMonstersMonster wraps MonstersStrategy for backward compatibility
func ParseMonstersMonster(lines []string) ([]map[string]any, error) {
	strategy := NewMonstersStrategy()
	entities, err := strategy.Parse(lines, &ParsingContext{})
	if err != nil {
		return nil, err
	}
	
	var result []map[string]any
	for _, entity := range entities {
		entityMap := map[string]any{
			"entity_type": entity.EntityType(),
			"entity":      entity,
		}
		result = append(result, entityMap)
	}
	return result, nil
}

// ParseWeapons wraps WeaponsStrategy for backward compatibility
func ParseWeapons(lines []string) ([]map[string]any, error) {
	strategy := NewWeaponsStrategy()
	entities, err := strategy.Parse(lines, &ParsingContext{})
	if err != nil {
		return nil, err
	}
	
	var result []map[string]any
	for _, entity := range entities {
		entityMap := map[string]any{
			"entity_type": entity.EntityType(),
			"entity":      entity,
		}
		result = append(result, entityMap)
	}
	return result, nil
}

// ParseArmor wraps ArmorStrategy for backward compatibility
func ParseArmor(lines []string) ([]map[string]any, error) {
	strategy := NewArmorStrategy()
	entities, err := strategy.Parse(lines, &ParsingContext{})
	if err != nil {
		return nil, err
	}
	
	var result []map[string]any
	for _, entity := range entities {
		entityMap := map[string]any{
			"entity_type": entity.EntityType(),
			"entity":      entity,
		}
		result = append(result, entityMap)
	}
	return result, nil
}

// ParseEquipment wraps EquipmentStrategy for backward compatibility
func ParseEquipment(lines []string) ([]map[string]any, error) {
	strategy := NewEquipmentStrategy()
	entities, err := strategy.Parse(lines, &ParsingContext{})
	if err != nil {
		return nil, err
	}
	
	var result []map[string]any
	for _, entity := range entities {
		entityMap := map[string]any{
			"entity_type": entity.EntityType(),
			"entity":      entity,
		}
		result = append(result, entityMap)
	}
	return result, nil
}

// ParseMagicItems wraps MagicItemsStrategy for backward compatibility
func ParseMagicItems(lines []string) ([]map[string]any, error) {
	strategy := NewMagicItemsStrategy()
	entities, err := strategy.Parse(lines, &ParsingContext{})
	if err != nil {
		return nil, err
	}
	
	var result []map[string]any
	for _, entity := range entities {
		entityMap := map[string]any{
			"entity_type": entity.EntityType(),
			"entity":      entity,
		}
		result = append(result, entityMap)
	}
	return result, nil
}

// ParseFeats wraps FeatsStrategy for backward compatibility
func ParseFeats(lines []string) ([]map[string]any, error) {
	strategy := NewFeatsStrategy()
	entities, err := strategy.Parse(lines, &ParsingContext{})
	if err != nil {
		return nil, err
	}
	
	var result []map[string]any
	for _, entity := range entities {
		entityMap := map[string]any{
			"entity_type": entity.EntityType(),
			"entity":      entity,
		}
		result = append(result, entityMap)
	}
	return result, nil
}

// ParseMonstersAnimal wraps AnimalsStrategy for backward compatibility
func ParseMonstersAnimal(lines []string) ([]map[string]any, error) {
	strategy := NewAnimalsStrategy()
	entities, err := strategy.Parse(lines, &ParsingContext{})
	if err != nil {
		return nil, err
	}
	
	var result []map[string]any
	for _, entity := range entities {
		entityMap := map[string]any{
			"entity_type": entity.EntityType(),
			"entity":      entity,
		}
		result = append(result, entityMap)
	}
	return result, nil
}

// ParseTools wraps EquipmentStrategy for tools parsing (backward compatibility)
func ParseTools(lines []string) ([]map[string]any, error) {
	// TODO: Create dedicated ToolsStrategy
	strategy := NewEquipmentStrategy()
	entities, err := strategy.Parse(lines, &ParsingContext{})
	if err != nil {
		return nil, err
	}
	
	var result []map[string]any
	for _, entity := range entities {
		entityMap := map[string]any{
			"entity_type": "strumento",
			"entity":      entity,
		}
		result = append(result, entityMap)
	}
	return result, nil
}

// ParseServices wraps EquipmentStrategy for services parsing (backward compatibility)
func ParseServices(lines []string) ([]map[string]any, error) {
	// TODO: Create dedicated ServicesStrategy
	strategy := NewEquipmentStrategy()
	entities, err := strategy.Parse(lines, &ParsingContext{})
	if err != nil {
		return nil, err
	}
	
	var result []map[string]any
	for _, entity := range entities {
		entityMap := map[string]any{
			"entity_type": "servizio",
			"entity":      entity,
		}
		result = append(result, entityMap)
	}
	return result, nil
}

// ParseGear wraps EquipmentStrategy for gear parsing (backward compatibility)
func ParseGear(lines []string) ([]map[string]any, error) {
	return ParseEquipment(lines)
}

// ParseAnimali wraps AnimalsStrategy for backward compatibility
func ParseAnimali(lines []string) ([]map[string]any, error) {
	return ParseMonstersAnimal(lines)
}

// ParseSpells wraps SpellsStrategy for backward compatibility
func ParseSpells(lines []string) ([]map[string]any, error) {
	strategy := NewSpellsStrategy()
	entities, err := strategy.Parse(lines, &ParsingContext{})
	if err != nil {
		return nil, err
	}
	
	var result []map[string]any
	for _, entity := range entities {
		entityMap := map[string]any{
			"entity_type": entity.EntityType(),
			"entity":      entity,
		}
		result = append(result, entityMap)
	}
	return result, nil
}

// CreateDefaultWork creates the default Italian work items configuration
func CreateDefaultWork() []WorkItem {
	return []WorkItem{
		// Document pages (Italian)
		{
			Filename:   "ita/01_informazioni_legali.md",
			Collection: "documenti",
			Parser:     ParseDocument("01_informazioni_legali.md"),
		},
		{
			Filename:   "ita/02_giocare_il_gioco.md",
			Collection: "documenti",
			Parser:     ParseDocument("02_giocare_il_gioco.md"),
		},
		{
			Filename:   "ita/03_creazione_personaggio.md",
			Collection: "documenti",
			Parser:     ParseDocument("03_creazione_personaggio.md"),
		},
		{
			Filename:   "ita/04_classi.md",
			Collection: "documenti",
			Parser:     ParseDocument("04_classi.md"),
		},
		{
			Filename:   "ita/05_origini_personaggio.md",
			Collection: "documenti",
			Parser:     ParseDocument("05_origini_personaggio.md"),
		},
		{
			Filename:   "ita/06_talenti.md",
			Collection: "documenti",
			Parser:     ParseDocument("06_talenti.md"),
		},
		{
			Filename:   "ita/07_equipaggiamento.md",
			Collection: "documenti",
			Parser:     ParseDocument("07_equipaggiamento.md"),
		},
		{
			Filename:   "ita/08_equipaggiamento_items.md",
			Collection: "documenti",
			Parser:     ParseDocument("08_equipaggiamento_items.md"),
		},
		{
			Filename:   "ita/09_armi_items.md",
			Collection: "documenti",
			Parser:     ParseDocument("09_armi_items.md"),
		},
		{
			Filename:   "ita/10_oggetti_magici_items.md",
			Collection: "documenti",
			Parser:     ParseDocument("10_oggetti_magici_items.md"),
		},
		{
			Filename:   "ita/11_armatura_items.md",
			Collection: "documenti",
			Parser:     ParseDocument("11_armatura_items.md"),
		},
		{
			Filename:   "ita/12_strumenti_items.md",
			Collection: "documenti",
			Parser:     ParseDocument("12_strumenti_items.md"),
		},
		{
			Filename:   "ita/13_servizi_items.md",
			Collection: "documenti",
			Parser:     ParseDocument("13_servizi_items.md"),
		},
		{
			Filename:   "ita/14_cavalcature_veicoli_items.md",
			Collection: "documenti",
			Parser:     ParseDocument("14_cavalcature_veicoli_items.md"),
		},
		{
			Filename:   "ita/15_incantesimi.md",
			Collection: "documenti",
			Parser:     ParseDocument("15_incantesimi.md"),
		},
		{
			Filename:   "ita/16_incantesimi_items.md",
			Collection: "documenti",
			Parser:     ParseDocument("16_incantesimi_items.md"),
		},
		{
			Filename:   "ita/17_glossario_regole.md",
			Collection: "documenti",
			Parser:     ParseDocument("17_glossario_regole.md"),
		},
		{
			Filename:   "ita/18_gameplay_toolbox.md",
			Collection: "documenti",
			Parser:     ParseDocument("18_gameplay_toolbox.md"),
		},
		{
			Filename:   "ita/19_mostri.md",
			Collection: "documenti",
			Parser:     ParseDocument("19_mostri.md"),
		},
		{
			Filename:   "ita/20_mostri_items.md",
			Collection: "documenti",
			Parser:     ParseDocument("20_mostri_items.md"),
		},
		{
			Filename:   "ita/21_animali.md",
			Collection: "documenti",
			Parser:     ParseDocument("21_animali.md"),
		},

		// Structured data (Italian only)
		{
			Filename:   "ita/04_classi.md",
			Collection: "classi",
			Parser:     ParseClasses,
		},
		{
			Filename:   "ita/05_origini_personaggio.md",
			Collection: "backgrounds",
			Parser:     ParseBackgrounds,
		},
		{
			Filename:   "ita/09_armi_items.md",
			Collection: "armi",
			Parser:     ParseWeapons,
		},
		{
			Filename:   "ita/11_armatura_items.md",
			Collection: "armature",
			Parser:     ParseArmor,
		},
		{
			Filename:   "ita/12_strumenti_items.md",
			Collection: "strumenti",
			Parser:     ParseTools,
		},
		{
			Filename:   "ita/13_servizi_items.md",
			Collection: "servizi",
			Parser:     ParseServices,
		},
		{
			Filename:   "ita/08_equipaggiamento_items.md",
			Collection: "equipaggiamento",
			Parser:     ParseGear,
		},
		{
			Filename:   "ita/10_oggetti_magici_items.md",
			Collection: "oggetti_magici",
			Parser:     ParseMagicItems,
		},
		{
			Filename:   "ita/16_incantesimi_items.md",
			Collection: "incantesimi",
			Parser:     ParseSpells,
		},
		{
			Filename:   "ita/06_talenti.md",
			Collection: "talenti",
			Parser:     ParseFeats,
		},
		{
			Filename:   "ita/20_mostri_items.md",
			Collection: "mostri",
			Parser:     ParseMonstersMonster,
		},
		{
			Filename:   "ita/21_animali.md",
			Collection: "animali",
			Parser:     ParseAnimali,
		},
	}
}

// Note: All parsers are implemented in their respective files:
// - ParseDocument in documents.go
// - ParseClasses in classes.go
// - ParseBackgrounds in backgrounds.go
// - ParseWeapons in weapons.go
// - ParseArmor in armor.go
// - ParseTools, ParseServices, ParseGear in equipment.go
// - ParseMagicItems in magic_items.go
// - ParseFeats in feats.go
// - ParseMonstersMonster in monsters.go
// - ParseAnimali in animali.go
