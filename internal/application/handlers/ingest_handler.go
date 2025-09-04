package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/services"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// IngestHandler handles ingestion HTTP requests
type IngestHandler struct {
	ingestService *services.IngestService
	defaultWork   []domain.WorkItem
	inputDir      string
}

// NewIngestHandler creates a new ingest handler
func NewIngestHandler(ingestService *services.IngestService, inputDir string) *IngestHandler {
	return &IngestHandler{
		ingestService: ingestService,
		defaultWork:   parsers.CreateDefaultWork(),
		inputDir:      inputDir,
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

	response := map[string]interface{}{
		"env": map[string]interface{}{
			"input_dir": h.inputDir,
			"db_name":   "dnd", // TODO: get from config
			"dry_run":   true,
		},
		"work_items": workItems,
		"messages":   []string{},
		"selected":   []int{},
	}

	c.JSON(http.StatusOK, response)
}

// PostRun handles POST /run - execute parsing
func (h *IngestHandler) PostRun(c *gin.Context) {
	// Parse form data or JSON
	var req struct {
		InputDir  string `form:"input_dir" json:"input_dir" binding:"required"`
		DBName    string `form:"db_name" json:"db_name" binding:"required"`
		DryRun    bool   `form:"dry_run" json:"dry_run"`
		Selected  []int  `form:"selected" json:"selected"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	if !domain.FileExists(req.InputDir) {
		messages = append(messages, "Cartella input non trovata: "+req.InputDir)
	}

	// Execute ingestion if valid
	if len(selected) > 0 && domain.FileExists(req.InputDir) {
		// Filter work items by selection
		var selectedWork []domain.WorkItem
		for _, idx := range selected {
			if idx >= 0 && idx < len(h.defaultWork) {
				selectedWork = append(selectedWork, h.defaultWork[idx])
			}
		}

		if len(selectedWork) > 0 {
			results, err := h.ingestService.ExecuteIngest(req.InputDir, selectedWork, req.DryRun)
			if err != nil {
				messages = append(messages, "Errore durante l'ingestion: "+err.Error())
			} else {
				// Process results
				for _, result := range results {
					if result.Error != nil {
						messages = append(messages, "Errore parser in "+result.Filename+": "+*result.Error)
						continue
					}

					messages = append(messages, "Parsing "+result.Filename+" → "+result.Collection)
					messages = append(messages, "Estratti "+strconv.Itoa(result.Parsed)+" documenti da "+result.Filename)

					if req.DryRun {
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
	if req.DryRun {
		messages = append(messages, "Dry-run completato. Totale analizzati: "+strconv.Itoa(total))
	} else {
		messages = append(messages, "Fatto. Totale upsert: "+strconv.Itoa(total))
	}

	// Return response
	env := map[string]interface{}{
		"input_dir": req.InputDir,
		"db_name":   req.DBName,
		"dry_run":   req.DryRun,
	}

	response := map[string]interface{}{
		"env":        env,
		"work_items": workItems,
		"messages":   messages,
		"selected":   selected,
	}

	c.JSON(http.StatusOK, response)
}

// GetHealthz handles GET /healthz - health check
func (h *IngestHandler) GetHealthz(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}