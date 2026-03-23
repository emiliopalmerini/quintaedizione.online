package mappers

import (
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/models"
)

// BuildStatBlockData populates the stat-block view model fields on ItemPageData
// based on the "_stat_block" type marker in the document's Fields.
func BuildStatBlockData(doc *domain.Document, data *models.ItemPageData) {
	statBlockType, _ := doc.Fields["_stat_block"].(string)

	switch statBlockType {
	case "spell":
		data.Spell = buildSpellStatBlock(doc)
	case "monster":
		data.Monster = buildMonsterStatBlock(doc)
	case "class":
		data.Class = buildClassStatBlock(doc)
	case "species":
		data.Species = buildSpeciesStatBlock(doc)
	default:
		// No stat-block template — BodyHTML is used
		data.BodyHTML = doc.Content.String()
	}
}

func buildSpellStatBlock(doc *domain.Document) *models.SpellStatBlock {
	level, _ := doc.Fields["level"].(int)
	ritual, _ := doc.Fields["ritual"].(bool)

	return &models.SpellStatBlock{
		Level:              level,
		School:             getStr(doc, "school"),
		CastingTime:        getStr(doc, "casting_time"),
		Range:              getStr(doc, "range"),
		Components:         getStr(doc, "components"),
		Duration:           getStr(doc, "duration"),
		Ritual:             ritual,
		DescriptionHTML:    getStr(doc, "description_html"),
		AtHigherLevelsHTML: getStr(doc, "at_higher_levels_html"),
		Classes:            getStr(doc, "classes"),
	}
}

func buildMonsterStatBlock(doc *domain.Document) *models.MonsterStatBlock {
	m := &models.MonsterStatBlock{
		Subtitle:                getStr(doc, "subtitle"),
		AC:                      getStr(doc, "ac"),
		Initiative:              getStr(doc, "initiative"),
		HP:                      getStr(doc, "hp"),
		Speed:                   getStr(doc, "speed"),
		Skills:                  getStr(doc, "skills"),
		ResistancesHTML:         getStr(doc, "resistances_html"),
		DamageImmunitiesHTML:    getStr(doc, "damage_immunities_html"),
		ConditionImmunitiesHTML: getStr(doc, "condition_immunities_html"),
		Senses:                  getStr(doc, "senses"),
		Languages:               getStr(doc, "languages"),
		CR:                      getStr(doc, "cr"),
		Equipment:               getStr(doc, "equipment"),
	}

	// Ability scores
	scores, _ := doc.Fields["ability_scores"].(map[string]int)
	mods, _ := doc.Fields["ability_mods"].(map[string]int)
	saves, _ := doc.Fields["saving_throws"].(map[string]string)

	if len(scores) > 0 {
		abilityOrder := []struct{ key, label string }{
			{"strength", "FOR"},
			{"dexterity", "DES"},
			{"constitution", "COS"},
			{"intelligence", "INT"},
			{"wisdom", "SAG"},
			{"charisma", "CAR"},
		}
		for _, a := range abilityOrder {
			m.AbilityScores = append(m.AbilityScores, models.AbilityScore{
				Label: a.label,
				Score: scores[a.key],
				Mod:   mods[a.key],
				Save:  saves[a.key],
			})
		}
	}

	// Features
	m.Traits = extractFeatures(doc, "traits")

	// Feature sections
	sections := []struct {
		field   string
		heading string
	}{
		{"actions", "Azioni"},
		{"bonus_actions", "Azioni Bonus"},
		{"reactions", "Reazioni"},
		{"legendary_actions", "Azioni Leggendarie"},
	}
	for _, s := range sections {
		features := extractFeatures(doc, s.field)
		if len(features) > 0 {
			m.Sections = append(m.Sections, models.FeatureSection{
				Heading:  s.heading,
				Features: features,
			})
		}
	}

	return m
}

func buildClassStatBlock(doc *domain.Document) *models.ClassStatBlock {
	c := &models.ClassStatBlock{
		DescriptionHTML:   getStr(doc, "description_html"),
		ProficienciesHTML: getStr(doc, "proficiencies_html"),
	}

	// Features
	if features, ok := doc.Fields["features"].([]map[string]any); ok {
		for _, f := range features {
			level, _ := f["level"].(int)
			c.Features = append(c.Features, models.ClassFeature{
				Name:            strFromMap(f, "name"),
				Level:           level,
				DescriptionHTML: strFromMap(f, "description_html"),
			})
		}
	}

	// Subclasses
	if subclasses, ok := doc.Fields["subclasses"].([]map[string]any); ok {
		for _, sc := range subclasses {
			sub := models.Subclass{
				Name:            strFromMap(sc, "name"),
				DescriptionHTML: strFromMap(sc, "description_html"),
			}
			if scFeatures, ok := sc["features"].([]map[string]any); ok {
				for _, f := range scFeatures {
					level, _ := f["level"].(int)
					sub.Features = append(sub.Features, models.ClassFeature{
						Name:            strFromMap(f, "name"),
						Level:           level,
						DescriptionHTML: strFromMap(f, "description_html"),
					})
				}
			}
			c.Subclasses = append(c.Subclasses, sub)
		}
	}

	return c
}

func buildSpeciesStatBlock(doc *domain.Document) *models.SpeciesStatBlock {
	return &models.SpeciesStatBlock{
		Subtitle:        getStr(doc, "subtitle"),
		DescriptionHTML: getStr(doc, "description_html"),
	}
}

// Helper functions

func getStr(doc *domain.Document, key string) string {
	s, _ := doc.Fields[key].(string)
	return s
}

func strFromMap(m map[string]any, key string) string {
	s, _ := m[key].(string)
	return s
}

func extractFeatures(doc *domain.Document, field string) []models.Feature {
	raw, ok := doc.Fields[field].([]map[string]any)
	if !ok {
		return nil
	}
	features := make([]models.Feature, 0, len(raw))
	for _, f := range raw {
		features = append(features, models.Feature{
			Name:            strFromMap(f, "name"),
			DescriptionHTML: strFromMap(f, "description_html"),
		})
	}
	return features
}
