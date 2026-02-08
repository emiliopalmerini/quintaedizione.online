package web

import (
	"fmt"
	"net/http"

	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/models"
	"github.com/emiliopalmerini/quintaedizione.online/pkg/mappers"
	"github.com/gin-gonic/gin"
)

// CollectionHandler handles collection-related requests.
type CollectionHandler struct {
	*baseHandler
}

// handleCollectionList renders the full collection page with filters and pagination.
func (h *CollectionHandler) handleCollectionList(c *gin.Context) {
	collection := c.Param("collection")
	params := ExtractPaginationParams(c)

	filters := h.extractFilters(c)

	rawItems, totalCount, err := h.contentService.GetCollectionItems(c.Request.Context(), collection, params.Query, filters, params.PageNum, params.PageSize)
	if err != nil {
		h.ErrorResponse(c, err, fmt.Sprintf("Errore nel caricamento della collezione %s", collection))
		return
	}

	documents := h.documentMapper.ToModels(collection, rawItems)

	pagination := CalculatePaginationData(params.PageNum, params.PageSize, totalCount)

	data := models.CollectionPageData{
		PageData: models.PageData{
			Title:       h.getCollectionTitle(collection),
			Collection:  collection,
			QueryString: c.Request.URL.RawQuery,
		},
		Documents:  documents,
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
		h.ErrorResponse(c, err, "Errore nel rendering della pagina collezione")
		return
	}

	h.setCacheHeaders(c, "collection")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

// handleCollectionRows renders only the table rows for HTMX partial updates.
func (h *CollectionHandler) handleCollectionRows(c *gin.Context) {
	collection := c.Param("collection")
	params := ExtractPaginationParams(c)

	filters := h.extractFilters(c)

	rawItems, totalCount, err := h.contentService.GetCollectionItems(c.Request.Context(), collection, params.Query, filters, params.PageNum, params.PageSize)
	if err != nil {
		h.ErrorResponse(c, err, fmt.Sprintf("Errore nel caricamento righe per %s", collection))
		return
	}

	documents := h.documentMapper.ToModels(collection, rawItems)

	pagination := CalculatePaginationData(params.PageNum, params.PageSize, totalCount)

	data := models.CollectionPageData{
		PageData: models.PageData{
			Collection:  collection,
			QueryString: c.Request.URL.RawQuery,
		},
		Documents:  documents,
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
		h.ErrorResponse(c, err, "Errore nel rendering delle righe")
		return
	}

	h.setCacheHeaders(c, "collection")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

// handleItemDetail renders the detail page for a single item.
func (h *CollectionHandler) handleItemDetail(c *gin.Context) {
	collection := c.Param("collection")
	slug := c.Param("slug")

	item, err := h.contentService.GetItem(c.Request.Context(), collection, slug)
	if err != nil {
		h.ErrorResponse(c, err, "Elemento non trovato")
		return
	}

	bodyHTML := mappers.GetString(item, "content", "")
	bodyRaw := mappers.GetString(item, "raw_content", "")

	prevSlug, nextSlug, err := h.contentService.GetAdjacentItems(c.Request.Context(), collection, slug)
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
			DocTitle:    docTitle,
			DocID:       slug,
			Collection:  collection,
			QueryString: c.Request.URL.RawQuery,
		},
		BodyRaw:         bodyRaw,
		BodyHTML:        bodyHTML,
		PrevID:          prevID,
		NextID:          nextID,
		CollectionLabel: h.getCollectionTitle(collection),
	}

	content, err := h.templateEngine.RenderItem(data)
	if err != nil {
		h.ErrorResponse(c, err, "Errore nel rendering della pagina elemento")
		return
	}

	h.setCacheHeaders(c, "item")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

// extractFilters extracts filter parameters from the query string.
func (h *CollectionHandler) extractFilters(c *gin.Context) map[string]string {
	filters := make(map[string]string)

	skipParams := map[string]bool{
		"page":      true,
		"page_size": true,
		"q":         true,
	}

	for param, values := range c.Request.URL.Query() {
		if !skipParams[param] && len(values) > 0 && values[0] != "" {
			filters[param] = values[0]
		}
	}

	return filters
}
