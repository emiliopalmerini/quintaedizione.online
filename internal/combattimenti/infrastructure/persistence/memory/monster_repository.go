package memory

import (
	"encoding/json"
	"io/fs"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/monster"
)

var xpRegex = regexp.MustCompile(`PE\s+([\d.]+)`)

type jsonAbilityScores struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}

type jsonSavingThrows struct {
	Strength     string `json:"strength"`
	Dexterity    string `json:"dexterity"`
	Constitution string `json:"constitution"`
	Intelligence string `json:"intelligence"`
	Wisdom       string `json:"wisdom"`
	Charisma     string `json:"charisma"`
}

type jsonNamedDescription struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type jsonMonster struct {
	ID                  string                 `json:"id"`
	Name                string                 `json:"name"`
	Group               string                 `json:"group"`
	Type                string                 `json:"type"`
	Subtype             string                 `json:"subtype"`
	Size                string                 `json:"size"`
	Alignment           string                 `json:"alignment"`
	AC                  string                 `json:"ac"`
	Initiative          string                 `json:"initiative"`
	HP                  string                 `json:"hp"`
	Speed               string                 `json:"speed"`
	AbilityScores       jsonAbilityScores      `json:"ability_scores"`
	AbilityMods         jsonAbilityScores      `json:"ability_mods"`
	SavingThrows        jsonSavingThrows       `json:"saving_throws"`
	Skills              string                 `json:"skills"`
	Resistances         string                 `json:"resistances"`
	DamageImmunities    string                 `json:"damage_immunities"`
	ConditionImmunities string                 `json:"condition_immunities"`
	Senses              string                 `json:"senses"`
	Languages           string                 `json:"languages"`
	CR                  string                 `json:"cr"`
	CRDetail            string                 `json:"cr_detail"`
	Equipment           string                 `json:"equipment"`
	Traits              []jsonNamedDescription `json:"traits"`
	Actions             []jsonNamedDescription `json:"actions"`
	BonusActions        []jsonNamedDescription `json:"bonus_actions"`
	Reactions           []jsonNamedDescription `json:"reactions"`
	LegendaryActions    []jsonNamedDescription `json:"legendary_actions"`
}

// MonsterRepository provides in-memory access to monster data.
type MonsterRepository struct {
	monsters       []monster.Monster
	availableTypes []string
	availableSizes []string
	availableCRs   []string
}

// NewMonsterRepository loads monsters from the provided filesystem (typically the shared SRD embed.FS).
// filename is the path within the FS (e.g. "srd-5.5e/monsters.json").
// sourceID is extracted from the directory name (e.g. "srd-5.5e").
func NewMonsterRepository(dataFS fs.FS, filename string) *MonsterRepository {
	// Extract source ID from directory prefix (e.g. "srd-5.5e/monsters.json" → "srd-5.5e")
	sourceID := ""
	if idx := strings.Index(filename, "/"); idx != -1 {
		sourceID = filename[:idx]
	}

	data, err := fs.ReadFile(dataFS, filename)
	if err != nil {
		log.Fatalf("failed to read %s: %v", filename, err)
	}

	var raw []jsonMonster
	if err := json.Unmarshal(data, &raw); err != nil {
		log.Fatalf("failed to parse %s: %v", filename, err)
	}

	monsters := make([]monster.Monster, len(raw))
	for i, m := range raw {
		monsters[i] = monster.Monster{
			ID:     m.ID,
			Source: sourceID,
			Name:   m.Name,
			Type:   m.Type,
			Size:   normalizeSize(m.Size),
			CR:     m.CR,
			XP:     parseXP(m.CRDetail),
			AC:     m.AC,
			HP:     m.HP,

			Group:               m.Group,
			Subtype:             m.Subtype,
			Alignment:           m.Alignment,
			Initiative:          m.Initiative,
			Speed:               m.Speed,
			AbilityScores:       monster.AbilityScores(m.AbilityScores),
			AbilityMods:         monster.AbilityScores(m.AbilityMods),
			SavingThrows:        monster.SavingThrows(m.SavingThrows),
			Skills:              m.Skills,
			Senses:              m.Senses,
			Languages:           m.Languages,
			Resistances:         m.Resistances,
			DamageImmunities:    m.DamageImmunities,
			ConditionImmunities: m.ConditionImmunities,
			Equipment:           m.Equipment,
			CRDetail:            m.CRDetail,
			Traits:              convertNamedDescriptions(m.Traits),
			Actions:             convertNamedDescriptions(m.Actions),
			BonusActions:        convertNamedDescriptions(m.BonusActions),
			Reactions:           convertNamedDescriptions(m.Reactions),
			LegendaryActions:    convertNamedDescriptions(m.LegendaryActions),
		}
	}

	sort.Slice(monsters, func(i, j int) bool {
		return monsters[i].XP < monsters[j].XP
	})

	repo := &MonsterRepository{monsters: monsters}
	repo.buildFacets()
	return repo
}

func convertNamedDescriptions(src []jsonNamedDescription) []monster.NamedDescription {
	if len(src) == 0 {
		return nil
	}
	result := make([]monster.NamedDescription, len(src))
	for i, nd := range src {
		result[i] = monster.NamedDescription{Name: nd.Name, Description: nd.Description}
	}
	return result
}

// parseXP extracts the XP value from a cr_detail string like "10 (PE 5.900; BC +4)".
// Italian uses dots as thousands separators, so "5.900" → 5900.
func parseXP(crDetail string) int {
	matches := xpRegex.FindStringSubmatch(crDetail)
	if len(matches) < 2 {
		return 0
	}
	numStr := strings.ReplaceAll(matches[1], ".", "")
	xp, err := strconv.Atoi(numStr)
	if err != nil {
		return 0
	}
	return xp
}

func (r *MonsterRepository) FindByMaxXP(maxXP int) []monster.Monster {
	var result []monster.Monster
	for _, m := range r.monsters {
		if m.XP <= maxXP {
			result = append(result, m)
		}
	}
	return result
}

func (r *MonsterRepository) Search(query string, maxXP int) []monster.Monster {
	query = strings.ToLower(query)
	var result []monster.Monster
	for _, m := range r.monsters {
		if m.XP > maxXP {
			continue
		}
		if query == "" || strings.Contains(strings.ToLower(m.Name), query) {
			result = append(result, m)
		}
	}
	return result
}

// normalizeSize maps Italian masculine/compound size variants to canonical feminine forms.
func normalizeSize(s string) string {
	if idx := strings.IndexAny(s, " "); idx != -1 {
		s = s[:idx]
	}
	switch s {
	case "Minuscolo":
		return "Minuscola"
	case "Piccolo":
		return "Piccola"
	case "Medio":
		return "Media"
	case "Mastodontico":
		return "Mastodontica"
	default:
		return s
	}
}

func crValue(cr string) float64 {
	switch cr {
	case "1/8":
		return 0.125
	case "1/4":
		return 0.25
	case "1/2":
		return 0.5
	}
	v, err := strconv.ParseFloat(cr, 64)
	if err != nil {
		return 0
	}
	return v
}

func (r *MonsterRepository) buildFacets() {
	typeSet := make(map[string]bool)
	sizeSet := make(map[string]bool)
	crSet := make(map[string]bool)

	for _, m := range r.monsters {
		typeSet[m.Type] = true
		sizeSet[m.Size] = true
		crSet[m.CR] = true
	}

	r.availableTypes = sortedKeys(typeSet)
	r.availableSizes = sortedKeys(sizeSet)

	crs := make([]string, 0, len(crSet))
	for cr := range crSet {
		crs = append(crs, cr)
	}
	sort.Slice(crs, func(i, j int) bool {
		return crValue(crs[i]) < crValue(crs[j])
	})
	r.availableCRs = crs
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (r *MonsterRepository) SearchWithFilters(filters monster.SearchFilters) []monster.Monster {
	query := strings.ToLower(filters.Query)
	var result []monster.Monster

	crMin := float64(-1)
	crMax := float64(1_000_000)
	if filters.CRMin != "" {
		crMin = crValue(filters.CRMin)
	}
	if filters.CRMax != "" {
		crMax = crValue(filters.CRMax)
	}

	for _, m := range r.monsters {
		if filters.MaxXP > 0 && m.XP > filters.MaxXP {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(m.Name), query) {
			continue
		}
		if filters.Type != "" && m.Type != filters.Type {
			continue
		}
		if filters.Size != "" && m.Size != filters.Size {
			continue
		}
		cr := crValue(m.CR)
		if cr < crMin || cr > crMax {
			continue
		}
		result = append(result, m)
	}
	return result
}

func (r *MonsterRepository) AvailableTypes() []string { return r.availableTypes }
func (r *MonsterRepository) AvailableSizes() []string { return r.availableSizes }
func (r *MonsterRepository) AvailableCRs() []string   { return r.availableCRs }
