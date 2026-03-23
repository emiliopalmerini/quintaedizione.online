package datastore

import (
	"strings"

	nethtml "golang.org/x/net/html"
)

// EnrichDescriptions adds data-term-def attributes to crosslinked spans
// that have data-term-link but no data-term-def. It uses short descriptions
// built from each entity's raw_content field.
// Must be called after LoadAll() so all collections are available.
func EnrichDescriptions(data map[string][]map[string]any) {
	index := buildDescriptionIndex(data)
	if len(index) == 0 {
		return
	}
	for _, docs := range data {
		for _, doc := range docs {
			enrichHTMLFields(doc, index)
		}
	}
}

// buildDescriptionIndex creates a map from "source/id" → short description
// for all loaded entities.
func buildDescriptionIndex(data map[string][]map[string]any) map[string]string {
	index := make(map[string]string)
	for _, docs := range data {
		for _, doc := range docs {
			id, _ := doc["_id"].(string)
			sourceShort, _ := doc["_source_short"].(string)
			if id == "" || sourceShort == "" {
				continue
			}

			// Prefer description_html (stat-block entities) over raw_content,
			// because raw_content for spells starts with properties (casting time,
			// range, etc.) which are useless in a tooltip.
			var desc string
			if descHTML, ok := doc["description_html"].(string); ok && descHTML != "" {
				desc = truncatePreview(stripHTMLTags(descHTML), 200)
			} else if raw, ok := doc["raw_content"].(string); ok && raw != "" {
				desc = truncatePreview(raw, 200)
			}

			if desc != "" {
				index[sourceShort+"/"+id] = desc
			}
		}
	}
	return index
}

// stripHTMLTags removes HTML tags from a string, returning plain text.
func stripHTMLTags(s string) string {
	doc, err := nethtml.Parse(strings.NewReader(s))
	if err != nil {
		return s
	}
	var b strings.Builder
	extractText(doc, &b)
	return b.String()
}

func extractText(n *nethtml.Node, b *strings.Builder) {
	if n.Type == nethtml.TextNode {
		b.WriteString(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractText(c, b)
	}
}

// truncatePreview extracts the first meaningful text from raw markdown content,
// stripping markdown formatting and truncating at word boundary.
func truncatePreview(raw string, maxLen int) string {
	if raw == "" {
		return ""
	}
	// Skip the first line if it's a subtitle/metadata (starts with * or ---)
	lines := strings.SplitN(raw, "\n", 3)
	text := raw
	if len(lines) > 1 && (strings.HasPrefix(lines[0], "*") || strings.HasPrefix(lines[0], "---")) {
		// Find the first non-empty line after metadata
		rest := strings.TrimLeft(strings.Join(lines[1:], "\n"), "\n ")
		if rest != "" {
			text = rest
		}
	}

	// Strip markdown formatting
	text = markdownStripReplacer.Replace(text)
	// Collapse whitespace
	text = strings.Join(strings.Fields(text), " ")

	if len(text) <= maxLen {
		return text
	}
	truncated := text[:maxLen]
	if lastSpace := strings.LastIndex(truncated, " "); lastSpace > maxLen/2 {
		truncated = truncated[:lastSpace]
	}
	return truncated + "…"
}

var markdownStripReplacer = strings.NewReplacer(
	"**", "", "__", "", "*", "", "_", "",
	"###", "", "##", "", "#", "",
)

// enrichHTMLFields recursively finds string fields containing glossary-term
// spans and adds data-term-def where missing.
func enrichHTMLFields(doc map[string]any, index map[string]string) {
	for key, val := range doc {
		switch v := val.(type) {
		case string:
			if strings.Contains(v, `data-term-link=`) {
				doc[key] = addDescriptionsToSpans(v, index)
			}
		case []map[string]any:
			for _, item := range v {
				enrichHTMLFields(item, index)
			}
		}
	}
}

// addDescriptionsToSpans parses HTML and adds data-term-def to glossary-term
// spans that have data-term-link but no data-term-def.
func addDescriptionsToSpans(htmlContent string, index map[string]string) string {
	doc, err := nethtml.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent
	}

	modified := false
	addDefsToNode(doc, index, &modified)

	if !modified {
		return htmlContent
	}
	return renderHTMLBody(doc, htmlContent)
}

func addDefsToNode(n *nethtml.Node, index map[string]string, modified *bool) {
	if n.Type == nethtml.ElementNode && n.Data == "span" {
		var hasLink, hasDef bool
		var link string
		for _, attr := range n.Attr {
			switch attr.Key {
			case "data-term-link":
				hasLink = true
				link = attr.Val
			case "data-term-def":
				hasDef = true
			}
		}
		if hasLink && !hasDef && link != "" {
			// Extract source/id from link: /srd/{collection}/{source}/{id}
			parts := strings.Split(strings.TrimPrefix(link, "/srd/"), "/")
			if len(parts) >= 3 {
				key := parts[1] + "/" + parts[2] // source/id
				if desc, ok := index[key]; ok && desc != "" {
					n.Attr = append(n.Attr, nethtml.Attribute{
						Key: "data-term-def",
						Val: desc,
					})
					*modified = true
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		addDefsToNode(c, index, modified)
	}
}

// renderHTMLBody renders the <body> children, stripping the wrapper.
func renderHTMLBody(doc *nethtml.Node, fallback string) string {
	var body *nethtml.Node
	var findBody func(*nethtml.Node)
	findBody = func(n *nethtml.Node) {
		if n.Type == nethtml.ElementNode && n.Data == "body" {
			body = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findBody(c)
			if body != nil {
				return
			}
		}
	}
	findBody(doc)

	if body == nil {
		return fallback
	}

	var buf strings.Builder
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		if err := nethtml.Render(&buf, c); err != nil {
			return fallback
		}
	}
	return buf.String()
}
