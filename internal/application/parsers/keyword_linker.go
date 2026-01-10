package parsers

import (
	"bytes"
	"encoding/json"
	"os"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

// KeywordLinker adds hyperlinks to game terminology in HTML content
type KeywordLinker struct {
	keywords map[string]string
	linked   map[string]bool
}

type keywordConfig struct {
	DamageTypes   map[string]string `json:"damage_types"`
	Conditions    map[string]string `json:"conditions"`
	CoreMechanics map[string]string `json:"core_mechanics"`
	CreatureTypes map[string]string `json:"creature_types"`
}

// NewKeywordLinker creates a KeywordLinker from a JSON configuration file
func NewKeywordLinker(configPath string) (*KeywordLinker, error) {
	config, err := loadKeywordConfig(configPath)
	if err != nil {
		return nil, err
	}

	return &KeywordLinker{
		keywords: flattenKeywords(config),
		linked:   make(map[string]bool),
	}, nil
}

// LinkKeywords adds links to keywords in HTML content (first occurrence only per document)
func (kl *KeywordLinker) LinkKeywords(htmlContent string) string {
	if htmlContent == "" {
		return htmlContent
	}

	kl.resetLinkedTracker()

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent
	}

	kl.processNode(doc)

	return renderHTML(doc, htmlContent)
}

// loadKeywordConfig reads and parses the keywords JSON file
func loadKeywordConfig(path string) (*keywordConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config keywordConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// flattenKeywords merges all keyword categories into a single map
func flattenKeywords(config *keywordConfig) map[string]string {
	keywords := make(map[string]string)

	mergeKeywords(keywords, config.DamageTypes)
	mergeKeywords(keywords, config.Conditions)
	mergeKeywords(keywords, config.CoreMechanics)
	mergeKeywords(keywords, config.CreatureTypes)

	return keywords
}

// mergeKeywords adds keywords from source to destination map
func mergeKeywords(dest, source map[string]string) {
	for keyword, url := range source {
		dest[keyword] = url
	}
}

// resetLinkedTracker clears the linked keywords tracker for a new document
func (kl *KeywordLinker) resetLinkedTracker() {
	kl.linked = make(map[string]bool)
}

// processNode recursively processes HTML nodes to add keyword links
func (kl *KeywordLinker) processNode(n *html.Node) {
	// Handle text nodes
	if n.Type == html.TextNode {
		if !isInSkippedTag(n) {
			kl.replaceTextNodeWithLinks(n)
		}
		return
	}

	// Process children for all other node types
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		kl.processNode(c)
	}
}

// isInSkippedTag checks if node is inside a tag where linking should be disabled
func isInSkippedTag(n *html.Node) bool {
	skippedTags := map[string]bool{
		"code": true, "pre": true, "a": true,
		"h1": true, "h2": true, "h3": true,
		"h4": true, "h5": true, "h6": true,
	}

	for parent := n.Parent; parent != nil; parent = parent.Parent {
		if parent.Type == html.ElementNode && skippedTags[parent.Data] {
			return true
		}
	}

	return false
}

// replaceTextNodeWithLinks replaces a text node with text + link nodes for keywords
func (kl *KeywordLinker) replaceTextNodeWithLinks(n *html.Node) {
	text := n.Data
	parent := n.Parent

	if parent == nil {
		return
	}

	// Find the first matching keyword
	for _, keyword := range kl.sortedKeywordsByLength() {
		if kl.linked[keyword] {
			continue
		}

		pattern := `\b` + regexp.QuoteMeta(keyword) + `\b`
		re := regexp.MustCompile(pattern)

		loc := re.FindStringIndex(text)
		if loc == nil {
			continue
		}

		// Mark as linked
		kl.linked[keyword] = true

		// Split text into: before + keyword + after
		before := text[:loc[0]]
		matched := text[loc[0]:loc[1]]
		after := text[loc[1]:]

		// Create nodes: text(before) + link(keyword) + text(after)
		nodes := make([]*html.Node, 0, 3)

		if before != "" {
			nodes = append(nodes, &html.Node{
				Type: html.TextNode,
				Data: before,
			})
		}

		// Create link node
		linkNode := &html.Node{
			Type: html.ElementNode,
			Data: "a",
			Attr: []html.Attribute{
				{Key: "href", Val: kl.keywords[keyword]},
				{Key: "class", Val: "keyword-link"},
			},
		}
		linkText := &html.Node{
			Type: html.TextNode,
			Data: matched,
		}
		linkNode.AppendChild(linkText)
		nodes = append(nodes, linkNode)

		if after != "" {
			nodes = append(nodes, &html.Node{
				Type: html.TextNode,
				Data: after,
			})
		}

		// Replace the original node with new nodes
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

// sortedKeywordsByLength returns keywords sorted by length (longest first)
// This prevents partial matching of shorter keywords within longer ones
func (kl *KeywordLinker) sortedKeywordsByLength() []string {
	keywords := make([]string, 0, len(kl.keywords))
	for keyword := range kl.keywords {
		keywords = append(keywords, keyword)
	}

	sort.Slice(keywords, func(i, j int) bool {
		return len(keywords[i]) > len(keywords[j])
	})

	return keywords
}

// renderHTML converts HTML node tree back to string
func renderHTML(doc *html.Node, fallback string) string {
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return fallback
	}
	return buf.String()
}
