package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
)

type Engine struct {
	templates    *template.Template
	templatesDir string
	isDev        bool
}

func NewEngine(templatesDir string) *Engine {
	return &Engine{
		templatesDir: templatesDir,
		isDev:        false,
	}
}

func NewDevEngine(templatesDir string) *Engine {
	return &Engine{
		templatesDir: templatesDir,
		isDev:        true,
	}
}

func (e *Engine) LoadTemplates() error {

	funcMap := template.FuncMap{
		"add":     add,
		"sub":     sub,
		"eq":      eq,
		"default": defaultValue,
		"safe":    safe,
	}

	basePath := filepath.Join(e.templatesDir, "base.html")
	tmpl, err := template.New("base.html").Funcs(funcMap).ParseFiles(basePath)
	if err != nil {
		return fmt.Errorf("failed to parse base template: %w", err)
	}

	pattern := filepath.Join(e.templatesDir, "*.html")
	tmpl, err = tmpl.ParseGlob(pattern)
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	e.templates = tmpl
	return nil
}

func (e *Engine) Render(templateName string, data interface{}) (string, error) {
	var buf bytes.Buffer

	if e.isDev {
		if err := e.LoadTemplates(); err != nil {
			return "", fmt.Errorf("failed to reload templates in dev mode: %w", err)
		}
	}

	if e.templates != nil {
		if err := e.templates.ExecuteTemplate(&buf, templateName, data); err != nil {
			return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
		}
		return buf.String(), nil
	}

	if err := e.LoadTemplates(); err != nil {
		return "", fmt.Errorf("failed to load templates: %w", err)
	}

	if err := e.templates.ExecuteTemplate(&buf, templateName, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	return buf.String(), nil
}

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
