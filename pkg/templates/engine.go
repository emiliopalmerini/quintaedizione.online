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
	isDev        bool
}

// NewEngine creates a new template engine
func NewEngine(templatesDir string) *Engine {
	return &Engine{
		templatesDir: templatesDir,
		isDev:        false,
	}
}

// NewDevEngine creates a new template engine with development features
func NewDevEngine(templatesDir string) *Engine {
	return &Engine{
		templatesDir: templatesDir,
		isDev:        true,
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

	// In development mode, reload templates for each request
	if e.isDev {
		if err := e.LoadTemplates(); err != nil {
			return "", fmt.Errorf("failed to reload templates in dev mode: %w", err)
		}
	}

	// Use cached templates if available
	if e.templates != nil {
		if err := e.templates.ExecuteTemplate(&buf, templateName, data); err != nil {
			return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
		}
		return buf.String(), nil
	}

	// Fallback to loading templates if not cached
	if err := e.LoadTemplates(); err != nil {
		return "", fmt.Errorf("failed to load templates: %w", err)
	}

	if err := e.templates.ExecuteTemplate(&buf, templateName, data); err != nil {
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
