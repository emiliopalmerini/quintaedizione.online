package parsers

import (
	"log"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type MarkdownRenderer struct {
	extensions    parser.Extensions
	opts          html.RendererOptions
	keywordLinker *KeywordLinker
}

func NewMarkdownRenderer(keywordConfigPath string) *MarkdownRenderer {

	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock

	opts := html.RendererOptions{
		Flags: html.CommonFlags | html.HrefTargetBlank,
	}

	// Initialize keyword linker if config path is provided
	var keywordLinker *KeywordLinker
	if keywordConfigPath != "" {
		var err error
		keywordLinker, err = NewKeywordLinker(keywordConfigPath)
		if err != nil {
			log.Printf("Warning: failed to initialize keyword linker: %v", err)
			keywordLinker = nil
		}
	}

	return &MarkdownRenderer{
		extensions:    extensions,
		opts:          opts,
		keywordLinker: keywordLinker,
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

	// Apply keyword linking if available
	if r.keywordLinker != nil {
		htmlContent = r.keywordLinker.LinkKeywords(htmlContent)
	}

	return strings.TrimSpace(htmlContent)
}

func (r *MarkdownRenderer) RenderLines(lines []string) string {
	content := strings.Join(lines, "\n")
	return r.Render(content)
}
