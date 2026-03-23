package datastore

import (
	"fmt"
	"strings"
)

// jsonSegment mirrors the srd.Segment type from quintaedizione-data-ita.
type jsonSegment struct {
	Type string `json:"type"`
	Text string `json:"text"`
	ID   string `json:"id,omitempty"`
}

// jsonContent is a local alias for []jsonSegment, matching srd.Content.
type jsonContent []jsonSegment

// segmentCollection maps a reference segment type to its SRD collection
// for types that have dedicated pages and can be rendered as links.
var segmentCollection = map[string]string{
	"spell":     "incantesimi",
	"equipment": "equipaggiamenti",
	"rule":      "regole",
	"condition": "glossario",
}

// segmentRuleLink maps reference types that don't have per-entity pages
// to a fixed rules page. These are rendered as markdown links to the
// relevant rules section.
var segmentRuleLink = map[string]string{
	"damage_type": "tipi-di-danno",
	"ability":     "le-sei-caratteristiche",
	"skill":       "competenze-nelle-abilita",
	// creature_type has no dedicated rules page — stays as plain text
}

// markdownLinkEscaper escapes characters that are meaningful inside markdown link text.
var markdownLinkEscaper = strings.NewReplacer("]", `\]`, ")", `\)`)

// plainText concatenates all segment texts, equivalent to Content.PlainText().
func (c jsonContent) plainText() string {
	if len(c) == 0 {
		return ""
	}
	if len(c) == 1 {
		return c[0].Text
	}
	var b strings.Builder
	for _, s := range c {
		b.WriteString(s.Text)
	}
	return b.String()
}

// toMarkdown converts segments to markdown, rendering entity references as
// markdown links. Types with dedicated pages (spells, conditions, etc.) link
// to their entity page. Types with rule pages (abilities, skills, damage types)
// link to the relevant rules section. Unknown types emit plain text.
func (c jsonContent) toMarkdown(sourceShort string) string {
	if len(c) == 0 {
		return ""
	}
	var b strings.Builder
	for _, s := range c {
		if collection, ok := segmentCollection[s.Type]; ok && s.ID != "" {
			fmt.Fprintf(&b, "[%s](/srd/%s/%s/%s)", markdownLinkEscaper.Replace(s.Text), collection, sourceShort, s.ID)
		} else if ruleID, ok := segmentRuleLink[s.Type]; ok && s.ID != "" {
			fmt.Fprintf(&b, "[%s](/srd/regole/%s/%s)", markdownLinkEscaper.Replace(s.Text), sourceShort, ruleID)
		} else {
			b.WriteString(s.Text)
		}
	}
	return b.String()
}
