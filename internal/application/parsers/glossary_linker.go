package parsers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io/fs"
	"regexp"
	"sort"
	"strings"

	nethtml "golang.org/x/net/html"
)

// GlossaryLinker wraps glossary terms in HTML content with <span> elements
// carrying data attributes for client-side tooltip display.
type GlossaryLinker struct {
	terms         []glossaryTerm
	compiledRegex []*termRegex
}

type glossaryTerm struct {
	ID         string `json:"id"`
	Term       string `json:"term"`
	Category   string `json:"category"`
	Definition string `json:"definition"`
}

type termRegex struct {
	term  glossaryTerm
	regex *regexp.Regexp
}

// NewGlossaryLinker creates a GlossaryLinker from the embedded glossary.json.
// It searches for glossary.json in the root or in source subdirectories.
func NewGlossaryLinker(fsys fs.FS) (*GlossaryLinker, error) {
	data, err := fs.ReadFile(fsys, "glossary.json")
	if err != nil {
		// Try to find glossary.json in source subdirectories
		entries, dirErr := fs.ReadDir(fsys, ".")
		if dirErr != nil {
			return nil, fmt.Errorf("read glossary.json: %w", err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			if d, e := fs.ReadFile(fsys, entry.Name()+"/glossary.json"); e == nil {
				data = d
				err = nil
				break
			}
		}
		if err != nil {
			return nil, fmt.Errorf("read glossary.json: %w", err)
		}
	}

	var terms []glossaryTerm
	if err := json.Unmarshal(data, &terms); err != nil {
		return nil, fmt.Errorf("parse glossary.json: %w", err)
	}

	// Sort by term length descending (longest match first)
	sort.Slice(terms, func(i, j int) bool {
		return len(terms[i].Term) > len(terms[j].Term)
	})

	compiled := make([]*termRegex, 0, len(terms))
	for _, t := range terms {
		pattern := `\b` + regexp.QuoteMeta(t.Term) + `\b`
		re, err := regexp.Compile("(?i)" + pattern)
		if err != nil {
			continue
		}
		compiled = append(compiled, &termRegex{term: t, regex: re})
	}

	return &GlossaryLinker{
		terms:         terms,
		compiledRegex: compiled,
	}, nil
}

// LinkGlossaryTerms wraps the first occurrence of each glossary term in the
// given HTML content with a <span class="glossary-term"> element.
func (gl *GlossaryLinker) LinkGlossaryTerms(htmlContent string) string {
	if htmlContent == "" || len(gl.compiledRegex) == 0 {
		return htmlContent
	}

	// Pre-filter: only process terms that appear in the content
	contentLower := strings.ToLower(htmlContent)
	var active []*termRegex
	for _, tr := range gl.compiledRegex {
		if strings.Contains(contentLower, strings.ToLower(tr.term.Term)) {
			active = append(active, tr)
		}
	}
	if len(active) == 0 {
		return htmlContent
	}

	doc, err := nethtml.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent
	}

	linked := make(map[string]bool)
	gl.processNode(doc, active, linked)

	return renderGlossaryHTML(doc, htmlContent)
}

func (gl *GlossaryLinker) processNode(n *nethtml.Node, active []*termRegex, linked map[string]bool) {
	if n.Type == nethtml.TextNode {
		if !isInSkippedGlossaryTag(n) {
			gl.replaceTextNode(n, active, linked)
		}
		return
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		gl.processNode(c, active, linked)
	}
}

func isInSkippedGlossaryTag(n *nethtml.Node) bool {
	skippedTags := map[string]bool{
		"code": true, "pre": true, "a": true,
		"h1": true, "h2": true, "h3": true,
		"h4": true, "h5": true, "h6": true,
		"span": true, // don't nest spans inside existing glossary-term spans
	}

	for parent := n.Parent; parent != nil; parent = parent.Parent {
		if parent.Type == nethtml.ElementNode && skippedTags[parent.Data] {
			return true
		}
	}

	return false
}

func (gl *GlossaryLinker) replaceTextNode(n *nethtml.Node, active []*termRegex, linked map[string]bool) {
	text := n.Data
	parent := n.Parent
	if parent == nil {
		return
	}

	for _, tr := range active {
		if linked[tr.term.ID] {
			continue
		}

		loc := tr.regex.FindStringIndex(text)
		if loc == nil {
			continue
		}

		linked[tr.term.ID] = true

		// Build replacement nodes
		var nodes []*nethtml.Node

		// Text before match
		if loc[0] > 0 {
			nodes = append(nodes, &nethtml.Node{
				Type: nethtml.TextNode,
				Data: text[:loc[0]],
			})
		}

		// Glossary span
		truncDef := truncateDefinition(tr.term.Definition, 250)
		spanNode := &nethtml.Node{
			Type: nethtml.ElementNode,
			Data: "span",
			Attr: []nethtml.Attribute{
				{Key: "class", Val: "glossary-term"},
				{Key: "data-term-id", Val: tr.term.ID},
				{Key: "data-term-def", Val: html.EscapeString(truncDef)},
				{Key: "data-term-cat", Val: tr.term.Category},
				{Key: "tabindex", Val: "0"},
			},
		}
		spanText := &nethtml.Node{
			Type: nethtml.TextNode,
			Data: text[loc[0]:loc[1]],
		}
		spanNode.AppendChild(spanText)
		nodes = append(nodes, spanNode)

		// Text after match
		if loc[1] < len(text) {
			nodes = append(nodes, &nethtml.Node{
				Type: nethtml.TextNode,
				Data: text[loc[1]:],
			})
		}

		// Replace original node
		prev := n.PrevSibling
		parent.RemoveChild(n)

		for _, node := range nodes {
			if prev == nil {
				parent.InsertBefore(node, parent.FirstChild)
				prev = node
			} else {
				if prev.NextSibling != nil {
					parent.InsertBefore(node, prev.NextSibling)
				} else {
					parent.AppendChild(node)
				}
				prev = node
			}
		}

		return
	}
}

func truncateDefinition(def string, maxLen int) string {
	if len(def) <= maxLen {
		return def
	}
	// Find last space before maxLen to avoid cutting mid-word
	truncated := def[:maxLen]
	if lastSpace := strings.LastIndex(truncated, " "); lastSpace > maxLen/2 {
		truncated = truncated[:lastSpace]
	}
	return truncated + "…"
}

func renderGlossaryHTML(doc *nethtml.Node, fallback string) string {
	var buf bytes.Buffer
	if err := nethtml.Render(&buf, doc); err != nil {
		return fallback
	}
	return buf.String()
}
