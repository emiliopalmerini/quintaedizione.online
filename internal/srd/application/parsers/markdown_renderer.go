package parsers

import (
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type MarkdownRenderer struct {
	extensions  parser.Extensions
	opts        html.RendererOptions
	crossLinker *CrossLinker
}

func NewMarkdownRenderer(crossLinker *CrossLinker) *MarkdownRenderer {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock

	opts := html.RendererOptions{
		Flags: html.CommonFlags | html.HrefTargetBlank,
	}

	return &MarkdownRenderer{
		extensions:  extensions,
		opts:        opts,
		crossLinker: crossLinker,
	}
}

func (r *MarkdownRenderer) Render(markdownContent string) string {
	if markdownContent == "" {
		return ""
	}

	p := parser.NewWithExtensions(r.extensions)

	doc := p.Parse([]byte(markdownContent))

	renderer := html.NewRenderer(r.opts)
	htmlBytes := markdown.Render(doc, renderer)

	htmlContent := string(htmlBytes)

	if r.crossLinker != nil {
		htmlContent = r.crossLinker.LinkTerms(htmlContent)
	}

	return strings.TrimSpace(htmlContent)
}

// RenderInline converts markdown to HTML and strips the wrapping <p> tags,
// so the result can be used inline inside other HTML elements.
func (r *MarkdownRenderer) RenderInline(markdown string) string {
	h := r.Render(markdown)
	h = strings.TrimPrefix(h, "<p>")
	h = strings.TrimSuffix(h, "</p>")
	return strings.TrimSpace(h)
}

func (r *MarkdownRenderer) RenderLines(lines []string) string {
	content := strings.Join(lines, "\n")
	return r.Render(content)
}
