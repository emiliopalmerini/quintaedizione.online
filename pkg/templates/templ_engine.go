package templates

import (
	"bytes"
	"context"
	"fmt"

	"github.com/a-h/templ"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/web/models"
	templComponents "github.com/emiliopalmerini/due-draghi-5e-srd/web/templates"
)

// TemplEngine handles Templ-based template rendering
type TemplEngine struct {
	isDev bool
}

// NewTemplEngine creates a new Templ-based template engine
func NewTemplEngine() *TemplEngine {
	return &TemplEngine{
		isDev: false,
	}
}

// NewDevTemplEngine creates a new Templ-based template engine with development features
func NewDevTemplEngine() *TemplEngine {
	return &TemplEngine{
		isDev: true,
	}
}

// RenderHome renders the home page
func (e *TemplEngine) RenderHome(data models.HomePageData) (string, error) {
	return e.renderComponent(templComponents.HomePage(data))
}

// RenderCollection renders a collection page
func (e *TemplEngine) RenderCollection(data models.CollectionPageData) (string, error) {
	return e.renderComponent(templComponents.CollectionPage(data))
}

// RenderItem renders an item page
func (e *TemplEngine) RenderItem(data models.ItemPageData) (string, error) {
	return e.renderComponent(templComponents.ItemPage(data))
}

// RenderRows renders rows partial (for HTMX updates)
func (e *TemplEngine) RenderRows(data models.CollectionPageData) (string, error) {
	return e.renderComponent(templComponents.RowsPartial(data))
}

// RenderError renders an error page
func (e *TemplEngine) RenderError(data models.ErrorPageData) (string, error) {
	return e.renderComponent(templComponents.ErrorPage(data))
}

// renderComponent is a helper that renders any Templ component to string
func (e *TemplEngine) renderComponent(component templ.Component) (string, error) {
	var buf bytes.Buffer
	ctx := context.Background()

	if err := component.Render(ctx, &buf); err != nil {
		return "", fmt.Errorf("failed to render template component: %w", err)
	}

	return buf.String(), nil
}

// Legacy method for backwards compatibility during transition
// This will be removed once all handlers are updated
func (e *TemplEngine) Render(templateName string, data interface{}) (string, error) {
	return "", fmt.Errorf("legacy Render method called with template %s - use specific Render methods instead", templateName)
}
