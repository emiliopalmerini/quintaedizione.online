package services

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/events"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/events/observers"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/pipeline"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/pipeline/stages"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/infrastructure"
)

// PipelineIngestService handles ingestion using the pipeline architecture
type PipelineIngestService struct {
	pipeline        *pipeline.Pipeline
	eventBus        events.EventBus
	progressTracker *observers.ProgressTracker
	errorCollector  *observers.ErrorCollector
	logger          parsers.Logger
	registry        *parsers.ParserRegistry
	repository      domain.ParserRepository
	documentBuilder *parsers.DocumentBuilder
	config          infrastructure.Config
}

// NewPipelineIngestService creates a new pipeline-based ingest service
func NewPipelineIngestService(
	repository domain.ParserRepository,
	registry *parsers.ParserRegistry,
	documentBuilder *parsers.DocumentBuilder,
	config infrastructure.Config,
	logger parsers.Logger,
) (*PipelineIngestService, error) {
	if logger == nil {
		logger = &parsers.NoOpLogger{}
	}

	// Create event bus using configuration
	eventBus := events.NewInMemoryEventBus(logger, config.Pipeline.EventBusBufferSize)

	// Create observers
	errorCollector := observers.NewErrorCollector(eventBus, logger)
	loggingObserver := observers.NewLoggingObserver(logger, observers.LogLevelInfo)

	// Subscribe observers to events
	eventBus.Subscribe("parsing_error", errorCollector.HandleEvent)
	eventBus.Subscribe("validation_error", errorCollector.HandleEvent)
	eventBus.Subscribe("persistence_error", errorCollector.HandleEvent)
	eventBus.Subscribe("pipeline_failed", errorCollector.HandleEvent)
	eventBus.Subscribe("stage_failed", errorCollector.HandleEvent)

	// Subscribe logging observer to all events
	eventTypes := []string{
		"pipeline_started", "pipeline_completed", "pipeline_failed",
		"stage_started", "stage_completed", "stage_failed",
		"file_processing_started", "file_processing_completed",
		"parsing_error", "validation_error", "persistence_error",
		"progress", "processing_summary",
	}
	for _, eventType := range eventTypes {
		eventBus.Subscribe(eventType, loggingObserver.HandleEvent)
	}

	service := &PipelineIngestService{
		eventBus:        eventBus,
		errorCollector:  errorCollector,
		logger:          logger,
		registry:        registry,
		repository:      repository,
		documentBuilder: documentBuilder,
		config:          config,
	}

	return service, nil
}

// ExecuteIngest processes work items using the pipeline architecture
func (s *PipelineIngestService) ExecuteIngest(baseDir string, workItems []parsers.WorkItem, dryRun bool) ([]*parsers.IngestResult, error) {
	if len(workItems) == 0 {
		s.logger.Info("no work items to process")
		return []*parsers.IngestResult{}, nil
	}

	// Create progress tracker for this batch
	progressTracker := observers.NewProgressTracker(len(workItems), s.eventBus, s.logger)
	s.progressTracker = progressTracker

	// Subscribe progress tracker to events
	s.eventBus.Subscribe("pipeline_started", progressTracker.HandleEvent)
	s.eventBus.Subscribe("pipeline_completed", progressTracker.HandleEvent)
	s.eventBus.Subscribe("pipeline_failed", progressTracker.HandleEvent)
	s.eventBus.Subscribe("file_processing_completed", progressTracker.HandleEvent)

	// Clear previous errors
	s.errorCollector.Clear()

	// Create pipeline for this execution
	pipelineStages := s.createPipelineStages(baseDir, dryRun)
	s.pipeline = pipeline.NewPipeline(pipelineStages, s.eventBus, s.logger, s.config.Pipeline)

	s.logger.Info("starting pipeline processing for %d files (dry_run: %v)", len(workItems), dryRun)

	var results []*parsers.IngestResult
	ctx := context.Background()

	// Process each work item through the pipeline
	for _, workItem := range workItems {
		result := s.processWorkItem(ctx, baseDir, workItem)
		results = append(results, result)
	}

	// Log final statistics
	stats := s.errorCollector.GetErrorStatistics()
	s.logger.Info("pipeline processing completed: %d total errors", stats.TotalErrors)

	return results, nil
}

// processWorkItem processes a single work item through the pipeline
func (s *PipelineIngestService) processWorkItem(ctx context.Context, baseDir string, workItem parsers.WorkItem) *parsers.IngestResult {
	result := parsers.NewIngestResult(workItem.Collection, workItem.Filename)

	// Create processing data
	fullPath := filepath.Join(baseDir, workItem.Filename)
	data := pipeline.NewProcessingData(&workItem, fullPath)

	// Execute pipeline
	if err := s.pipeline.Execute(ctx, data); err != nil {
		s.logger.Error("pipeline execution failed for %s: %v", workItem.Filename, err)
		result.SetError(fmt.Errorf("pipeline execution failed: %w", err))
		return result
	}

	// Extract results from processing data
	if len(data.Errors) > 0 {
		// There were errors during processing
		errorMsg := fmt.Sprintf("%d errors occurred during processing", len(data.Errors))
		result.SetError(fmt.Errorf("%s", errorMsg))
	}

	// Set counts from metadata
	if parsedCount, ok := data.Metadata["parsed_count"].(int); ok {
		result.Parsed = parsedCount
	}

	if writtenCount, ok := data.Metadata["written_count"].(int); ok {
		result.Written = writtenCount
	} else {
		result.Written = 0
	}

	// Set preview for dry run
	if preview, ok := data.Metadata["preview"]; ok {
		if previewData, ok := preview.([]map[string]any); ok {
			// Convert preview to JSON string (similar to original implementation)
			result.SetPreview(fmt.Sprintf("%v", previewData))
		}
	}

	return result
}

// createPipelineStages creates the pipeline stages for processing based on configuration
func (s *PipelineIngestService) createPipelineStages(baseDir string, dryRun bool) []pipeline.ProcessingStage {
	var pipelineStages []pipeline.ProcessingStage

	// Create stages based on configuration
	for _, stageName := range s.config.Pipeline.DefaultStages {
		switch stageName {
		case "file_reader":
			fileReader := stages.NewFileReaderStage(baseDir, s.eventBus, s.logger)
			pipelineStages = append(pipelineStages, fileReader)

		case "content_parser":
			contentParser := stages.NewContentParserStage(s.registry, s.eventBus, s.logger)
			pipelineStages = append(pipelineStages, contentParser)

		case "validation":
			if s.config.Pipeline.ValidationEnabled {
				validation := stages.NewValidationStage(s.eventBus, s.logger)
				pipelineStages = append(pipelineStages, validation)
			}

		case "transformation":
			if s.config.Pipeline.TransformationEnabled {
				transformation := stages.NewTransformationStage(s.documentBuilder, s.eventBus, s.logger)
				pipelineStages = append(pipelineStages, transformation)
			}

		case "persistence":
			persistence := stages.NewPersistenceStage(s.repository, dryRun, s.eventBus, s.logger)
			pipelineStages = append(pipelineStages, persistence)

		case "error_handling":
			errorHandling := stages.NewErrorHandlingStage(s.eventBus, s.logger)
			pipelineStages = append(pipelineStages, errorHandling)

		default:
			s.logger.Error("unknown pipeline stage: %s", stageName)
		}
	}

	return pipelineStages
}

// FilterWork filters work items by collection names (same as original)
func (s *PipelineIngestService) FilterWork(items []parsers.WorkItem, only []string) []parsers.WorkItem {
	if len(only) == 0 {
		return items
	}

	wanted := make(map[string]bool)
	for _, collection := range only {
		wanted[collection] = true
	}

	var filtered []parsers.WorkItem
	for _, item := range items {
		if wanted[item.Collection] {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// GetCollectionStats returns statistics for collections (same as original)
func (s *PipelineIngestService) GetCollectionStats() (map[string]int64, error) {
	if s.repository == nil {
		return nil, fmt.Errorf("repository not available")
	}

	collections := []string{
		"documenti", "classi", "backgrounds", "armi", "armature",
		"strumenti", "servizi", "equipaggiamento", "oggetti_magici",
		"incantesimi", "talenti", "mostri", "animali",
	}

	stats := make(map[string]int64)

	for _, collection := range collections {
		// Note: We'd need to add Count method to ParserRepository interface
		// For now, return 0 or implement differently
		stats[collection] = 0
	}

	return stats, nil
}

// GetProgressTracker returns the current progress tracker
func (s *PipelineIngestService) GetProgressTracker() *observers.ProgressTracker {
	return s.progressTracker
}

// GetErrorCollector returns the error collector
func (s *PipelineIngestService) GetErrorCollector() *observers.ErrorCollector {
	return s.errorCollector
}

// Close shuts down the service and cleans up resources
func (s *PipelineIngestService) Close() error {
	if s.eventBus != nil {
		s.eventBus.Close()
	}
	return nil
}
