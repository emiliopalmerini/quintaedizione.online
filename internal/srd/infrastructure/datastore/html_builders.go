package datastore

import (
	"fmt"
	"strings"
)

// ContentRenderer converts markdown to HTML. The Loader uses this interface
// to decouple data loading from the rendering implementation.
type ContentRenderer interface {
	// Render converts markdown to HTML (with glossary linking, etc.).
	Render(markdown string) string
	// RenderInline converts markdown to HTML and strips the wrapping <p> tags.
	RenderInline(markdown string) string
}

// renderContent renders markdown to HTML (with cross-linking via the renderer).
func (l *Loader) renderContent(md string) string {
	return l.renderer.Render(md)
}

// renderContentInline is like renderContent but strips the <p> wrapper.
func (l *Loader) renderContentInline(md string) string {
	return l.renderer.RenderInline(md)
}

// sourceShort returns the current source's short name for URL building.
func (l *Loader) sourceShort() string {
	if l.source != nil {
		return l.source.ShortName
	}
	return ""
}

// ---------------------------------------------------------------------------
// Markdown builders — produce markdown for the "raw_content" field (copy-to-clipboard)
// ---------------------------------------------------------------------------

func (l *Loader) buildSpellMarkdown(s jsonSpell) string {
	var b strings.Builder

	// Subtitle line
	if s.Level == 0 {
		fmt.Fprintf(&b, "*Trucchetto %s*\n\n", s.School)
	} else {
		fmt.Fprintf(&b, "*Livello %d %s*\n\n", s.Level, s.School)
	}

	fmt.Fprintf(&b, "**Tempo di Lancio:** %s\n\n", s.CastingTime)
	fmt.Fprintf(&b, "**Gittata:** %s\n\n", s.Range)
	fmt.Fprintf(&b, "**Componenti:** %s\n\n", s.Components)
	fmt.Fprintf(&b, "**Durata:** %s\n\n", s.Duration)

	if s.Ritual {
		b.WriteString("*Rituale*\n\n")
	}

	b.WriteString(s.Description.plainText())

	if len(s.AtHigherLevels) > 0 {
		b.WriteString("\n\n**Ai Livelli Superiori.** ")
		b.WriteString(s.AtHigherLevels.plainText())
	}

	fmt.Fprintf(&b, "\n\n*Classi: %s*", strings.Join(s.Classes, ", "))

	return b.String()
}

func (l *Loader) buildMonsterMarkdown(m jsonMonster) string {
	var b strings.Builder

	// Type line
	fmt.Fprintf(&b, "*%s %s, %s*\n\n", m.Type, m.Size, m.Alignment)

	// Core stats
	b.WriteString("---\n\n")
	fmt.Fprintf(&b, "**CA** %s", m.AC)
	if m.Initiative != "" {
		fmt.Fprintf(&b, " · **Iniziativa** %s", m.Initiative)
	}
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "**PF** %s\n\n", m.HP)
	fmt.Fprintf(&b, "**Velocità** %s\n\n", m.Speed)

	// Ability scores table
	if len(m.AbilityScores) > 0 {
		b.WriteString("---\n\n")
		abilityOrder := []struct{ key, label string }{
			{"strength", "FOR"},
			{"dexterity", "DES"},
			{"constitution", "COS"},
			{"intelligence", "INT"},
			{"wisdom", "SAG"},
			{"charisma", "CAR"},
		}
		b.WriteString("| ")
		for _, a := range abilityOrder {
			b.WriteString(a.label + " | ")
		}
		b.WriteString("\n|")
		for range abilityOrder {
			b.WriteString(":---:|")
		}
		b.WriteString("\n| ")
		for _, a := range abilityOrder {
			score := m.AbilityScores[a.key]
			mod := m.AbilityMods[a.key]
			save := m.SavingThrows[a.key]
			fmt.Fprintf(&b, "%d (%+d) TS %s | ", score, mod, save)
		}
		b.WriteString("\n\n")
	}

	// Secondary stats
	b.WriteString("---\n\n")
	if m.Skills != "" {
		fmt.Fprintf(&b, "**Abilità** %s\n\n", m.Skills)
	}
	if len(m.Resistances) > 0 {
		fmt.Fprintf(&b, "**Resistenze** %s\n\n", m.Resistances.plainText())
	}
	if len(m.DamageImmunities) > 0 {
		fmt.Fprintf(&b, "**Immunità ai Danni** %s\n\n", m.DamageImmunities.plainText())
	}
	if len(m.ConditionImmunities) > 0 {
		fmt.Fprintf(&b, "**Immunità alle Condizioni** %s\n\n", m.ConditionImmunities.plainText())
	}
	if m.Senses != "" {
		fmt.Fprintf(&b, "**Sensi** %s\n\n", m.Senses)
	}
	if m.Languages != "" {
		fmt.Fprintf(&b, "**Lingue** %s\n\n", m.Languages)
	}
	if m.CRDetail != "" {
		fmt.Fprintf(&b, "**GS** %s\n\n", m.CRDetail)
	} else {
		fmt.Fprintf(&b, "**GS** %s\n\n", m.CR)
	}
	if m.Equipment != "" {
		fmt.Fprintf(&b, "**Equipaggiamento** %s\n\n", m.Equipment)
	}

	// Traits
	if len(m.Traits) > 0 {
		b.WriteString("---\n\n")
		for _, t := range m.Traits {
			fmt.Fprintf(&b, "***%s.*** %s\n\n", t.Name, t.Description.plainText())
		}
	}

	// Actions
	if len(m.Actions) > 0 {
		b.WriteString("### Azioni\n\n")
		for _, a := range m.Actions {
			fmt.Fprintf(&b, "***%s.*** %s\n\n", a.Name, a.Description.plainText())
		}
	}

	// Bonus Actions
	if len(m.BonusActions) > 0 {
		b.WriteString("### Azioni Bonus\n\n")
		for _, a := range m.BonusActions {
			fmt.Fprintf(&b, "***%s.*** %s\n\n", a.Name, a.Description.plainText())
		}
	}

	// Reactions
	if len(m.Reactions) > 0 {
		b.WriteString("### Reazioni\n\n")
		for _, r := range m.Reactions {
			fmt.Fprintf(&b, "***%s.*** %s\n\n", r.Name, r.Description.plainText())
		}
	}

	// Legendary Actions
	if len(m.LegendaryActions) > 0 {
		b.WriteString("### Azioni Leggendarie\n\n")
		for _, la := range m.LegendaryActions {
			fmt.Fprintf(&b, "***%s.*** %s\n\n", la.Name, la.Description.plainText())
		}
	}

	return b.String()
}

func (l *Loader) buildClassMarkdown(c jsonClass) string {
	var b strings.Builder

	// Description (contains traits box + level progression table)
	if len(c.Description) > 0 {
		b.WriteString(c.Description.plainText())
		b.WriteString("\n\n")
	}

	// Proficiencies (contains hit points, competencies, equipment, and class table)
	if c.Proficiencies != "" {
		b.WriteString(c.Proficiencies)
		b.WriteString("\n\n")
	}

	// Features grouped by level
	if len(c.Features) > 0 {
		b.WriteString("---\n\n")
		b.WriteString("## Privilegi di classe\n\n")
		for _, f := range c.Features {
			fmt.Fprintf(&b, "### %s (Livello %d)\n\n", f.Name, f.Level)
			b.WriteString(f.Description.plainText())
			b.WriteString("\n\n")
		}
	}

	// Subclasses
	for _, sc := range c.Subclasses {
		b.WriteString("---\n\n")
		fmt.Fprintf(&b, "## %s\n\n", sc.Name)
		if len(sc.Description) > 0 {
			b.WriteString(sc.Description.plainText())
			b.WriteString("\n\n")
		}
		for _, f := range sc.Features {
			fmt.Fprintf(&b, "### %s (Livello %d)\n\n", f.Name, f.Level)
			b.WriteString(f.Description.plainText())
			b.WriteString("\n\n")
		}
	}

	return b.String()
}

func (l *Loader) buildSpeciesMarkdown(s jsonSpecies) string {
	var b strings.Builder

	fmt.Fprintf(&b, "*%s %s, %s*\n\n", s.CreatureType, s.Size, s.Speed)

	b.WriteString(s.Description.plainText())

	return b.String()
}

// buildBackgroundHTML renders background content as HTML.
func (l *Loader) buildBackgroundHTML(bg jsonBackground) string {
	return l.renderContent(l.buildBackgroundContentMarkdown(bg, l.sourceShort()))
}

func (l *Loader) buildBackgroundMarkdown(bg jsonBackground) string {
	return l.buildBackgroundContentMarkdown(bg, "")
}

// buildBackgroundContentMarkdown builds background content as markdown.
// If sourceShort is non-empty, entity references become markdown links.
func (l *Loader) buildBackgroundContentMarkdown(bg jsonBackground, sourceShort string) string {
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
	if len(bg.Description) > 0 {
		if sourceShort != "" {
			b.WriteString(bg.Description.toMarkdown(sourceShort))
		} else {
			b.WriteString(bg.Description.plainText())
		}
	}

	return strings.TrimSpace(b.String())
}
