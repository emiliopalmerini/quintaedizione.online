package datastore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/application/parsers"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain"
)

// Source is an alias for the domain Source type.
type Source = domain.Source

// Loader reads embedded JSON files and converts them into the document format
// expected by the in-memory store (map[string]any with _id, title, content, raw_content, filters).
type Loader struct {
	fsys     fs.FS
	renderer *parsers.MarkdownRenderer
	prefix   string // source directory prefix (e.g. "srd-5.5e")
	source   *Source
}

// NewLoader creates a Loader that reads from the given filesystem and renders
// markdown descriptions to HTML using the provided renderer.
func NewLoader(fsys fs.FS, renderer *parsers.MarkdownRenderer) *Loader {
	return &Loader{fsys: fsys, renderer: renderer}
}

// LoadAll discovers all source directories, loads their data, and returns
// documents grouped by collection name. Also returns the list of loaded sources.
func (l *Loader) LoadAll() (map[string][]map[string]any, []Source, error) {
	result := make(map[string][]map[string]any)
	var sources []Source

	entries, err := fs.ReadDir(l.fsys, ".")
	if err != nil {
		return nil, nil, fmt.Errorf("read data directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		src, err := l.loadSource(entry.Name())
		if err != nil {
			fmt.Printf("Warning: skipping %s: %v\n", entry.Name(), err)
			continue
		}

		l.prefix = entry.Name()
		l.source = &src
		sources = append(sources, src)

		if err := l.loadSourceData(result); err != nil {
			return nil, nil, fmt.Errorf("source %s: %w", src.ID, err)
		}
	}

	if len(sources) == 0 {
		return nil, nil, fmt.Errorf("no sources found")
	}

	return result, sources, nil
}

func (l *Loader) loadSource(dir string) (Source, error) {
	data, err := fs.ReadFile(l.fsys, dir+"/source.json")
	if err != nil {
		return Source{}, fmt.Errorf("read source.json: %w", err)
	}
	var src Source
	if err := json.Unmarshal(data, &src); err != nil {
		return Source{}, fmt.Errorf("parse source.json: %w", err)
	}
	if src.ID == "" {
		return Source{}, fmt.Errorf("source.json missing id")
	}
	return src, nil
}

func (l *Loader) loadSourceData(result map[string][]map[string]any) error {
	loaders := []struct {
		name string
		fn   func(map[string][]map[string]any) error
	}{
		{"spells", l.loadSpells},
		{"monsters", l.loadMonsters},
		{"classes", l.loadClasses},
		{"backgrounds", l.loadBackgrounds},
		{"equipment", l.loadEquipment},
		{"magic_items", l.loadMagicItems},
		{"feats", l.loadFeats},
		{"rules", l.loadRules},
		{"glossary", l.loadGlossary},
		{"species", l.loadSpecies},
	}

	for _, loader := range loaders {
		if err := loader.fn(result); err != nil {
			// Skip collections whose JSON files don't exist in this source
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return fmt.Errorf("%s: %w", loader.name, err)
		}
	}
	return nil
}

func (l *Loader) readJSON(filename string, v any) error {
	path := filename
	if l.prefix != "" {
		path = l.prefix + "/" + filename
	}
	data, err := fs.ReadFile(l.fsys, path)
	if err != nil {
		// Wrap with %w to preserve fs.ErrNotExist for callers
		return fmt.Errorf("read %s: %w", path, err)
	}
	return json.Unmarshal(data, v)
}

// tagDoc injects source metadata into a document.
// The _id stays clean (no prefix). The store uses (source, id) as a composite key.
func (l *Loader) tagDoc(doc map[string]any) {
	if l.source != nil {
		doc["_source"] = l.source.ID
		doc["_source_short"] = l.source.ShortName
	}
}

// --- Spells ---

type jsonSpell struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Level          int      `json:"level"`
	School         string   `json:"school"`
	Classes        []string `json:"classes"`
	CastingTime    string   `json:"casting_time"`
	Range          string   `json:"range"`
	Components     string   `json:"components"`
	Duration       string   `json:"duration"`
	Description    string   `json:"description"`
	AtHigherLevels string   `json:"at_higher_levels"`
	Ritual         bool     `json:"ritual"`
}

func (l *Loader) loadSpells(result map[string][]map[string]any) error {
	var spells []jsonSpell
	if err := l.readJSON("spells.json", &spells); err != nil {
		return err
	}

	docs := make([]map[string]any, 0, len(spells))
	for _, s := range spells {
		doc := map[string]any{
			"_id":         s.ID,
			"title":       s.Name,
			"content":     l.buildSpellHTML(s),
			"raw_content": l.buildSpellMarkdown(s),
			"scuola":      s.School,
			"livello":     s.Level,
			"classe":      strings.Join(s.Classes, ", "),
		}
		l.tagDoc(doc)
		docs = append(docs, doc)
	}

	result["incantesimi"] = append(result["incantesimi"], docs...)
	return nil
}

// --- Monsters ---

type jsonFeature struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type jsonMonster struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	Type                string            `json:"type"`
	Size                string            `json:"size"`
	Alignment           string            `json:"alignment"`
	CR                  string            `json:"cr"`
	CRDetail            string            `json:"cr_detail"`
	Source              string            `json:"source"`
	AC                  string            `json:"ac"`
	Initiative          string            `json:"initiative"`
	HP                  string            `json:"hp"`
	Speed               string            `json:"speed"`
	Skills              string            `json:"skills"`
	Resistances         string            `json:"resistances"`
	DamageImmunities    string            `json:"damage_immunities"`
	ConditionImmunities string            `json:"condition_immunities"`
	Senses              string            `json:"senses"`
	Languages           string            `json:"languages"`
	Equipment           string            `json:"equipment"`
	Traits              []jsonFeature     `json:"traits"`
	Actions             []jsonFeature     `json:"actions"`
	BonusActions        []jsonFeature     `json:"bonus_actions"`
	Reactions           []jsonFeature     `json:"reactions"`
	LegendaryActions    []jsonFeature     `json:"legendary_actions"`
	Group               string            `json:"group"`
	AbilityScores       map[string]int    `json:"ability_scores"`
	AbilityMods         map[string]int    `json:"ability_mods"`
	SavingThrows        map[string]string `json:"saving_throws"`
}

func (l *Loader) loadMonsters(result map[string][]map[string]any) error {
	var monsters []jsonMonster
	if err := l.readJSON("monsters.json", &monsters); err != nil {
		return err
	}

	docs := make([]map[string]any, 0, len(monsters))
	for _, m := range monsters {
		doc := map[string]any{
			"_id":          m.ID,
			"title":        m.Name,
			"content":      l.buildMonsterHTML(m),
			"raw_content":  l.buildMonsterMarkdown(m),
			"tipo":         m.Type,
			"taglia":       m.Size,
			"allineamento": m.Alignment,
			"grado_sfida":  m.CR,
		}
		l.tagDoc(doc)
		docs = append(docs, doc)
	}

	result["mostri"] = append(result["mostri"], docs...)
	return nil
}

// --- Classes ---

type jsonClassFeature struct {
	Name        string `json:"name"`
	Level       int    `json:"level"`
	Description string `json:"description"`
}

type jsonSubclass struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Features    []jsonClassFeature `json:"features"`
}

type jsonClass struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	HitDie        string             `json:"hit_die"`
	Proficiencies string             `json:"proficiencies"`
	Description   string             `json:"description"`
	Features      []jsonClassFeature `json:"features"`
	Subclasses    []jsonSubclass     `json:"subclasses"`
	SpellList     []string           `json:"spell_list"`
}

func (l *Loader) loadClasses(result map[string][]map[string]any) error {
	var classes []jsonClass
	if err := l.readJSON("classes.json", &classes); err != nil {
		return err
	}

	docs := make([]map[string]any, 0, len(classes))
	for _, c := range classes {
		doc := map[string]any{
			"_id":         c.ID,
			"title":       c.Name,
			"content":     l.buildClassHTML(c),
			"raw_content": l.buildClassMarkdown(c),
		}
		l.tagDoc(doc)
		docs = append(docs, doc)
	}

	result["classi"] = append(result["classi"], docs...)
	return nil
}

// --- Backgrounds ---

type jsonBackground struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	AbilityScores      string `json:"ability_scores"`
	Feat               string `json:"feat"`
	SkillProficiencies string `json:"skill_proficiencies"`
	ToolProficiency    string `json:"tool_proficiency"`
	Equipment          string `json:"equipment"`
	Description        string `json:"description"`
}

func (l *Loader) loadBackgrounds(result map[string][]map[string]any) error {
	var backgrounds []jsonBackground
	if err := l.readJSON("backgrounds.json", &backgrounds); err != nil {
		return err
	}

	docs := make([]map[string]any, 0, len(backgrounds))
	for _, bg := range backgrounds {
		raw := l.buildBackgroundMarkdown(bg)
		doc := map[string]any{
			"_id":         bg.ID,
			"title":       bg.Name,
			"content":     l.renderer.Render(raw),
			"raw_content": raw,
		}
		l.tagDoc(doc)
		docs = append(docs, doc)
	}

	result["backgrounds"] = append(result["backgrounds"], docs...)
	return nil
}

func (l *Loader) buildBackgroundMarkdown(bg jsonBackground) string {
	var b strings.Builder

	if bg.AbilityScores != "" {
		fmt.Fprintf(&b, "**Punteggi di Caratteristica:** %s\n\n", bg.AbilityScores)
	}
	if bg.Feat != "" {
		fmt.Fprintf(&b, "**Talento:** %s\n\n", bg.Feat)
	}
	if bg.SkillProficiencies != "" {
		fmt.Fprintf(&b, "**Competenze nelle Abilità:** %s\n\n", bg.SkillProficiencies)
	}
	if bg.ToolProficiency != "" {
		fmt.Fprintf(&b, "**Competenza negli Strumenti:** %s\n\n", bg.ToolProficiency)
	}
	if bg.Equipment != "" {
		fmt.Fprintf(&b, "**Equipaggiamento:** %s\n\n", bg.Equipment)
	}
	if bg.Description != "" {
		b.WriteString(bg.Description)
	}

	return strings.TrimSpace(b.String())
}

// --- Equipment ---

type jsonEquipment struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Category    string            `json:"category"`
	Subcategory string            `json:"subcategory"`
	Properties  map[string]string `json:"properties"`
	Description string            `json:"description"`
}

// equipmentCollectionName maps a JSON equipment item to its Go collection.
func equipmentCollectionName(category string) string {
	if category == "services" {
		return "servizi"
	}
	return "equipaggiamenti"
}

func (l *Loader) loadEquipment(result map[string][]map[string]any) error {
	var items []jsonEquipment
	if err := l.readJSON("equipment.json", &items); err != nil {
		return err
	}

	for _, item := range items {
		collection := equipmentCollectionName(item.Category)

		doc := map[string]any{
			"_id":         item.ID,
			"title":       item.Name,
			"content":     l.renderer.Render(item.Description),
			"raw_content": item.Description,
			"categoria":   item.Subcategory,
		}

		// Copy properties as flat fields for display strategies
		for k, v := range item.Properties {
			doc[k] = v
		}

		l.tagDoc(doc)
		result[collection] = append(result[collection], doc)
	}

	return nil
}

// --- Magic Items ---

type jsonMagicItem struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Type              string `json:"type"`
	Rarity            string `json:"rarity"`
	Attunement        bool   `json:"attunement"`
	AttunementDetails string `json:"attunement_details"`
	Description       string `json:"description"`
}

func (l *Loader) loadMagicItems(result map[string][]map[string]any) error {
	var items []jsonMagicItem
	if err := l.readJSON("magic_items.json", &items); err != nil {
		return err
	}

	docs := make([]map[string]any, 0, len(items))
	for _, item := range items {
		doc := map[string]any{
			"_id":         item.ID,
			"title":       item.Name,
			"content":     l.renderer.Render(item.Description),
			"raw_content": item.Description,
			"rarita":      item.Rarity,
			"tipo":        item.Type,
		}
		l.tagDoc(doc)
		docs = append(docs, doc)
	}

	result["oggetti_magici"] = append(result["oggetti_magici"], docs...)
	return nil
}

// --- Feats ---

type jsonFeat struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Category     string `json:"category"`
	Prerequisite string `json:"prerequisite"`
	Repeatable   bool   `json:"repeatable"`
	Benefit      string `json:"benefit"`
}

func (l *Loader) loadFeats(result map[string][]map[string]any) error {
	var feats []jsonFeat
	if err := l.readJSON("feats.json", &feats); err != nil {
		return err
	}

	docs := make([]map[string]any, 0, len(feats))
	for _, f := range feats {
		doc := map[string]any{
			"_id":         f.ID,
			"title":       f.Name,
			"content":     l.renderer.Render(f.Benefit),
			"raw_content": f.Benefit,
			"categoria":   f.Category,
		}
		l.tagDoc(doc)
		docs = append(docs, doc)
	}

	result["talenti"] = append(result["talenti"], docs...)
	return nil
}

// --- Rules ---

type jsonRule struct {
	ID       string     `json:"id"`
	Title    string     `json:"title"`
	Content  string     `json:"content"`
	Children []jsonRule `json:"children"`
}

func (l *Loader) loadRules(result map[string][]map[string]any) error {
	// Discover all rules_*.json files in the source directory
	dir := l.prefix
	if dir == "" {
		dir = "."
	}
	entries, err := fs.ReadDir(l.fsys, dir)
	if err != nil {
		return fmt.Errorf("read rules dir: %w", err)
	}
	var ruleFiles []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "rules_") && strings.HasSuffix(name, ".json") {
			ruleFiles = append(ruleFiles, name)
		}
	}
	if len(ruleFiles) == 0 {
		return nil
	}

	var docs []map[string]any
	for _, filename := range ruleFiles {
		var rules []jsonRule
		if err := l.readJSON(filename, &rules); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return fmt.Errorf("%s: %w", filename, err)
		}

		for _, r := range rules {
			// Flatten: use children as the actual content items since each rule file
			// is typically one top-level entry with children
			if len(r.Children) > 0 {
				for _, child := range r.Children {
					doc := l.ruleToDoc(child)
					l.tagDoc(doc)
					docs = append(docs, doc)
				}
			} else {
				doc := l.ruleToDoc(r)
				l.tagDoc(doc)
				docs = append(docs, doc)
			}
		}
	}

	result["regole"] = append(result["regole"], docs...)
	return nil
}

func (l *Loader) ruleToDoc(r jsonRule) map[string]any {
	// Build full content including children
	raw := r.Content
	for _, child := range r.Children {
		raw += "\n\n### " + child.Title + "\n\n" + child.Content
	}

	return map[string]any{
		"_id":         r.ID,
		"title":       r.Title,
		"content":     l.renderer.Render(raw),
		"raw_content": raw,
	}
}

// --- Glossary ---

type jsonGlossaryEntry struct {
	ID         string   `json:"id"`
	Term       string   `json:"term"`
	Category   string   `json:"category"`
	Definition string   `json:"definition"`
	SeeAlso    []string `json:"see_also"`
}

func (l *Loader) loadGlossary(result map[string][]map[string]any) error {
	var entries []jsonGlossaryEntry
	if err := l.readJSON("glossary.json", &entries); err != nil {
		return err
	}

	docs := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		raw := e.Definition
		if len(e.SeeAlso) > 0 {
			raw += "\n\n*Vedi anche: " + strings.Join(e.SeeAlso, ", ") + "*"
		}
		doc := map[string]any{
			"_id":         e.ID,
			"title":       e.Term,
			"content":     l.renderer.Render(raw),
			"raw_content": raw,
			"categoria":   e.Category,
		}
		l.tagDoc(doc)
		docs = append(docs, doc)
	}

	result["glossario"] = append(result["glossario"], docs...)
	return nil
}

// --- Species ---

type jsonSpeciesTrait struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type jsonSpecies struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	CreatureType string             `json:"creature_type"`
	Size         string             `json:"size"`
	Speed        string             `json:"speed"`
	Traits       []jsonSpeciesTrait `json:"traits"`
	Description  string             `json:"description"`
}

func (l *Loader) loadSpecies(result map[string][]map[string]any) error {
	var species []jsonSpecies
	if err := l.readJSON("species.json", &species); err != nil {
		return err
	}

	docs := make([]map[string]any, 0, len(species))
	for _, s := range species {
		doc := map[string]any{
			"_id":           s.ID,
			"title":         s.Name,
			"content":       l.buildSpeciesHTML(s),
			"raw_content":   l.buildSpeciesMarkdown(s),
			"tipo_creatura": s.CreatureType,
			"taglia":        s.Size,
			"velocita":      s.Speed,
		}
		l.tagDoc(doc)
		docs = append(docs, doc)
	}

	result["specie"] = append(result["specie"], docs...)
	return nil
}
