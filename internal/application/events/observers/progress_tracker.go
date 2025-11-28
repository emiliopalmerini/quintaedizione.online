package observers

import (
	"sync"
	"time"

	"github.com/emiliopalmerini/quintaedizione.online/internal/application/events"
	"github.com/emiliopalmerini/quintaedizione.online/internal/application/parsers"
)

type ProgressTracker struct {
	totalFiles     int
	processedFiles int
	currentFile    string
	startTime      time.Time
	mu             sync.RWMutex
	logger         parsers.Logger
	eventBus       events.EventBus

	successfulFiles int
	failedFiles     int
	totalParsed     int
	totalWritten    int
	totalErrors     int
}

func NewProgressTracker(totalFiles int, eventBus events.EventBus, logger parsers.Logger) *ProgressTracker {
	if logger == nil {
		logger = &parsers.NoOpLogger{}
	}

	return &ProgressTracker{
		totalFiles:      totalFiles,
		processedFiles:  0,
		startTime:       time.Now(),
		logger:          logger,
		eventBus:        eventBus,
		successfulFiles: 0,
		failedFiles:     0,
		totalParsed:     0,
		totalWritten:    0,
		totalErrors:     0,
	}
}

func (pt *ProgressTracker) HandleEvent(event events.Event) {
	switch e := event.(type) {
	case *events.PipelineStartedEvent:
		pt.handlePipelineStarted(e)
	case *events.PipelineCompletedEvent:
		pt.handlePipelineCompleted(e)
	case *events.PipelineFailedEvent:
		pt.handlePipelineFailed(e)
	case *events.FileProcessingCompletedEvent:
		pt.handleFileProcessingCompleted(e)
	}
}

func (pt *ProgressTracker) handlePipelineStarted(event *events.PipelineStartedEvent) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.currentFile = event.FilePath
	pt.logger.Debug("started processing file: %s", event.FilePath)

	if pt.eventBus != nil {
		pt.publishProgressEvent()
	}
}

func (pt *ProgressTracker) handlePipelineCompleted(event *events.PipelineCompletedEvent) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.processedFiles++
	pt.totalParsed += event.ParsedCount

	if event.HasErrors {
		pt.failedFiles++
		pt.logger.Info("completed file %s with errors (%d/%d)", event.FilePath, pt.processedFiles, pt.totalFiles)
	} else {
		pt.successfulFiles++
		pt.logger.Debug("completed file %s successfully (%d/%d)", event.FilePath, pt.processedFiles, pt.totalFiles)
	}

	if pt.eventBus != nil {
		pt.publishProgressEvent()
	}

	if pt.processedFiles >= pt.totalFiles {
		pt.publishFinalSummary()
	}
}

func (pt *ProgressTracker) handlePipelineFailed(event *events.PipelineFailedEvent) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.processedFiles++
	pt.failedFiles++
	pt.totalErrors++

	pt.logger.Error("failed to process file %s: %v (%d/%d)", event.FilePath, event.Error, pt.processedFiles, pt.totalFiles)

	if pt.eventBus != nil {
		pt.publishProgressEvent()
	}

	if pt.processedFiles >= pt.totalFiles {
		pt.publishFinalSummary()
	}
}

func (pt *ProgressTracker) handleFileProcessingCompleted(event *events.FileProcessingCompletedEvent) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.totalWritten += event.WrittenCount
	pt.logger.Debug("file %s: parsed %d, written %d", event.FilePath, event.ParsedCount, event.WrittenCount)
}

func (pt *ProgressTracker) publishProgressEvent() {
	progressPercent := 0.0
	if pt.totalFiles > 0 {
		progressPercent = float64(pt.processedFiles) / float64(pt.totalFiles) * 100.0
	}

	progressEvent := &events.ProgressEvent{
		BaseEvent:       events.BaseEvent{EventTime: time.Now()},
		TotalFiles:      pt.totalFiles,
		ProcessedFiles:  pt.processedFiles,
		CurrentFile:     pt.currentFile,
		ProgressPercent: progressPercent,
	}

	pt.eventBus.PublishAsync(progressEvent)
}

func (pt *ProgressTracker) publishFinalSummary() {
	duration := time.Since(pt.startTime)

	pt.logger.Info("Processing completed in %v:", duration)
	pt.logger.Info("  Total files: %d", pt.totalFiles)
	pt.logger.Info("  Successful: %d", pt.successfulFiles)
	pt.logger.Info("  Failed: %d", pt.failedFiles)
	pt.logger.Info("  Total parsed: %d", pt.totalParsed)
	pt.logger.Info("  Total written: %d", pt.totalWritten)
	pt.logger.Info("  Total errors: %d", pt.totalErrors)

	summaryEvent := &events.ProcessingSummaryEvent{
		BaseEvent:       events.BaseEvent{EventTime: time.Now()},
		TotalFiles:      pt.totalFiles,
		SuccessfulFiles: pt.successfulFiles,
		FailedFiles:     pt.failedFiles,
		TotalParsed:     pt.totalParsed,
		TotalWritten:    pt.totalWritten,
		TotalErrors:     pt.totalErrors,
		Duration:        duration,
	}

	if pt.eventBus != nil {
		pt.eventBus.PublishAsync(summaryEvent)
	}
}

func (pt *ProgressTracker) GetProgress() ProgressInfo {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	progressPercent := 0.0
	if pt.totalFiles > 0 {
		progressPercent = float64(pt.processedFiles) / float64(pt.totalFiles) * 100.0
	}

	return ProgressInfo{
		TotalFiles:      pt.totalFiles,
		ProcessedFiles:  pt.processedFiles,
		SuccessfulFiles: pt.successfulFiles,
		FailedFiles:     pt.failedFiles,
		CurrentFile:     pt.currentFile,
		ProgressPercent: progressPercent,
		TotalParsed:     pt.totalParsed,
		TotalWritten:    pt.totalWritten,
		TotalErrors:     pt.totalErrors,
		Duration:        time.Since(pt.startTime),
		IsComplete:      pt.processedFiles >= pt.totalFiles,
	}
}

type ProgressInfo struct {
	TotalFiles      int           `json:"total_files"`
	ProcessedFiles  int           `json:"processed_files"`
	SuccessfulFiles int           `json:"successful_files"`
	FailedFiles     int           `json:"failed_files"`
	CurrentFile     string        `json:"current_file"`
	ProgressPercent float64       `json:"progress_percent"`
	TotalParsed     int           `json:"total_parsed"`
	TotalWritten    int           `json:"total_written"`
	TotalErrors     int           `json:"total_errors"`
	Duration        time.Duration `json:"duration"`
	IsComplete      bool          `json:"is_complete"`
}
