package handlers

import (
	"net/http"
	"strconv"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/services"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/infrastructure"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/templates"
	"github.com/gin-gonic/gin"
)

// IngestHandler handles ingestion HTTP requests
type IngestHandler struct {
	ingestService  *services.IngestService
	templateEngine *templates.Engine
	defaultWork    []parsers.WorkItem
	inputDir       string
}

// NewIngestHandler creates a new ingest handler
func NewIngestHandler(ingestService *services.IngestService, templateEngine *templates.Engine, inputDir string) *IngestHandler {
	return &IngestHandler{
		ingestService:  ingestService,
		templateEngine: templateEngine,
		defaultWork:    parsers.CreateDefaultWork(),
		inputDir:       inputDir,
	}
}

// GetIndex handles GET / - show parser form
func (h *IngestHandler) GetIndex(c *gin.Context) {
	workItems := make([]map[string]interface{}, len(h.defaultWork))
	for i, item := range h.defaultWork {
		workItems[i] = map[string]interface{}{
			"idx":        i,
			"collection": item.Collection,
			"filename":   item.Filename,
		}
	}

	data := map[string]interface{}{
		"env": map[string]interface{}{
			"input_dir": h.inputDir,
			"db_name":   "dnd", // TODO: get from config
			"dry_run":   true,
		},
		"work_items": workItems,
		"messages":   []string{},
		"selected":   []int{},
	}

	// Check if this is an HTMX request or API request
	if c.GetHeader("HX-Request") != "" || c.GetHeader("Accept") == "application/json" {
		c.JSON(http.StatusOK, data)
		return
	}

	// Render HTML template
	html, err := h.templateEngine.Render("parser.html", data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render template: " + err.Error()})
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// PostRun handles POST /run - execute parsing
func (h *IngestHandler) PostRun(c *gin.Context) {
	// Parse form data or JSON
	var req struct {
		InputDir string `form:"input_dir" json:"input_dir" binding:"required"`
		DBName   string `form:"db_name" json:"db_name" binding:"required"`
		Selected []int  `form:"selected" json:"selected"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle checkbox manually - if present, it's checked (true), otherwise false
	dryRun := c.PostForm("dry_run") != ""

	selected := req.Selected

	// Build work items list
	workItems := make([]map[string]interface{}, len(h.defaultWork))
	for i, item := range h.defaultWork {
		workItems[i] = map[string]interface{}{
			"idx":        i,
			"collection": item.Collection,
			"filename":   item.Filename,
		}
	}

	var messages []string
	var total int

	// Validate input
	if len(selected) == 0 {
		messages = append(messages, "Nessuna collezione selezionata.")
	}

	if !infrastructure.FileExists(req.InputDir) {
		messages = append(messages, "Cartella input non trovata: "+req.InputDir)
	}

	// Execute ingestion if valid
	if len(selected) > 0 && infrastructure.FileExists(req.InputDir) {
		// Filter work items by selection
		var selectedWork []parsers.WorkItem
		for _, idx := range selected {
			if idx >= 0 && idx < len(h.defaultWork) {
				selectedWork = append(selectedWork, h.defaultWork[idx])
			}
		}

		if len(selectedWork) > 0 {
			results, err := h.ingestService.ExecuteIngest(req.InputDir, selectedWork, dryRun)
			if err != nil {
				messages = append(messages, "Errore durante l'ingestion: "+err.Error())
			} else {
				// Process results
				for _, result := range results {
					if result.Error != "" {
						messages = append(messages, "Errore parser in "+result.Filename+": "+result.Error)
						continue
					}

					messages = append(messages, "Parsing "+result.Filename+" → "+result.Collection)
					messages = append(messages, "Estratti "+strconv.Itoa(result.Parsed)+" documenti da "+result.Filename)

					if dryRun {
						total += result.Parsed
					} else {
						total += result.Written
						messages = append(messages, "Upsert "+strconv.Itoa(result.Written)+" documenti in "+req.DBName+"."+result.Collection)
					}
				}
			}
		}
	}

	// Add final summary
	if dryRun {
		messages = append(messages, "Dry-run completato. Totale analizzati: "+strconv.Itoa(total))
	} else {
		messages = append(messages, "Fatto. Totale upsert: "+strconv.Itoa(total))
	}

	// Return response
	env := map[string]interface{}{
		"input_dir": req.InputDir,
		"db_name":   req.DBName,
		"dry_run":   dryRun,
	}

	response := map[string]interface{}{
		"env":        env,
		"work_items": workItems,
		"messages":   messages,
		"selected":   selected,
	}

	// Check if this is an HTMX request
	if c.GetHeader("HX-Request") != "" {
		// Return just the results section for HTMX
		if len(messages) > 0 {
			html := `<div class="results-section">
				<h2 class="notion-h2">Risultati</h2>
				<div class="messages" style="background: var(--notion-bg-code); border-radius: 6px; padding: 1rem; font-family: 'Monaco', 'Menlo', monospace; font-size: 12px; line-height: 1.5;">`
			for _, msg := range messages {
				html += `<div style="margin-bottom: 0.25rem; color: var(--notion-text-light);">` + msg + `</div>`
			}
			html += `</div></div>`
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, html)
		} else {
			c.String(http.StatusOK, "")
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetHealthz handles GET /healthz - health check
func (h *IngestHandler) GetHealthz(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}
