package parsers

import (
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// MarkdownRenderer converts markdown to HTML
type MarkdownRenderer struct {
	extensions parser.Extensions
	opts       html.RendererOptions
}

// NewMarkdownRenderer creates a new markdown to HTML renderer
func NewMarkdownRenderer() *MarkdownRenderer {
	// Configure parser extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock

	// Configure HTML renderer options
	opts := html.RendererOptions{
		Flags: html.CommonFlags | html.HrefTargetBlank,
	}

	return &MarkdownRenderer{
		extensions: extensions,
		opts:       opts,
	}
}

// Render converts markdown string to HTML
func (r *MarkdownRenderer) Render(markdownContent string) string {
	if markdownContent == "" {
		return ""
	}

	// Create a new parser for each call (gomarkdown requirement)
	p := parser.NewWithExtensions(r.extensions)

	// Parse markdown
	doc := p.Parse([]byte(markdownContent))

	// Render to HTML
	renderer := html.NewRenderer(r.opts)
	htmlBytes := markdown.Render(doc, renderer)

	return strings.TrimSpace(string(htmlBytes))
}

// RenderLines converts markdown lines to HTML
func (r *MarkdownRenderer) RenderLines(lines []string) string {
	content := strings.Join(lines, "\n")
	return r.Render(content)
}
