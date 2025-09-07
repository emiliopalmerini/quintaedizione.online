package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/services"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/templates"
	"github.com/gin-gonic/gin"
)

// Handlers contains web handlers for the editor
type Handlers struct {
	contentService *services.ContentService
	templateEngine *templates.Engine
}

// NewHandlers creates a new Handlers instance
func NewHandlers(contentService *services.ContentService, templateEngine *templates.Engine) *Handlers {
	return &Handlers{
		contentService: contentService,
		templateEngine: templateEngine,
	}
}

// RegisterRoutes registers all web routes
func (h *Handlers) RegisterRoutes(router *gin.Engine) {
	// Main page
	router.GET("/", h.handleHome)

	// Collection pages
	router.GET("/:collection", h.handleCollectionList)
	router.GET("/:collection/rows", h.handleCollectionRows) // HTMX rows endpoint
	router.GET("/:collection/:slug", h.handleItemDetail)
}

// handleHome renders the home page
func (h *Handlers) handleHome(c *gin.Context) {
	// Get collection stats for display
	collections, err := h.contentService.GetCollectionStats(c.Request.Context())
	if err != nil {
		// If error, use default collections without counts
		collections = []map[string]any{
			{"name": "incantesimi", "label": "Incantesimi", "count": 0},
			{"name": "mostri", "label": "Mostri", "count": 0},
			{"name": "classi", "label": "Classi", "count": 0},
			{"name": "backgrounds", "label": "Background", "count": 0},
			{"name": "equipaggiamento", "label": "Equipaggiamento", "count": 0},
			{"name": "oggetti_magici", "label": "Oggetti Magici", "count": 0},
			{"name": "armi", "label": "Armi", "count": 0},
			{"name": "armature", "label": "Armature", "count": 0},
		}
	}

	// Calculate total items
	total := int64(0)
	for _, col := range collections {
		if count, ok := col["count"].(int64); ok {
			total += count
			// Also set label from title if not present
			if title, ok := col["title"].(string); ok && col["label"] == nil {
				col["label"] = title
			}
		}
	}

	data := gin.H{
		"title":       "5e SRD 2024",
		"description": "Il Fantastico Visualizzatore di SRD (5e 2024)",
		"collections": collections,
		"total":       total,
	}

	h.renderTemplate(c, "home.html", data)
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

	// Get items from service
	items, totalCount, err := h.contentService.GetCollectionItems(c.Request.Context(), collection, q, pageNum, pageSizeNum)
	if err != nil {
		h.renderError(c, fmt.Sprintf("Errore nel caricamento della collezione %s: %v", collection, err), http.StatusInternalServerError)
		return
	}

	// Calculate pagination
	totalPages := int((totalCount + int64(pageSizeNum) - 1) / int64(pageSizeNum))
	startItem := (pageNum-1)*pageSizeNum + 1
	endItem := pageNum * pageSizeNum
	if endItem > int(totalCount) {
		endItem = int(totalCount)
	}

	data := gin.H{
		"title":       getCollectionTitle(collection),
		"collection":  collection,
		"documents":   items,
		"q":           q,
		"page":        pageNum,
		"page_size":   pageSizeNum,
		"total":       totalCount,
		"total_pages": totalPages,
		"has_next":    pageNum < totalPages,
		"has_prev":    pageNum > 1,
		"start_item":  startItem,
		"end_item":    endItem,
		"qs":          c.Request.URL.RawQuery,
	}

	h.renderTemplate(c, "collection.html", data)
}

// handleItemDetail renders an item detail page
func (h *Handlers) handleItemDetail(c *gin.Context) {
	collection := c.Param("collection")
	slug := c.Param("slug")

	item, err := h.contentService.GetItem(c.Request.Context(), collection, slug)
	if err != nil {
		h.renderError(c, "Elemento non trovato", http.StatusNotFound)
		return
	}

	// Prepare markdown content
	var bodyRaw string
	var bodyHTML string

	// Try different fields for content
	if content, ok := item["contenuto_markdown"].(string); ok {
		bodyRaw = content
	} else if desc, ok := item["descrizione"].(string); ok {
		bodyRaw = desc
	} else if body, ok := item["body"].(string); ok {
		bodyRaw = body
	}

	// For now, bodyHTML = bodyRaw (client-side rendering)
	bodyHTML = bodyRaw

	data := gin.H{
		"doc_title":        item["nome"],
		"doc_id":           slug,
		"collection":       collection,
		"collection_label": getCollectionTitle(collection),
		"body_raw":         bodyRaw,
		"body_html":        bodyHTML,
		"qs":               c.Request.URL.RawQuery,
		// TODO: Add prev_id and next_id for navigation
		"prev_id": nil,
		"next_id": nil,
	}

	h.renderTemplate(c, "item.html", data)
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

	// Get filtered items
	items, totalCount, err := h.contentService.GetCollectionItems(c.Request.Context(), collection, q, pageNum, pageSizeNum)
	if err != nil {
		h.renderError(c, fmt.Sprintf("Errore nel caricamento righe per %s: %v", collection, err), http.StatusInternalServerError)
		return
	}

	// Calculate pagination
	totalPages := int((totalCount + int64(pageSizeNum) - 1) / int64(pageSizeNum))
	startItem := (pageNum-1)*pageSizeNum + 1
	endItem := pageNum * pageSizeNum
	if endItem > int(totalCount) {
		endItem = int(totalCount)
	}

	data := gin.H{
		"collection":  collection,
		"documents":   items,
		"total":       totalCount,
		"page":        pageNum,
		"page_size":   pageSizeNum,
		"total_pages": totalPages,
		"has_next":    pageNum < totalPages,
		"has_prev":    pageNum > 1,
		"start_item":  startItem,
		"end_item":    endItem,
		"qs":          c.Request.URL.RawQuery,
	}

	h.renderTemplate(c, "rows.html", data)
}

// Helper methods
func (h *Handlers) renderTemplate(c *gin.Context, template string, data gin.H) {
	content, err := h.templateEngine.Render(template, data)
	if err != nil {
		h.renderError(c, "Errore nel rendering della pagina", http.StatusInternalServerError)
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

func (h *Handlers) renderError(c *gin.Context, message string, statusCode int) {
	data := gin.H{
		"title":       "Errore",
		"error":       message,
		"status_code": statusCode,
	}

	content, err := h.templateEngine.Render("error.html", data)
	if err != nil {
		// Fallback to simple error response
		c.String(statusCode, "Errore: %s", message)
		return
	}

	c.Data(statusCode, "text/html; charset=utf-8", []byte(content))
}

func getCollectionTitle(collection string) string {
	titles := map[string]string{
		"incantesimi":     "Incantesimi",
		"mostri":          "Mostri",
		"classi":          "Classi",
		"backgrounds":     "Background",
		"equipaggiamento": "Equipaggiamento",
		"armi":            "Armi",
		"armature":        "Armature",
		"oggetti_magici":  "Oggetti Magici",
		"talenti":         "Talenti",
		"servizi":         "Servizi",
		"strumenti":       "Strumenti",
		"animali":         "Animali",
	}

	if title, exists := titles[collection]; exists {
		return title
	}

	return collection
}
