package parsers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strings"

	nethtml "golang.org/x/net/html"
)

// collectionCategory maps SRD collection URL segments to Italian category labels.
var collectionCategory = map[string]string{
	"incantesimi":     "incantesimo",
	"glossario":       "condizione",
	"equipaggiamenti": "equipaggiamento",
	"regole":          "regola",
	"mostri":          "mostro",
	"classi":          "classe",
	"backgrounds":     "background",
	"talenti":         "talento",
	"specie":          "specie",
	"oggetti_magici":  "oggetto magico",
	"servizi":         "servizio",
}

// CrossLinker wraps glossary terms in HTML content with <span> elements
// carrying data attributes for client-side tooltip display. It unifies
// glossary term enrichment into a single post-processing pass on rendered HTML.
//
// It handles two types of cross-references:
//  1. Data-driven links: <a href="/srd/..."> produced by the markdown pipeline
//     are converted to <span class="glossary-term"> with data-term-link.
//  2. Glossary terms: first occurrence of each term is wrapped with
//     <span class="glossary-term"> with data-term-def.
type CrossLinker struct {
	compiledRegex []*crossLinkRegex
	glossaryByID  map[string]*crossLinkTerm // for merging definitions onto converted links
}

// crossLinkSegment mirrors the srd.Segment type from quintaedizione-data-ita.
type crossLinkSegment struct {
	Type string `json:"type"`
	Text string `json:"text"`
	ID   string `json:"id,omitempty"`
}

// crossLinkContent is a local alias for []crossLinkSegment, matching srd.Content.
type crossLinkContent []crossLinkSegment

func (c crossLinkContent) plainText() string {
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

type crossLinkTerm struct {
	ID         string           `json:"id"`
	Term       string           `json:"term"`
	Category   string           `json:"category"`
	Definition crossLinkContent `json:"definition"`
}

type crossLinkRegex struct {
	term  crossLinkTerm
	regex *regexp.Regexp
}

// NewCrossLinker creates a CrossLinker from the embedded glossary.json.
// It searches for glossary.json in the root or in source subdirectories.
func NewCrossLinker(fsys fs.FS) (*CrossLinker, error) {
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

	var terms []crossLinkTerm
	if err := json.Unmarshal(data, &terms); err != nil {
		return nil, fmt.Errorf("parse glossary.json: %w", err)
	}

	// Sort by term length descending (longest match first)
	sort.Slice(terms, func(i, j int) bool {
		return len(terms[i].Term) > len(terms[j].Term)
	})

	compiled := make([]*crossLinkRegex, 0, len(terms))
	for _, t := range terms {
		pattern := `\b` + regexp.QuoteMeta(t.Term) + `\b`
		re, err := regexp.Compile("(?i)" + pattern)
		if err != nil {
			continue
		}
		compiled = append(compiled, &crossLinkRegex{term: t, regex: re})
	}

	byID := make(map[string]*crossLinkTerm, len(terms))
	for i := range terms {
		byID[terms[i].ID] = &terms[i]
	}

	return &CrossLinker{
		compiledRegex: compiled,
		glossaryByID:  byID,
	}, nil
}

// LinkTerms processes HTML content in two passes:
//  1. Converts internal <a href="/srd/..."> links to <span class="glossary-term"> spans
//  2. Wraps first occurrence of glossary terms in text nodes with <span class="glossary-term">
func (cl *CrossLinker) LinkTerms(htmlContent string) string {
	if htmlContent == "" {
		return htmlContent
	}

	doc, err := nethtml.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent
	}

	linked := make(map[string]bool)

	// Pre-fill linked map with terms already wrapped in glossary-term spans
	prefillLinkedFromExistingSpans(doc, linked)

	// Pass 1: convert internal /srd/ links to glossary-term spans
	cl.convertInternalLinks(doc, linked)

	// Pass 2: glossary term regex matching on text nodes
	if len(cl.compiledRegex) > 0 {
		contentLower := strings.ToLower(htmlContent)
		var active []*crossLinkRegex
		for _, tr := range cl.compiledRegex {
			if strings.Contains(contentLower, strings.ToLower(tr.term.Term)) {
				active = append(active, tr)
			}
		}
		if len(active) > 0 {
			cl.processNode(doc, active, linked)
		}
	}

	return renderBody(doc, htmlContent)
}

// convertInternalLinks finds <a href="/srd/..."> elements and replaces them
// with <span class="glossary-term"> carrying data-term-link, data-term-id, data-term-cat.
// If the term has a glossary entry, data-term-def is also added.
func (cl *CrossLinker) convertInternalLinks(n *nethtml.Node, linked map[string]bool) {
	// Collect links first, then convert (avoid modifying tree during traversal)
	var links []*nethtml.Node
	collectInternalLinks(n, &links)

	for _, a := range links {
		href := getAttr(a, "href")
		if href == "" {
			continue
		}

		collection, id := parseSRDLink(href)
		if id == "" {
			continue
		}

		cat := collectionCategory[collection]
		if cat == "" {
			cat = collection
		}

		// Mark as linked to prevent glossary regex from creating a duplicate
		linked[id] = true

		// Build span attributes
		attrs := []nethtml.Attribute{
			{Key: "class", Val: "glossary-term"},
			{Key: "data-term-id", Val: id},
			{Key: "data-term-cat", Val: cat},
			{Key: "data-term-link", Val: href},
			{Key: "tabindex", Val: "0"},
		}

		// Merge glossary definition if available
		if term, ok := cl.glossaryByID[id]; ok {
			truncDef := truncateDefinition(term.Definition.plainText(), 250)
			if truncDef != "" {
				attrs = append(attrs, nethtml.Attribute{
					Key: "data-term-def",
					Val: truncDef,
				})
			}
			// Use glossary category if more specific
			if term.Category != "" {
				for i := range attrs {
					if attrs[i].Key == "data-term-cat" {
						attrs[i].Val = term.Category
						break
					}
				}
			}
		}

		// Create replacement span
		span := &nethtml.Node{
			Type: nethtml.ElementNode,
			Data: "span",
			Attr: attrs,
		}

		// Move link's children to the span
		for c := a.FirstChild; c != nil; {
			next := c.NextSibling
			a.RemoveChild(c)
			span.AppendChild(c)
			c = next
		}

		// Replace <a> with <span> in parent
		a.Parent.InsertBefore(span, a)
		a.Parent.RemoveChild(a)
	}
}

func collectInternalLinks(n *nethtml.Node, links *[]*nethtml.Node) {
	if n.Type == nethtml.ElementNode && n.Data == "a" {
		href := getAttr(n, "href")
		if strings.HasPrefix(href, "/srd/") {
			*links = append(*links, n)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectInternalLinks(c, links)
	}
}

func getAttr(n *nethtml.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// parseSRDLink extracts the collection and entity ID from a /srd/ URL.
// e.g., "/srd/glossario/5.5e/accecato" → ("glossario", "accecato")
// e.g., "/srd/regole/5.5e/le-sei-caratteristiche" → ("regole", "le-sei-caratteristiche")
func parseSRDLink(href string) (collection, id string) {
	// Expected format: /srd/{collection}/{source}/{id}
	parts := strings.Split(strings.TrimPrefix(href, "/srd/"), "/")
	if len(parts) >= 3 {
		return parts[0], parts[2]
	}
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}
	return "", ""
}

func (cl *CrossLinker) processNode(n *nethtml.Node, active []*crossLinkRegex, linked map[string]bool) {
	if n.Type == nethtml.TextNode {
		if !isInSkippedTag(n) {
			cl.replaceTextNode(n, active, linked)
		}
		return
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		cl.processNode(c, active, linked)
	}
}

func isInSkippedTag(n *nethtml.Node) bool {
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

func (cl *CrossLinker) replaceTextNode(n *nethtml.Node, active []*crossLinkRegex, linked map[string]bool) {
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
		truncDef := truncateDefinition(tr.term.Definition.plainText(), 250)
		spanNode := &nethtml.Node{
			Type: nethtml.ElementNode,
			Data: "span",
			Attr: []nethtml.Attribute{
				{Key: "class", Val: "glossary-term"},
				{Key: "data-term-id", Val: tr.term.ID},
				{Key: "data-term-def", Val: truncDef},
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

// stripMarkdown removes common markdown formatting markers from text
// for use in plain-text contexts like tooltip attributes.
var markdownStripper = strings.NewReplacer(
	"**", "",
	"__", "",
	"*", "",
	"_", "",
)

func stripMarkdown(s string) string {
	return markdownStripper.Replace(s)
}

func truncateDefinition(def string, maxLen int) string {
	def = stripMarkdown(def)
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

// prefillLinkedFromExistingSpans scans the DOM for <span class="glossary-term">
// elements and marks their data-term-id as already linked, preventing duplicate wrapping.
func prefillLinkedFromExistingSpans(n *nethtml.Node, linked map[string]bool) {
	if n.Type == nethtml.ElementNode && n.Data == "span" {
		for _, attr := range n.Attr {
			if attr.Key == "data-term-id" && attr.Val != "" {
				linked[attr.Val] = true
				break
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		prefillLinkedFromExistingSpans(c, linked)
	}
}

// renderBody renders only the children of the <body> tag,
// stripping the <html><head><body> wrapper that html.Parse adds.
func renderBody(doc *nethtml.Node, fallback string) string {
	body := findBodyNode(doc)
	if body == nil {
		var buf bytes.Buffer
		if err := nethtml.Render(&buf, doc); err != nil {
			return fallback
		}
		return buf.String()
	}

	var buf bytes.Buffer
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		if err := nethtml.Render(&buf, c); err != nil {
			return fallback
		}
	}
	return buf.String()
}

func findBodyNode(n *nethtml.Node) *nethtml.Node {
	if n.Type == nethtml.ElementNode && n.Data == "body" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findBodyNode(c); found != nil {
			return found
		}
	}
	return nil
}
