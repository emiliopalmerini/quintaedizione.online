package web

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/collections"
	webmappers "github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/mappers"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/models"
	pkgweb "github.com/emiliopalmerini/quintaedizione.online/pkg/web"
)

// CollectionHandler handles collection-related requests.
type CollectionHandler struct {
	*baseHandler
}

// loadCollectionData fetches and assembles the shared data for collection list/rows views.
func (h *CollectionHandler) loadCollectionData(ctx context.Context, collection, queryString string, params pkgweb.PaginationParams, filters map[string]string) (*models.CollectionPageData, error) {
	docs, totalCount, err := h.contentService.GetCollectionItems(ctx, collection, params.Query, filters, params.PageNum, params.PageSize)
	if err != nil {
		return nil, err
	}

	documents := h.documentMapper.ToModels(collection, docs)
	pagination := pkgweb.CalculatePaginationData(params.PageNum, params.PageSize, totalCount)

	facetCounts, _ := h.contentService.GetFacetCounts(ctx, collection, params.Query, filters)
	filterOptions := h.buildFilterOptionsWithCounts(collection, filters, facetCounts)

	return &models.CollectionPageData{
		PageData: models.PageData{
			Collection:  collection,
			QueryString: queryString,
		},
		Documents:   documents,
		Filters:     filterOptions,
		QuickFilter: buildQuickFilterData(collection, filters),
		Query:       params.Query,
		Page:        params.PageNum,
		PageSize:    params.PageSize,
		Total:       totalCount,
		TotalPages:  pagination.TotalPages,
		HasNext:     pagination.HasNext,
		HasPrev:     pagination.HasPrev,
		StartItem:   pagination.StartItem,
		EndItem:     pagination.EndItem,
	}, nil
}

// handleCollectionList renders the full collection page with filters and pagination.
func (h *CollectionHandler) handleCollectionList(w http.ResponseWriter, r *http.Request) {
	collection := r.PathValue("collection")
	params := pkgweb.ExtractPaginationParams(r)
	filters := h.extractFilters(r)

	data, err := h.loadCollectionData(r.Context(), collection, r.URL.RawQuery, params, filters)
	if err != nil {
		h.ErrorResponse(w, r, err, fmt.Sprintf("Errore nel caricamento della collezione %s", collection))
		return
	}

	collectionTitle := h.getCollectionTitle(collection)
	data.PageData.Title = collectionTitle
	data.PageData.Description = fmt.Sprintf("Elenco completo di %s — SRD 5e in italiano.", collectionTitle)

	content, err := h.templateEngine.RenderCollection(r.Context(), *data)
	if err != nil {
		h.ErrorResponse(w, r, err, "Errore nel rendering della pagina collezione")
		return
	}

	h.renderHTML(w, content, "collection")
}

// handleCollectionRows renders only the table rows for HTMX partial updates.
func (h *CollectionHandler) handleCollectionRows(w http.ResponseWriter, r *http.Request) {
	collection := r.PathValue("collection")
	params := pkgweb.ExtractPaginationParams(r)
	filters := h.extractFilters(r)

	data, err := h.loadCollectionData(r.Context(), collection, r.URL.RawQuery, params, filters)
	if err != nil {
		h.ErrorResponse(w, r, err, fmt.Sprintf("Errore nel caricamento righe per %s", collection))
		return
	}

	content, err := h.templateEngine.RenderRows(r.Context(), *data)
	if err != nil {
		h.ErrorResponse(w, r, err, "Errore nel rendering delle righe")
		return
	}

	h.renderHTML(w, content, "collection")
}

// handleItemDetail renders the detail page for a single item.
func (h *CollectionHandler) handleItemDetail(w http.ResponseWriter, r *http.Request) {
	collection := r.PathValue("collection")
	source := r.PathValue("source")
	slug := r.PathValue("slug")

	// Composite key: source/slug
	itemID := source + "/" + slug
	doc, err := h.contentService.GetItem(r.Context(), collection, itemID)
	if err != nil {
		h.ErrorResponse(w, r, err, "Elemento non trovato")
		return
	}

	prevSlug, nextSlug, position, total, err := h.contentService.GetAdjacentItems(r.Context(), collection, itemID)
	if err != nil {
		fmt.Printf("Warning: Could not get adjacent items for %s/%s: %v\n", collection, slug, err)
	}

	prevID := ""
	if prevSlug != nil {
		prevID = *prevSlug
	}
	nextID := ""
	if nextSlug != nil {
		nextID = *nextSlug
	}

	// Look up all editions of this document for the version switcher
	var versionTabs []models.VersionTab
	if versions, err := h.contentService.GetItemVersions(r.Context(), collection, slug); err != nil {
		fmt.Printf("Warning: Could not get item versions for %s/%s: %v\n", collection, slug, err)
	} else if len(versions) > 1 {
		versionTabs = make([]models.VersionTab, len(versions))
		for i, v := range versions {
			versionTabs[i] = models.VersionTab{
				SourceShort: v.SourceShort,
				URL:         "/srd/" + collection + "/" + v.SourceShort + "/" + slug,
				IsCurrent:   v.SourceShort == doc.Source,
				Label:       v.SourceShort,
			}
		}
	}

	data := models.ItemPageData{
		PageData: models.PageData{
			Title:       doc.Title,
			Description: truncateDescription(doc.RawContent.String(), 160),
			DocTitle:    doc.Title,
			DocID:       slug,
			Collection:  collection,
			QueryString: r.URL.RawQuery,
		},
		BodyRaw:         doc.RawContent.String(),
		PrevID:          prevID,
		NextID:          nextID,
		Position:        position,
		Total:           total,
		CollectionLabel: h.getCollectionTitle(collection),
		SourceShort:     doc.Source,
		Versions:        versionTabs,
	}

	// Build stat-block view model or fall back to BodyHTML
	webmappers.BuildStatBlockData(doc, &data)

	content, err := h.templateEngine.RenderItem(r.Context(), data)
	if err != nil {
		h.ErrorResponse(w, r, err, "Errore nel rendering della pagina elemento")
		return
	}

	h.renderHTML(w, content, "item")
}

// buildFilterOptionsWithCounts returns filter options with optional facet counts.
func (h *CollectionHandler) buildFilterOptionsWithCounts(collection string, activeFilters map[string]string, facetCounts map[string]map[string]int64) []models.FilterOption {
	defs, err := h.contentService.GetAvailableFilters(collection)
	if err != nil || len(defs) == 0 {
		return nil
	}

	options := make([]models.FilterOption, 0, len(defs))
	for _, def := range defs {
		if len(def.EnumValues) == 0 {
			continue
		}
		counts := facetCounts[def.Name]
		values := make([]models.FilterValueOption, 0, len(def.EnumValues))
		for _, v := range def.EnumValues {
			var count int64
			if counts != nil {
				count = counts[v]
			}
			values = append(values, models.FilterValueOption{
				Value: v,
				Count: count,
			})
		}
		currentValue := activeFilters[def.Name]
		var currentValues []string
		if currentValue != "" {
			currentValues = strings.Split(currentValue, ",")
		}
		options = append(options, models.FilterOption{
			Name:          def.Name,
			Label:         def.Description,
			Values:        values,
			CurrentValue:  currentValue,
			CurrentValues: currentValues,
		})
	}
	return options
}

// buildQuickFilterData returns the quick filter chip data for a collection,
// with active state computed from the current filter values.
func buildQuickFilterData(collection string, activeFilters map[string]string) *models.QuickFilterData {
	cn, ok := collections.FromString(collection)
	if !ok {
		return nil
	}
	qf, ok := collections.GetQuickFilter(cn)
	if !ok {
		return nil
	}

	currentValue := activeFilters[qf.FilterName]
	var activeValues map[string]bool
	if currentValue != "" {
		activeValues = make(map[string]bool)
		for _, v := range strings.Split(currentValue, ",") {
			activeValues[v] = true
		}
	}

	chips := make([]models.QuickFilterChip, len(qf.Chips))
	for i, chip := range qf.Chips {
		active := activeValues != nil && chipMatchesFilter(chip.Values, activeValues)
		chips[i] = models.QuickFilterChip{
			Label:  chip.Label,
			Values: chip.Values,
			Active: active,
		}
	}

	return &models.QuickFilterData{
		FilterName: qf.FilterName,
		Chips:      chips,
	}
}

// chipMatchesFilter returns true if all chip values are in the active filter set.
func chipMatchesFilter(chipValues []string, activeValues map[string]bool) bool {
	for _, v := range chipValues {
		if !activeValues[v] {
			return false
		}
	}
	return true
}

// extractFilters extracts filter parameters from the query string.
func (h *CollectionHandler) extractFilters(r *http.Request) map[string]string {
	filters := make(map[string]string)

	skipParams := map[string]bool{
		"page":      true,
		"page_size": true,
		"q":         true,
	}

	for param, values := range r.URL.Query() {
		if !skipParams[param] && len(values) > 0 && values[0] != "" {
			filters[param] = values[0]
		}
	}

	return filters
}
