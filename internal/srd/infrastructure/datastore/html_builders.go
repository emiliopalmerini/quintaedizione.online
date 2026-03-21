package datastore

import (
	"fmt"
	"html"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/application/parsers"
)

// escapeHTML wraps html.EscapeString for convenience.
func escapeHTML(s string) string {
	return html.EscapeString(s)
}

// renderInline renders markdown through the renderer and strips the wrapping
// <p>…</p> tags so the result can be used inline inside other HTML elements.
func renderInline(renderer *parsers.MarkdownRenderer, md string) string {
	h := renderer.Render(md)
	h = strings.TrimPrefix(h, "<p>")
	h = strings.TrimSuffix(h, "</p>")
	return strings.TrimSpace(h)
}

// ---------------------------------------------------------------------------
// HTML builders — produce ready-to-use HTML for the "content" field
// ---------------------------------------------------------------------------

func (l *Loader) buildSpellHTML(s jsonSpell) string {
	var b strings.Builder

	b.WriteString(`<div class="stat-block">`)

	// Subtitle
	if s.Level == 0 {
		fmt.Fprintf(&b, `<p class="stat-block-subtitle">Trucchetto %s</p>`, escapeHTML(s.School))
	} else {
		fmt.Fprintf(&b, `<p class="stat-block-subtitle">Livello %d %s</p>`, s.Level, escapeHTML(s.School))
	}

	// Properties
	b.WriteString(`<div class="stat-block-properties">`)
	fmt.Fprintf(&b, `<div class="stat-block-property"><strong>Tempo di Lancio:</strong> %s</div>`, escapeHTML(s.CastingTime))
	fmt.Fprintf(&b, `<div class="stat-block-property"><strong>Gittata:</strong> %s</div>`, escapeHTML(s.Range))
	fmt.Fprintf(&b, `<div class="stat-block-property"><strong>Componenti:</strong> %s</div>`, escapeHTML(s.Components))
	fmt.Fprintf(&b, `<div class="stat-block-property"><strong>Durata:</strong> %s</div>`, escapeHTML(s.Duration))
	b.WriteString(`</div>`)

	if s.Ritual {
		b.WriteString(`<p class="stat-block-ritual">Rituale</p>`)
	}

	// Description (markdown → HTML with glossary linking)
	if s.Description != "" {
		fmt.Fprintf(&b, `<div class="stat-block-description">%s</div>`, l.renderer.Render(s.Description))
	}

	// At higher levels
	if s.AtHigherLevels != "" {
		fmt.Fprintf(&b, `<div class="stat-block-higher-levels"><strong>Ai Livelli Superiori.</strong> %s</div>`,
			renderInline(l.renderer, s.AtHigherLevels))
	}

	// Classes
	if len(s.Classes) > 0 {
		fmt.Fprintf(&b, `<p class="stat-block-classes">Classi: %s</p>`, escapeHTML(strings.Join(s.Classes, ", ")))
	}

	b.WriteString(`</div>`)
	return b.String()
}

func (l *Loader) buildMonsterHTML(m jsonMonster) string {
	var b strings.Builder

	b.WriteString(`<div class="stat-block">`)

	// Subtitle (type + size + alignment)
	fmt.Fprintf(&b, `<p class="stat-block-subtitle">%s %s, %s</p>`,
		escapeHTML(m.Type), escapeHTML(m.Size), escapeHTML(m.Alignment))

	// Core stats
	b.WriteString(`<hr class="stat-block-divider">`)
	b.WriteString(`<div class="stat-block-properties">`)
	acLine := escapeHTML(m.AC)
	if m.Initiative != "" {
		acLine += " &middot; <strong>Iniziativa</strong> " + escapeHTML(m.Initiative)
	}
	fmt.Fprintf(&b, `<div class="stat-block-property"><strong>CA</strong> %s</div>`, acLine)
	fmt.Fprintf(&b, `<div class="stat-block-property"><strong>PF</strong> %s</div>`, escapeHTML(m.HP))
	fmt.Fprintf(&b, `<div class="stat-block-property"><strong>Velocità</strong> %s</div>`, escapeHTML(m.Speed))
	b.WriteString(`</div>`)

	// Ability scores grid
	if len(m.AbilityScores) > 0 {
		b.WriteString(`<hr class="stat-block-divider">`)
		b.WriteString(`<div class="stat-block-abilities">`)

		abilityOrder := []struct{ key, label string }{
			{"strength", "FOR"},
			{"dexterity", "DES"},
			{"constitution", "COS"},
			{"intelligence", "INT"},
			{"wisdom", "SAG"},
			{"charisma", "CAR"},
		}

		for _, a := range abilityOrder {
			score := m.AbilityScores[a.key]
			mod := m.AbilityMods[a.key]
			save := m.SavingThrows[a.key]

			b.WriteString(`<div class="stat-block-ability">`)
			fmt.Fprintf(&b, `<div class="stat-block-ability-label">%s</div>`, a.label)
			fmt.Fprintf(&b, `<div class="stat-block-ability-score">%d</div>`, score)
			fmt.Fprintf(&b, `<div class="stat-block-ability-mod">%+d</div>`, mod)
			fmt.Fprintf(&b, `<div class="stat-block-ability-save">TS %s</div>`, escapeHTML(save))
			b.WriteString(`</div>`)
		}

		b.WriteString(`</div>`)
	}

	// Secondary stats
	b.WriteString(`<hr class="stat-block-divider">`)
	b.WriteString(`<div class="stat-block-properties">`)

	writeProperty := func(label, value string) {
		if value != "" {
			fmt.Fprintf(&b, `<div class="stat-block-property"><strong>%s</strong> %s</div>`,
				escapeHTML(label), escapeHTML(value))
		}
	}

	writeProperty("Abilità", m.Skills)
	writeProperty("Resistenze", m.Resistances)
	writeProperty("Immunità ai Danni", m.DamageImmunities)
	writeProperty("Immunità alle Condizioni", m.ConditionImmunities)
	writeProperty("Sensi", m.Senses)
	writeProperty("Lingue", m.Languages)

	if m.CRDetail != "" {
		writeProperty("GS", m.CRDetail)
	} else {
		writeProperty("GS", m.CR)
	}
	writeProperty("Equipaggiamento", m.Equipment)

	b.WriteString(`</div>`)

	// Traits
	if len(m.Traits) > 0 {
		b.WriteString(`<hr class="stat-block-divider">`)
		writeFeatures(&b, l.renderer, m.Traits)
	}

	// Actions
	writeFeatureSection(&b, l.renderer, "Azioni", m.Actions)
	writeFeatureSection(&b, l.renderer, "Azioni Bonus", m.BonusActions)
	writeFeatureSection(&b, l.renderer, "Reazioni", m.Reactions)
	writeFeatureSection(&b, l.renderer, "Azioni Leggendarie", m.LegendaryActions)

	b.WriteString(`</div>`)
	return b.String()
}

func writeFeatures(b *strings.Builder, renderer *parsers.MarkdownRenderer, features []jsonFeature) {
	for _, f := range features {
		fmt.Fprintf(b, `<div class="stat-block-feature"><span class="stat-block-feature-name">%s.</span> %s</div>`,
			escapeHTML(f.Name), renderInline(renderer, f.Description))
	}
}

func writeFeatureSection(b *strings.Builder, renderer *parsers.MarkdownRenderer, heading string, features []jsonFeature) {
	if len(features) == 0 {
		return
	}
	fmt.Fprintf(b, `<div class="stat-block-section"><h3 class="stat-block-section-heading">%s</h3>`, escapeHTML(heading))
	writeFeatures(b, renderer, features)
	b.WriteString(`</div>`)
}

func (l *Loader) buildClassHTML(c jsonClass) string {
	var b strings.Builder

	b.WriteString(`<div class="stat-block">`)

	// Description (may contain level progression table — already markdown)
	if c.Description != "" {
		fmt.Fprintf(&b, `<div class="stat-block-description">%s</div>`, l.renderer.Render(c.Description))
	}

	// Proficiencies (contains hit points, competencies, equipment, and class table — all markdown)
	if c.Proficiencies != "" {
		fmt.Fprintf(&b, `<div class="stat-block-description">%s</div>`, l.renderer.Render(c.Proficiencies))
	}

	// Features grouped by level
	if len(c.Features) > 0 {
		b.WriteString(`<hr class="stat-block-divider">`)
		b.WriteString(`<h2>Privilegi di classe</h2>`)
		for _, f := range c.Features {
			fmt.Fprintf(&b, `<h3>%s (Livello %d)</h3>`, escapeHTML(f.Name), f.Level)
			fmt.Fprintf(&b, `<div class="stat-block-description">%s</div>`, l.renderer.Render(f.Description))
		}
	}

	// Subclasses
	for _, sc := range c.Subclasses {
		b.WriteString(`<hr class="stat-block-divider">`)
		fmt.Fprintf(&b, `<h2>%s</h2>`, escapeHTML(sc.Name))
		if sc.Description != "" {
			fmt.Fprintf(&b, `<div class="stat-block-description">%s</div>`, l.renderer.Render(sc.Description))
		}
		for _, f := range sc.Features {
			fmt.Fprintf(&b, `<h3>%s (Livello %d)</h3>`, escapeHTML(f.Name), f.Level)
			fmt.Fprintf(&b, `<div class="stat-block-description">%s</div>`, l.renderer.Render(f.Description))
		}
	}

	b.WriteString(`</div>`)
	return b.String()
}

func (l *Loader) buildSpeciesHTML(s jsonSpecies) string {
	var b strings.Builder

	b.WriteString(`<div class="stat-block">`)

	// Subtitle
	fmt.Fprintf(&b, `<p class="stat-block-subtitle">%s %s, %s</p>`,
		escapeHTML(s.CreatureType), escapeHTML(s.Size), escapeHTML(s.Speed))

	// Description
	if s.Description != "" {
		fmt.Fprintf(&b, `<div class="stat-block-description">%s</div>`, l.renderer.Render(s.Description))
	}

	b.WriteString(`</div>`)
	return b.String()
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

	b.WriteString(s.Description)

	if s.AtHigherLevels != "" {
		b.WriteString("\n\n**Ai Livelli Superiori.** ")
		b.WriteString(s.AtHigherLevels)
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
	if m.Resistances != "" {
		fmt.Fprintf(&b, "**Resistenze** %s\n\n", m.Resistances)
	}
	if m.DamageImmunities != "" {
		fmt.Fprintf(&b, "**Immunità ai Danni** %s\n\n", m.DamageImmunities)
	}
	if m.ConditionImmunities != "" {
		fmt.Fprintf(&b, "**Immunità alle Condizioni** %s\n\n", m.ConditionImmunities)
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
			fmt.Fprintf(&b, "***%s.*** %s\n\n", t.Name, t.Description)
		}
	}

	// Actions
	if len(m.Actions) > 0 {
		b.WriteString("### Azioni\n\n")
		for _, a := range m.Actions {
			fmt.Fprintf(&b, "***%s.*** %s\n\n", a.Name, a.Description)
		}
	}

	// Bonus Actions
	if len(m.BonusActions) > 0 {
		b.WriteString("### Azioni Bonus\n\n")
		for _, a := range m.BonusActions {
			fmt.Fprintf(&b, "***%s.*** %s\n\n", a.Name, a.Description)
		}
	}

	// Reactions
	if len(m.Reactions) > 0 {
		b.WriteString("### Reazioni\n\n")
		for _, r := range m.Reactions {
			fmt.Fprintf(&b, "***%s.*** %s\n\n", r.Name, r.Description)
		}
	}

	// Legendary Actions
	if len(m.LegendaryActions) > 0 {
		b.WriteString("### Azioni Leggendarie\n\n")
		for _, la := range m.LegendaryActions {
			fmt.Fprintf(&b, "***%s.*** %s\n\n", la.Name, la.Description)
		}
	}

	return b.String()
}

func (l *Loader) buildClassMarkdown(c jsonClass) string {
	var b strings.Builder

	// Description (contains traits box + level progression table)
	if c.Description != "" {
		b.WriteString(c.Description)
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
			b.WriteString(f.Description)
			b.WriteString("\n\n")
		}
	}

	// Subclasses
	for _, sc := range c.Subclasses {
		b.WriteString("---\n\n")
		fmt.Fprintf(&b, "## %s\n\n", sc.Name)
		if sc.Description != "" {
			b.WriteString(sc.Description)
			b.WriteString("\n\n")
		}
		for _, f := range sc.Features {
			fmt.Fprintf(&b, "### %s (Livello %d)\n\n", f.Name, f.Level)
			b.WriteString(f.Description)
			b.WriteString("\n\n")
		}
	}

	return b.String()
}

func (l *Loader) buildSpeciesMarkdown(s jsonSpecies) string {
	var b strings.Builder

	fmt.Fprintf(&b, "*%s %s, %s*\n\n", s.CreatureType, s.Size, s.Speed)

	b.WriteString(s.Description)

	return b.String()
}
