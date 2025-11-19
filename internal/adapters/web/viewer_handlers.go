package web

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/web/config"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/web/display"
	webmappers "github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/web/mappers"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/web/models"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/services"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain/collections"
	infraconfig "github.com/emiliopalmerini/due-draghi-5e-srd/internal/infrastructure/config"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mappers"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/templates"
	"github.com/gin-gonic/gin"
)

// Handlers contains web handlers for the viewer
type Handlers struct {
	contentService     *services.ContentService
	templateEngine     *templates.TemplEngine
	documentMapper     webmappers.DocumentMapper
	collectionMetadata infraconfig.CollectionMetadata
}

// NewHandlers creates a new Handlers instance
func NewHandlers(contentService *services.ContentService, templateEngine *templates.TemplEngine) *Handlers {
	displayFactory := display.NewDisplayElementFactory()
	documentMapper := webmappers.NewDocumentMapper(displayFactory)

	// Load collection metadata - fallback to hardcoded titles if config fails
	collectionMetadata, err := infraconfig.NewCollectionMetadata()
	if err != nil {
		// Log error and continue with nil metadata (will fallback to hardcoded)
		fmt.Printf("Warning: Failed to load collection metadata: %v\n", err)
	}

	return &Handlers{
		contentService:     contentService,
		templateEngine:     templateEngine,
		documentMapper:     documentMapper,
		collectionMetadata: collectionMetadata,
	}
}

// RegisterRoutes registers all web routes
func (h *Handlers) RegisterRoutes(router *gin.Engine) {
	// Main page
	router.GET("/", h.handleHome)

	// Global search
	router.GET("/search", h.handleGlobalSearch)
	router.GET("/search/dropdown", h.handleSearchDropdown)

	// Collection pages
	router.GET("/:collection", h.handleCollectionList)
	router.GET("/:collection/rows", h.handleCollectionRows) // HTMX rows endpoint
	router.GET("/:collection/:slug", h.handleItemDetail)

	// Quick search for breadcrumb
	router.GET("/quicksearch/:collection", h.handleQuickSearch)
}

// handleHome renders the home page
func (h *Handlers) handleHome(c *gin.Context) {
	// Get collection stats for display
	collections, err := h.contentService.GetCollectionStats(c.Request.Context())
	if err != nil {
		// If error, use default collections from registry without counts
		collections = h.getDefaultCollections()
	}

	// Convert collections to typed format and calculate total
	typedCollections := make([]models.Collection, 0, len(collections))
	total := int64(0)

	for _, col := range collections {
		name := mappers.GetString(col, "collection", "")
		count := mappers.GetInt64(col, "count", 0)

		collection := models.Collection{
			Name:  name,
			Count: count,
		}

		if count > 0 {
			total += count
		}

		// Get Italian label using helper method
		collection.Label = h.getCollectionTitle(name)

		typedCollections = append(typedCollections, collection)
	}

	data := models.HomePageData{
		PageData: models.PageData{
			Title:       "5e SRD 2024",
			Description: "Il Fantastico Visualizzatore di SRD (5e 2024)",
		},
		Collections: typedCollections,
		Total:       total,
	}

	content, err := h.templateEngine.RenderHome(data)
	if err != nil {
		h.ErrorResponse(c, err, "Errore nel rendering della pagina home")
		return
	}

	h.setCacheHeaders(c, "home")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

// handleCollectionList renders a collection list page
func (h *Handlers) handleCollectionList(c *gin.Context) {
	collection := c.Param("collection")
	page := c.DefaultQuery("page", "1")
	q := c.Query("q")
	pageSize := c.DefaultQuery("page_size", "20")

	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	pageSizeNum, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeNum < 1 || pageSizeNum > 100 {
		pageSizeNum = 20
	}

	// Extract filter parameters
	filters := h.extractFilters(c)

	// Get items from service (filters handled internally)
	rawItems, totalCount, err := h.contentService.GetCollectionItems(c.Request.Context(), collection, q, filters, pageNum, pageSizeNum)
	if err != nil {
		h.ErrorResponse(c, err, fmt.Sprintf("Errore nel caricamento della collezione %s", collection))
		return
	}

	// Convert to typed documents using mapper
	documents := h.documentMapper.ToModels(collection, rawItems)

	// Calculate pagination
	pagination := CalculatePaginationData(pageNum, pageSizeNum, totalCount)

	data := models.CollectionPageData{
		PageData: models.PageData{
			Title:       h.getCollectionTitle(collection),
			Collection:  collection,
			QueryString: c.Request.URL.RawQuery,
		},
		Documents:  documents,
		Query:      q,
		Page:       pageNum,
		PageSize:   pageSizeNum,
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

// handleItemDetail renders an item detail page
func (h *Handlers) handleItemDetail(c *gin.Context) {
	collection := c.Param("collection")
	slug := c.Param("slug")

	item, err := h.contentService.GetItem(c.Request.Context(), collection, slug)
	if err != nil {
		h.ErrorResponse(c, err, "Elemento non trovato")
		return
	}

	// Extract pre-rendered HTML content and raw markdown
	bodyHTML := mappers.GetString(item, "content", "")
	bodyRaw := mappers.GetString(item, "raw_content", "")

	// Get navigation items
	prevSlug, nextSlug, err := h.contentService.GetAdjacentItems(c.Request.Context(), collection, slug)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: Could not get adjacent items for %s/%s: %v\n", collection, slug, err)
	}

	// Get doc title from root level
	docTitle := mappers.GetString(item, "title", "")

	// Handle navigation slugs (they're pointers)
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

// handleCollectionRows handles HTMX requests for collection rows
func (h *Handlers) handleCollectionRows(c *gin.Context) {
	collection := c.Param("collection")
	page := c.DefaultQuery("page", "1")
	q := c.Query("q")
	pageSize := c.DefaultQuery("page_size", "20")

	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	pageSizeNum, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeNum < 1 || pageSizeNum > 100 {
		pageSizeNum = 20
	}

	// Extract filter parameters
	filters := h.extractFilters(c)

	// Get filtered items
	rawItems, totalCount, err := h.contentService.GetCollectionItems(c.Request.Context(), collection, q, filters, pageNum, pageSizeNum)
	if err != nil {
		h.ErrorResponse(c, err, fmt.Sprintf("Errore nel caricamento righe per %s", collection))
		return
	}

	// Convert to typed documents using mapper
	documents := h.documentMapper.ToModels(collection, rawItems)

	// Calculate pagination
	pagination := CalculatePaginationData(pageNum, pageSizeNum, totalCount)

	data := models.CollectionPageData{
		PageData: models.PageData{
			Collection:  collection,
			QueryString: c.Request.URL.RawQuery,
		},
		Documents:  documents,
		Query:      q,
		Page:       pageNum,
		PageSize:   pageSizeNum,
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

// formatTraitContent cleans content and adds line breaks before bold trait names to improve readability
func formatTraitContent(content string) string {
	// Use simple string replacements for safety
	formatted := content

	// Remove unwanted category sections
	formatted = strings.ReplaceAll(formatted, "### Talenti Generali", "")
	formatted = strings.ReplaceAll(formatted, "### Talenti Razziali", "")
	formatted = strings.ReplaceAll(formatted, "### Categoria Background", "")

	// Simple trait formatting - add line breaks before bold trait names
	// Using safe regex patterns that we know work
	formatted = regexp.MustCompile(`(\s)(\*\*\*[^*]+\.\*\*\*)`).ReplaceAllString(formatted, "\n\n$2")
	formatted = regexp.MustCompile(`(\s)(\*\*[^*]+\.\*\*)`).ReplaceAllString(formatted, "\n\n$2")

	// Clean up multiple newlines
	formatted = regexp.MustCompile(`\n{3,}`).ReplaceAllString(formatted, "\n\n")
	formatted = strings.TrimSpace(formatted)

	return formatted
}

// setCacheHeaders sets appropriate cache headers for D&D content responses
func (h *Handlers) setCacheHeaders(c *gin.Context, cacheTypeStr string) {
	// Map string cache type to config type
	var cacheType config.CacheType
	switch cacheTypeStr {
	case "home":
		cacheType = config.CacheTypeHome
	case "collection":
		cacheType = config.CacheTypeCollection
	case "item":
		cacheType = config.CacheTypeItem
	case "search":
		cacheType = config.CacheTypeSearch
	default:
		cacheType = config.CacheTypeCollection
	}

	maxAge := config.GetCacheDuration(cacheType)

	if maxAge > 0 {
		c.Header("Cache-Control", fmt.Sprintf("max-age=%d, public", maxAge))
	} else {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
	}
}

// Helper methods
func (h *Handlers) getCollectionTitle(collection string) string {
	// Use configuration if available
	if h.collectionMetadata != nil {
		return h.collectionMetadata.GetTitle(collection)
	}

	// Fallback to config package titles
	return config.GetCollectionTitle(collection)
}

// extractFilters extracts all query parameters (except special ones) as filter parameters
func (h *Handlers) extractFilters(c *gin.Context) map[string]string {
	filters := make(map[string]string)

	// Skip special query parameters that are not filters
	skipParams := map[string]bool{
		"page":      true,
		"page_size": true,
		"q":         true,
	}

	// Extract all other query parameters as potential filters
	for param, values := range c.Request.URL.Query() {
		if !skipParams[param] && len(values) > 0 && values[0] != "" {
			filters[param] = values[0]
		}
	}

	return filters
}

// handleQuickSearch handles HTMX requests for breadcrumb quick search
func (h *Handlers) handleQuickSearch(c *gin.Context) {
	collection := c.Param("collection")
	query := c.Query("q")

	// If no query, return empty results
	if query == "" {
		h.setCacheHeaders(c, "search")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(""))
		return
	}

	// Get search results (limit to 5 for quick search)
	rawItems, _, err := h.contentService.GetCollectionItems(c.Request.Context(), collection, query, nil, 1, 5)
	if err != nil {
		h.setCacheHeaders(c, "search")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(""))
		return
	}

	// Generate HTML for results
	html := ""
	for _, item := range rawItems {
		title := mappers.GetString(item, "title", "")
		slug := mappers.GetString(item, "_id", "")

		if title != "" && slug != "" {
			html += fmt.Sprintf(`<a href="/%s/%s" class="search-result" tabindex="-1">
				<div class="search-result-title">%s</div>
			</a>`, collection, slug, title)
		}
	}

	if html == "" {
		html = `<div class="search-result" style="color: var(--notion-text-light);">Nessun risultato trovato</div>`
	}

	h.setCacheHeaders(c, "search")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// getDefaultCollections returns default collections from the registry
func (h *Handlers) getDefaultCollections() []map[string]any {
	allCollections := collections.GetAllWithInfo()
	result := make([]map[string]any, 0, len(allCollections))

	for _, info := range allCollections {
		result = append(result, map[string]any{
			"name":  info.Name.String(),
			"label": info.Title,
			"count": 0,
		})
	}

	return result
}

// handleGlobalSearch handles global search across all collections
func (h *Handlers) handleGlobalSearch(c *gin.Context) {
	query := c.Query("q")

	if query == "" {
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Search across all collections (limit 5 results per collection)
	searchResults, err := h.contentService.GlobalSearch(c.Request.Context(), query, 5)
	if err != nil {
		h.ErrorResponse(c, err, "Errore durante la ricerca")
		return
	}

	// Convert to page data
	results := make([]models.CollectionSearchResult, 0, len(searchResults))
	totalResults := int64(0)

	for _, sr := range searchResults {
		// Convert raw items to typed documents
		documents := h.documentMapper.ToModels(sr.Collection, sr.Items)

		results = append(results, models.CollectionSearchResult{
			CollectionName:  sr.Collection,
			CollectionLabel: h.getCollectionTitle(sr.Collection),
			Documents:       documents,
			Total:           sr.Total,
			HasMore:         sr.Total > int64(len(sr.Items)),
		})

		totalResults += sr.Total
	}

	data := models.SearchPageData{
		PageData: models.PageData{
			Title:       fmt.Sprintf("Risultati per: %s", query),
			Description: "Risultati della ricerca globale",
			QueryString: c.Request.URL.RawQuery,
		},
		Query:   query,
		Results: results,
		Total:   totalResults,
	}

	content, err := h.templateEngine.RenderSearch(data)
	if err != nil {
		h.ErrorResponse(c, err, "Errore nel rendering della pagina di ricerca")
		return
	}

	h.setCacheHeaders(c, "search")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

// handleSearchDropdown handles HTMX dropdown search requests
func (h *Handlers) handleSearchDropdown(c *gin.Context) {
	query := c.Query("q")

	// Return empty if no query
	if query == "" {
		h.setCacheHeaders(c, "search")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(""))
		return
	}

	// Search across all collections (limit 3 results per collection for dropdown)
	searchResults, err := h.contentService.GlobalSearch(c.Request.Context(), query, 3)
	if err != nil {
		h.setCacheHeaders(c, "search")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(""))
		return
	}

	// Convert to page data
	results := make([]models.CollectionSearchResult, 0, len(searchResults))

	for _, sr := range searchResults {
		// Convert raw items to typed documents
		documents := h.documentMapper.ToModels(sr.Collection, sr.Items)

		results = append(results, models.CollectionSearchResult{
			CollectionName:  sr.Collection,
			CollectionLabel: h.getCollectionTitle(sr.Collection),
			Documents:       documents,
			Total:           sr.Total,
			HasMore:         sr.Total > int64(len(sr.Items)),
		})
	}

	content, err := h.templateEngine.RenderSearchDropdown(results, query)
	if err != nil {
		h.setCacheHeaders(c, "search")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(""))
		return
	}

	h.setCacheHeaders(c, "search")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}
