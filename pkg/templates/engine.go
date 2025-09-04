package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
)

// Engine handles template rendering
type Engine struct {
	templates    *template.Template
	templatesDir string
}

// NewEngine creates a new template engine
func NewEngine(templatesDir string) *Engine {
	return &Engine{
		templatesDir: templatesDir,
	}
}

// LoadTemplates loads all templates from the templates directory
func (e *Engine) LoadTemplates() error {
	// Create function map with helper functions
	funcMap := template.FuncMap{
		"add":        add,
		"sub":        sub,
		"mul":        mul,
		"div":        div,
		"mod":        mod,
		"eq":         eq,
		"ne":         ne,
		"lt":         lt,
		"le":         le,
		"gt":         gt,
		"ge":         ge,
		"and":        and,
		"or":         or,
		"not":        not,
		"len":        length,
		"capitalize": capitalize,
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"trim":       strings.TrimSpace,
		"replace":    strings.ReplaceAll,
		"contains":   strings.Contains,
		"hasPrefix":  strings.HasPrefix,
		"hasSuffix":  strings.HasSuffix,
		"split":      strings.Split,
		"join":       strings.Join,
		"default":    defaultValue,
		"safe":       safe,
		"markdown":   renderMarkdown,
		"slice":      slice,
	}

	// Parse base template first
	basePath := filepath.Join(e.templatesDir, "base.html")
	tmpl, err := template.New("base.html").Funcs(funcMap).ParseFiles(basePath)
	if err != nil {
		return fmt.Errorf("failed to parse base template: %w", err)
	}

	// Parse all other templates
	pattern := filepath.Join(e.templatesDir, "*.html")
	tmpl, err = tmpl.ParseGlob(pattern)
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}


	e.templates = tmpl
	return nil
}

// Render renders a template with the given data
func (e *Engine) Render(templateName string, data interface{}) (string, error) {
	var buf bytes.Buffer
	
	// Create a new template set for this specific render to avoid block conflicts
	tmpl := template.New("base").Funcs(template.FuncMap{
		"add":        add,
		"sub":        sub,
		"mul":        mul,
		"div":        div,
		"mod":        mod,
		"eq":         eq,
		"ne":         ne,
		"lt":         lt,
		"le":         le,
		"gt":         gt,
		"ge":         ge,
		"and":        and,
		"or":         or,
		"not":        not,
		"len":        length,
		"capitalize": capitalize,
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"trim":       strings.TrimSpace,
		"replace":    strings.ReplaceAll,
		"contains":   strings.Contains,
		"hasPrefix":  strings.HasPrefix,
		"hasSuffix":  strings.HasSuffix,
		"slice":      slice,
		"safe":       safe,
		"default":    defaultValue,
	})
	
	// Parse base template
	basePattern := filepath.Join(e.templatesDir, "base.html")
	tmpl, err := tmpl.ParseFiles(basePattern)
	if err != nil {
		return "", fmt.Errorf("failed to parse base template: %w", err)
	}
	
	// Parse the specific template
	specificPattern := filepath.Join(e.templatesDir, templateName)
	tmpl, err = tmpl.ParseFiles(specificPattern)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", templateName, err)
	}
	
	// Execute base template which will call the content blocks from the specific template
	if err := tmpl.ExecuteTemplate(&buf, "base.html", data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	return buf.String(), nil
}

// Template helper functions

func add(a, b int) int       { return a + b }
func sub(a, b int) int       { return a - b }
func mul(a, b int) int       { return a * b }
func div(a, b int) int       { return a / b }
func mod(a, b int) int       { return a % b }
func eq(a, b interface{}) bool  { return a == b }
func ne(a, b interface{}) bool  { return a != b }
func lt(a, b int) bool       { return a < b }
func le(a, b int) bool       { return a <= b }
func gt(a, b int) bool       { return a > b }
func ge(a, b int) bool       { return a >= b }
func and(a, b bool) bool     { return a && b }
func or(a, b bool) bool      { return a || b }
func not(a bool) bool        { return !a }

func length(v interface{}) int {
	switch s := v.(type) {
	case string:
		return len(s)
	case []interface{}:
		return len(s)
	case []string:
		return len(s)
	case []int:
		return len(s)
	case map[string]interface{}:
		return len(s)
	default:
		return 0
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

func defaultValue(value, defaultVal interface{}) interface{} {
	if value == nil || value == "" {
		return defaultVal
	}
	return value
}

func safe(s string) template.HTML {
	return template.HTML(s)
}

func renderMarkdown(content string) template.HTML {
	// Simple markdown rendering - in production you'd use a proper markdown library
	html := strings.ReplaceAll(content, "\n", "<br>")
	html = strings.ReplaceAll(html, "**", "<strong>")
	html = strings.ReplaceAll(html, "*", "<em>")
	return template.HTML(html)
}

func slice(args ...interface{}) interface{} {
	if len(args) < 1 {
		return nil
	}

	// If first arg is string and we have start/end args, do string slicing
	if str, ok := args[0].(string); ok && len(args) == 3 {
		start, startOk := args[1].(int)
		end, endOk := args[2].(int)
		
		if startOk && endOk {
			if start < 0 {
				start = 0
			}
			if end > len(str) {
				end = len(str)
			}
			if start > end {
				return ""
			}
			return str[start:end]
		}
	}

	// Otherwise, create a slice from the arguments
	result := make([]interface{}, len(args))
	for i, arg := range args {
		result[i] = arg
	}
	return result
}