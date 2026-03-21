package parsers

import (
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type MarkdownRenderer struct {
	extensions     parser.Extensions
	opts           html.RendererOptions
	glossaryLinker *GlossaryLinker
}

func NewMarkdownRenderer(glossaryLinker *GlossaryLinker) *MarkdownRenderer {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock

	opts := html.RendererOptions{
		Flags: html.CommonFlags | html.HrefTargetBlank,
	}

	return &MarkdownRenderer{
		extensions:     extensions,
		opts:           opts,
		glossaryLinker: glossaryLinker,
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

	if r.glossaryLinker != nil {
		htmlContent = r.glossaryLinker.LinkGlossaryTerms(htmlContent)
	}

	return strings.TrimSpace(htmlContent)
}

func (r *MarkdownRenderer) RenderLines(lines []string) string {
	content := strings.Join(lines, "\n")
	return r.Render(content)
}
