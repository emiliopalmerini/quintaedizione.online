package templates

import (
	"bytes"
	"context"
	"fmt"

	"github.com/a-h/templ"
	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/models"
	templComponents "github.com/emiliopalmerini/quintaedizione.online/web/templates"
)

type TemplEngine struct {
	isDev bool
}

func NewTemplEngine() *TemplEngine {
	return &TemplEngine{
		isDev: false,
	}
}

func NewDevTemplEngine() *TemplEngine {
	return &TemplEngine{
		isDev: true,
	}
}

func (e *TemplEngine) RenderHome(data models.HomePageData) (string, error) {
	return e.renderComponent(templComponents.HomePage(data))
}

func (e *TemplEngine) RenderCollection(data models.CollectionPageData) (string, error) {
	return e.renderComponent(templComponents.CollectionPage(data))
}

func (e *TemplEngine) RenderItem(data models.ItemPageData) (string, error) {
	return e.renderComponent(templComponents.ItemPage(data))
}

func (e *TemplEngine) RenderRows(data models.CollectionPageData) (string, error) {
	return e.renderComponent(templComponents.RowsPartial(data))
}

func (e *TemplEngine) RenderError(data models.ErrorPageData) (string, error) {
	return e.renderComponent(templComponents.ErrorPage(data))
}

func (e *TemplEngine) RenderSearch(data models.SearchPageData) (string, error) {
	return e.renderComponent(templComponents.SearchPage(data))
}

func (e *TemplEngine) RenderSearchDropdown(results []models.CollectionSearchResult, query string) (string, error) {
	return e.renderComponent(templComponents.SearchDropdown(results, query))
}

func (e *TemplEngine) renderComponent(component templ.Component) (string, error) {
	var buf bytes.Buffer
	ctx := context.Background()

	if err := component.Render(ctx, &buf); err != nil {
		return "", fmt.Errorf("failed to render template component: %w", err)
	}

	return buf.String(), nil
}

func (e *TemplEngine) Render(templateName string, data interface{}) (string, error) {
	return "", fmt.Errorf("legacy Render method called with template %s - use specific Render methods instead", templateName)
}
