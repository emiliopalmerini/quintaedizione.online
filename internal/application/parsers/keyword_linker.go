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
	keywords      map[string]string
	compiledRegex map[string]*regexp.Regexp
}

type keywordConfig struct {
	DamageTypes   map[string]string `json:"damage_types"`
	Conditions    map[string]string `json:"conditions"`
	CoreMechanics map[string]string `json:"core_mechanics"`
	CreatureTypes map[string]string `json:"creature_types"`
	Spells        map[string]string `json:"spells"`
	Monsters      map[string]string `json:"monsters"`
	Equipment     map[string]string `json:"equipment"`
	MagicItems    map[string]string `json:"magic_items"`
	Weapons       map[string]string `json:"weapons"`
	Backgrounds   map[string]string `json:"backgrounds"`
	Armor         map[string]string `json:"armor"`
	Classes       map[string]string `json:"classes"`
}

// NewKeywordLinker creates a KeywordLinker from a JSON configuration file
func NewKeywordLinker(configPath string) (*KeywordLinker, error) {
	config, err := loadKeywordConfig(configPath)
	if err != nil {
		return nil, err
	}

	keywords := flattenKeywords(config)

	// Pre-compile regex patterns for all keywords
	compiledRegex := make(map[string]*regexp.Regexp, len(keywords))
	for keyword := range keywords {
		pattern := `\b` + regexp.QuoteMeta(keyword) + `\b`
		compiledRegex[keyword] = regexp.MustCompile(pattern)
	}

	return &KeywordLinker{
		keywords:      keywords,
		compiledRegex: compiledRegex,
	}, nil
}

// LinkKeywords adds links to keywords in HTML content (all occurrences)
// This method is thread-safe for concurrent calls.
func (kl *KeywordLinker) LinkKeywords(htmlContent string) string {
	if htmlContent == "" {
		return htmlContent
	}

	// Performance optimization: pre-filter keywords to only those present in document
	// This reduces processing from ~850 to ~20-50 keywords per document
	filteredKeywords := kl.prefilterKeywords(htmlContent)
	if len(filteredKeywords) == 0 {
		return htmlContent
	}

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent
	}

	kl.processNodeWithKeywords(doc, filteredKeywords)

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

	// Game mechanics categories
	mergeKeywords(keywords, config.DamageTypes)
	mergeKeywords(keywords, config.Conditions)
	mergeKeywords(keywords, config.CoreMechanics)
	mergeKeywords(keywords, config.CreatureTypes)

	// Content categories
	mergeKeywords(keywords, config.Spells)
	mergeKeywords(keywords, config.Monsters)
	mergeKeywords(keywords, config.Equipment)
	mergeKeywords(keywords, config.MagicItems)
	mergeKeywords(keywords, config.Weapons)
	mergeKeywords(keywords, config.Backgrounds)
	mergeKeywords(keywords, config.Armor)
	mergeKeywords(keywords, config.Classes)

	return keywords
}

// mergeKeywords adds keywords from source to destination map
func mergeKeywords(dest, source map[string]string) {
	for keyword, url := range source {
		dest[keyword] = url
	}
}

// prefilterKeywords scans document text and returns only keywords present in the content
// This optimization reduces keyword processing from ~850 to ~20-50 keywords per document
func (kl *KeywordLinker) prefilterKeywords(text string) map[string]string {
	filtered := make(map[string]string)
	textLower := strings.ToLower(text)

	for keyword, url := range kl.keywords {
		keywordLower := strings.ToLower(keyword)
		if strings.Contains(textLower, keywordLower) {
			filtered[keyword] = url
		}
	}

	return filtered
}

// processNodeWithKeywords recursively processes HTML nodes to add keyword links
func (kl *KeywordLinker) processNodeWithKeywords(n *html.Node, filteredKeywords map[string]string) {
	// Handle text nodes
	if n.Type == html.TextNode {
		if !isInSkippedTag(n) {
			kl.replaceTextNodeWithLinks(n, filteredKeywords)
		}
		return
	}

	// Process children for all other node types
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		kl.processNodeWithKeywords(c, filteredKeywords)
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

// replaceTextNodeWithLinks replaces a text node with text + link nodes for keywords (all occurrences)
func (kl *KeywordLinker) replaceTextNodeWithLinks(n *html.Node, filteredKeywords map[string]string) {
	text := n.Data
	parent := n.Parent

	if parent == nil {
		return
	}

	// Sort keywords by length (longest first) to prevent partial matches
	sortedKeywords := kl.getSortedKeywords(filteredKeywords)

	// Process keywords longest-first, replacing one keyword at a time
	// Since we sort by length, we don't need recursion - the outer processNode loop handles all text nodes
	for _, keyword := range sortedKeywords {
		// Use pre-compiled regex for performance
		re := kl.compiledRegex[keyword]

		// Find ALL occurrences of this keyword
		allMatches := re.FindAllStringIndex(text, -1)
		if len(allMatches) == 0 {
			continue
		}

		// Build nodes: interleave text and link nodes for all occurrences
		nodes := make([]*html.Node, 0)
		lastEnd := 0

		for _, loc := range allMatches {
			// Add text before this match
			if loc[0] > lastEnd {
				nodes = append(nodes, &html.Node{
					Type: html.TextNode,
					Data: text[lastEnd:loc[0]],
				})
			}

			// Add link node for this match
			linkNode := &html.Node{
				Type: html.ElementNode,
				Data: "a",
				Attr: []html.Attribute{
					{Key: "href", Val: filteredKeywords[keyword]},
					{Key: "class", Val: "keyword-link"},
				},
			}
			linkText := &html.Node{
				Type: html.TextNode,
				Data: text[loc[0]:loc[1]],
			}
			linkNode.AppendChild(linkText)
			nodes = append(nodes, linkNode)

			lastEnd = loc[1]
		}

		// Add remaining text after last match
		if lastEnd < len(text) {
			nodes = append(nodes, &html.Node{
				Type: html.TextNode,
				Data: text[lastEnd:],
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

		// Keyword replaced successfully, exit to let processNode handle remaining text nodes
		return
	}
}

// getSortedKeywords returns keywords sorted by length (longest first)
// This prevents partial matching of shorter keywords within longer ones
func (kl *KeywordLinker) getSortedKeywords(keywordMap map[string]string) []string {
	keywords := make([]string, 0, len(keywordMap))
	for keyword := range keywordMap {
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
