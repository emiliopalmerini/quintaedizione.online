package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
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
	// Create function map with essential helper functions
	funcMap := template.FuncMap{
		"add":     add,
		"sub":     sub,
		"eq":      eq,
		"default": defaultValue,
		"safe":    safe,
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
		"add":     add,
		"sub":     sub,
		"eq":      eq,
		"default": defaultValue,
		"safe":    safe,
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

	// Parse common partial templates that might be included
	partialTemplates := []string{
		"rows.html",
	}

	for _, partial := range partialTemplates {
		partialPath := filepath.Join(e.templatesDir, partial)
		tmpl, err = tmpl.ParseFiles(partialPath)
		if err != nil {
			// Partial templates are optional, don't fail if missing
			fmt.Printf("Warning: Could not load partial template %s: %v\n", partial, err)
		}
	}

	// Execute base template which will call the content blocks from the specific template
	if err := tmpl.ExecuteTemplate(&buf, "base.html", data); err != nil {
		fmt.Printf("Template execution error for %s: %v\n", templateName, err)
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	return buf.String(), nil
}

// Template helper functions

func add(a, b int) int         { return a + b }
func sub(a, b int) int         { return a - b }
func eq(a, b interface{}) bool { return a == b }

func defaultValue(value, defaultVal interface{}) interface{} {
	if value == nil || value == "" {
		return defaultVal
	}
	return value
}

func safe(s string) template.HTML {
	return template.HTML(s)
}
