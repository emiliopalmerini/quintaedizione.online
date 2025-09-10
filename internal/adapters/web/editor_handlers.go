package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/web/models"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/services"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/templates"
	"github.com/gin-gonic/gin"
)

// Handlers contains web handlers for the editor
type Handlers struct {
	contentService *services.ContentService
	templateEngine *templates.TemplEngine
}

// NewHandlers creates a new Handlers instance
func NewHandlers(contentService *services.ContentService, templateEngine *templates.TemplEngine) *Handlers {
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
			{"name": "equipaggiamenti", "label": "Equipaggiamento", "count": 0},
			{"name": "oggetti_magici", "label": "Oggetti Magici", "count": 0},
			{"name": "armi", "label": "Armi", "count": 0},
			{"name": "armature", "label": "Armature", "count": 0},
			{"name": "talenti", "label": "Talenti", "count": 0},
			{"name": "servizi", "label": "Servizi", "count": 0},
			{"name": "strumenti", "label": "Strumenti", "count": 0},
			{"name": "animali", "label": "Animali", "count": 0},
			{"name": "regole", "label": "Regole", "count": 0},
			{"name": "cavalcature_veicoli", "label": "Cavalcature e Veicoli", "count": 0},
		}
	}

	// Convert collections to typed format and calculate total
	typedCollections := make([]models.Collection, 0, len(collections))
	total := int64(0)
	
	for _, col := range collections {
		collection := models.Collection{
			Name:  col["name"].(string),
			Count: 0,
		}
		
		if count, ok := col["count"].(int64); ok {
			collection.Count = count
			total += count
		}
		
		if label, ok := col["label"].(string); ok {
			collection.Label = label
		} else if title, ok := col["title"].(string); ok {
			collection.Label = title
		} else {
			collection.Label = collection.Name
		}
		
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

	// Get items from service
	rawItems, totalCount, err := h.contentService.GetCollectionItems(c.Request.Context(), collection, q, pageNum, pageSizeNum)
	if err != nil {
		h.ErrorResponse(c, err, fmt.Sprintf("Errore nel caricamento della collezione %s", collection))
		return
	}

	// Convert to typed documents
	documents := make([]models.Document, 0, len(rawItems))
	for _, item := range rawItems {
		doc := models.Document{}
		
		// Extract _id from document root
		if id, ok := item["_id"].(string); ok {
			doc.ID = id
		}
		
		// Extract nome and slug from value object
		if valueObj, ok := item["value"].(map[string]interface{}); ok {
			if nome, ok := valueObj["nome"].(string); ok {
				doc.Nome = nome
			}
			if slug, ok := valueObj["slug"].(string); ok {
				doc.Slug = slug
			}
		}
		
		// Extract display elements (processed by ContentService)
		if displayElements, ok := item["display_elements"].([]interface{}); ok {
			for _, elem := range displayElements {
				if elemMap, ok := elem.(map[string]interface{}); ok {
					if value, ok := elemMap["value"].(string); ok {
						doc.DisplayElements = append(doc.DisplayElements, models.DocumentDisplayField{Value: value})
					}
				}
			}
		}
		
		// Extract translated flag from document root
		if translated, ok := item["translated"].(bool); ok {
			doc.Translated = translated
		}
		
		documents = append(documents, doc)
	}

	// Calculate pagination
	totalPages := int((totalCount + int64(pageSizeNum) - 1) / int64(pageSizeNum))
	startItem := (pageNum-1)*pageSizeNum + 1
	endItem := pageNum * pageSizeNum
	if endItem > int(totalCount) {
		endItem = int(totalCount)
	}

	data := models.CollectionPageData{
		PageData: models.PageData{
			Title:       getCollectionTitle(collection),
			Collection:  collection,
			QueryString: c.Request.URL.RawQuery,
		},
		Documents:   documents,
		Query:       q,
		Page:        pageNum,
		PageSize:    pageSizeNum,
		Total:       totalCount,
		TotalPages:  totalPages,
		HasNext:     pageNum < totalPages,
		HasPrev:     pageNum > 1,
		StartItem:   startItem,
		EndItem:     endItem,
	}

	content, err := h.templateEngine.RenderCollection(data)
	if err != nil {
		h.ErrorResponse(c, err, "Errore nel rendering della pagina collezione")
		return
	}

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

	// Prepare markdown content
	var bodyRaw string
	var bodyHTML string

	// Try different fields for content
	if content, ok := item["contenuto"].(string); ok {
		bodyRaw = content
	} else if content, ok := item["contenuto_markdown"].(string); ok {
		bodyRaw = content
	} else if desc, ok := item["descrizione"].(string); ok {
		bodyRaw = desc
	} else if body, ok := item["body"].(string); ok {
		bodyRaw = body
	}

	// For now, bodyHTML = bodyRaw (client-side rendering)
	bodyHTML = bodyRaw

	// Get navigation items
	prevSlug, nextSlug, err := h.contentService.GetAdjacentItems(c.Request.Context(), collection, slug)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: Could not get adjacent items for %s/%s: %v\n", collection, slug, err)
	}

	// Get doc title
	docTitle := ""
	if nome, ok := item["nome"].(string); ok {
		docTitle = nome
	}

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
		CollectionLabel: getCollectionTitle(collection),
	}

	content, err := h.templateEngine.RenderItem(data)
	if err != nil {
		h.ErrorResponse(c, err, "Errore nel rendering della pagina elemento")
		return
	}

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

	// Get filtered items
	rawItems, totalCount, err := h.contentService.GetCollectionItems(c.Request.Context(), collection, q, pageNum, pageSizeNum)
	if err != nil {
		h.ErrorResponse(c, err, fmt.Sprintf("Errore nel caricamento righe per %s", collection))
		return
	}

	// Convert to typed documents (same logic as collection handler)
	documents := make([]models.Document, 0, len(rawItems))
	for _, item := range rawItems {
		doc := models.Document{}
		
		// Extract _id from document root
		if id, ok := item["_id"].(string); ok {
			doc.ID = id
		}
		
		// Extract nome and slug from value object
		if valueObj, ok := item["value"].(map[string]interface{}); ok {
			if nome, ok := valueObj["nome"].(string); ok {
				doc.Nome = nome
			}
			if slug, ok := valueObj["slug"].(string); ok {
				doc.Slug = slug
			}
		}
		
		// Extract display elements (processed by ContentService)
		if displayElements, ok := item["display_elements"].([]interface{}); ok {
			for _, elem := range displayElements {
				if elemMap, ok := elem.(map[string]interface{}); ok {
					if value, ok := elemMap["value"].(string); ok {
						doc.DisplayElements = append(doc.DisplayElements, models.DocumentDisplayField{Value: value})
					}
				}
			}
		}
		
		// Extract translated flag from document root
		if translated, ok := item["translated"].(bool); ok {
			doc.Translated = translated
		}
		
		documents = append(documents, doc)
	}

	// Calculate pagination
	totalPages := int((totalCount + int64(pageSizeNum) - 1) / int64(pageSizeNum))
	startItem := (pageNum-1)*pageSizeNum + 1
	endItem := pageNum * pageSizeNum
	if endItem > int(totalCount) {
		endItem = int(totalCount)
	}

	data := models.CollectionPageData{
		PageData: models.PageData{
			Collection:  collection,
			QueryString: c.Request.URL.RawQuery,
		},
		Documents:   documents,
		Query:       q,
		Page:        pageNum,
		PageSize:    pageSizeNum,
		Total:       totalCount,
		TotalPages:  totalPages,
		HasNext:     pageNum < totalPages,
		HasPrev:     pageNum > 1,
		StartItem:   startItem,
		EndItem:     endItem,
	}

	content, err := h.templateEngine.RenderRows(data)
	if err != nil {
		h.ErrorResponse(c, err, "Errore nel rendering delle righe")
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

// Helper methods
func getCollectionTitle(collection string) string {
	titles := map[string]string{
		"incantesimi":         "Incantesimi",
		"mostri":              "Mostri",
		"classi":              "Classi",
		"backgrounds":         "Background",
		"equipaggiamenti":     "Equipaggiamento",
		"armi":                "Armi",
		"armature":            "Armature",
		"oggetti_magici":      "Oggetti Magici",
		"talenti":             "Talenti",
		"servizi":             "Servizi",
		"strumenti":           "Strumenti",
		"animali":             "Animali",
		"regole":              "Regole",
		"cavalcature_veicoli": "Cavalcature e Veicoli",
	}

	if title, exists := titles[collection]; exists {
		return title
	}

	return collection
}
