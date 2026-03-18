package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/models"
	"github.com/emiliopalmerini/quintaedizione.online/pkg/mappers"
)

// CollectionHandler handles collection-related requests.
type CollectionHandler struct {
	*baseHandler
}

// handleCollectionList renders the full collection page with filters and pagination.
func (h *CollectionHandler) handleCollectionList(w http.ResponseWriter, r *http.Request) {
	collection := r.PathValue("collection")
	params := ExtractPaginationParams(r)

	filters := h.extractFilters(r)

	rawItems, totalCount, err := h.contentService.GetCollectionItems(r.Context(), collection, params.Query, filters, params.PageNum, params.PageSize)
	if err != nil {
		h.ErrorResponse(w, r, err, fmt.Sprintf("Errore nel caricamento della collezione %s", collection))
		return
	}

	documents := h.documentMapper.ToModels(collection, rawItems)

	pagination := CalculatePaginationData(params.PageNum, params.PageSize, totalCount)

	facetCounts, _ := h.contentService.GetFacetCounts(r.Context(), collection, params.Query, filters)
	filterOptions := h.buildFilterOptionsWithCounts(collection, filters, facetCounts)

	collectionTitle := h.getCollectionTitle(collection)
	data := models.CollectionPageData{
		PageData: models.PageData{
			Title:       collectionTitle,
			Description: fmt.Sprintf("Elenco completo di %s — SRD 5e in italiano.", collectionTitle),
			Collection:  collection,
			QueryString: r.URL.RawQuery,
		},
		Documents:  documents,
		Filters:    filterOptions,
		Query:      params.Query,
		Page:       params.PageNum,
		PageSize:   params.PageSize,
		Total:      totalCount,
		TotalPages: pagination.TotalPages,
		HasNext:    pagination.HasNext,
		HasPrev:    pagination.HasPrev,
		StartItem:  pagination.StartItem,
		EndItem:    pagination.EndItem,
	}

	content, err := h.templateEngine.RenderCollection(data)
	if err != nil {
		h.ErrorResponse(w, r, err, "Errore nel rendering della pagina collezione")
		return
	}

	h.setCacheHeaders(w, "collection")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(content))
}

// handleCollectionRows renders only the table rows for HTMX partial updates.
func (h *CollectionHandler) handleCollectionRows(w http.ResponseWriter, r *http.Request) {
	collection := r.PathValue("collection")
	params := ExtractPaginationParams(r)

	filters := h.extractFilters(r)

	rawItems, totalCount, err := h.contentService.GetCollectionItems(r.Context(), collection, params.Query, filters, params.PageNum, params.PageSize)
	if err != nil {
		h.ErrorResponse(w, r, err, fmt.Sprintf("Errore nel caricamento righe per %s", collection))
		return
	}

	documents := h.documentMapper.ToModels(collection, rawItems)

	pagination := CalculatePaginationData(params.PageNum, params.PageSize, totalCount)

	facetCounts, _ := h.contentService.GetFacetCounts(r.Context(), collection, params.Query, filters)
	filterOptions := h.buildFilterOptionsWithCounts(collection, filters, facetCounts)

	data := models.CollectionPageData{
		PageData: models.PageData{
			Collection:  collection,
			QueryString: r.URL.RawQuery,
		},
		Documents:  documents,
		Filters:    filterOptions,
		Query:      params.Query,
		Page:       params.PageNum,
		PageSize:   params.PageSize,
		Total:      totalCount,
		TotalPages: pagination.TotalPages,
		HasNext:    pagination.HasNext,
		HasPrev:    pagination.HasPrev,
		StartItem:  pagination.StartItem,
		EndItem:    pagination.EndItem,
	}

	content, err := h.templateEngine.RenderRows(data)
	if err != nil {
		h.ErrorResponse(w, r, err, "Errore nel rendering delle righe")
		return
	}

	h.setCacheHeaders(w, "collection")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(content))
}

// handleItemDetail renders the detail page for a single item.
func (h *CollectionHandler) handleItemDetail(w http.ResponseWriter, r *http.Request) {
	collection := r.PathValue("collection")
	source := r.PathValue("source")
	slug := r.PathValue("slug")

	// Composite key: source/slug
	itemID := source + "/" + slug
	item, err := h.contentService.GetItem(r.Context(), collection, itemID)
	if err != nil {
		h.ErrorResponse(w, r, err, "Elemento non trovato")
		return
	}

	bodyHTML := mappers.GetString(item, "content", "")
	bodyRaw := mappers.GetString(item, "raw_content", "")

	prevSlug, nextSlug, position, total, err := h.contentService.GetAdjacentItems(r.Context(), collection, itemID)
	if err != nil {
		fmt.Printf("Warning: Could not get adjacent items for %s/%s: %v\n", collection, slug, err)
	}

	docTitle := mappers.GetString(item, "title", "")

	prevID := ""
	if prevSlug != nil {
		prevID = *prevSlug
	}
	nextID := ""
	if nextSlug != nil {
		nextID = *nextSlug
	}

	data := models.ItemPageData{
		PageData: models.PageData{
			Title:       docTitle,
			Description: truncateDescription(bodyRaw, 160),
			DocTitle:    docTitle,
			DocID:       slug,
			Collection:  collection,
			QueryString: r.URL.RawQuery,
		},
		BodyRaw:         bodyRaw,
		BodyHTML:        bodyHTML,
		PrevID:          prevID,
		NextID:          nextID,
		Position:        position,
		Total:           total,
		CollectionLabel: h.getCollectionTitle(collection),
	}

	content, err := h.templateEngine.RenderItem(data)
	if err != nil {
		h.ErrorResponse(w, r, err, "Errore nel rendering della pagina elemento")
		return
	}

	h.setCacheHeaders(w, "item")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(content))
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
