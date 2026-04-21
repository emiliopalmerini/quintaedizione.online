package templates

import (
	"bytes"
	"context"
	"fmt"

	"github.com/a-h/templ"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/models"
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

func (e *TemplEngine) RenderHome(ctx context.Context, data models.HomePageData) (string, error) {
	return e.renderComponent(ctx, templComponents.HomePage(data))
}

func (e *TemplEngine) RenderArea(ctx context.Context, data models.AreaPageData) (string, error) {
	return e.renderComponent(ctx, templComponents.AreaPage(data))
}

func (e *TemplEngine) RenderCollection(ctx context.Context, data models.CollectionPageData) (string, error) {
	return e.renderComponent(ctx, templComponents.CollectionPage(data))
}

func (e *TemplEngine) RenderItem(ctx context.Context, data models.ItemPageData) (string, error) {
	return e.renderComponent(ctx, templComponents.ItemPage(data))
}

func (e *TemplEngine) RenderRows(ctx context.Context, data models.CollectionPageData) (string, error) {
	return e.renderComponent(ctx, templComponents.RowsPartial(data))
}

func (e *TemplEngine) RenderError(ctx context.Context, data models.ErrorPageData) (string, error) {
	return e.renderComponent(ctx, templComponents.ErrorPage(data))
}

func (e *TemplEngine) RenderSearch(ctx context.Context, data models.SearchPageData) (string, error) {
	return e.renderComponent(ctx, templComponents.SearchPage(data))
}

func (e *TemplEngine) RenderSearchDropdown(ctx context.Context, results []models.CollectionSearchResult, query string) (string, error) {
	return e.renderComponent(ctx, templComponents.SearchDropdown(results, query))
}

func (e *TemplEngine) RenderSearchBrowse(ctx context.Context, data models.SearchBrowseData) (string, error) {
	return e.renderComponent(ctx, templComponents.SearchBrowseDropdown(data))
}

func (e *TemplEngine) renderComponent(ctx context.Context, component templ.Component) (string, error) {
	var buf bytes.Buffer

	if err := component.Render(ctx, &buf); err != nil {
		return "", fmt.Errorf("failed to render template component: %w", err)
	}

	return buf.String(), nil
}

func (e *TemplEngine) Render(templateName string, data interface{}) (string, error) {
	return "", fmt.Errorf("legacy Render method called with template %s - use specific Render methods instead", templateName)
}
