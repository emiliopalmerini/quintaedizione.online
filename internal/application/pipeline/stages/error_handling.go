package stages

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/events"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/pipeline"
)

// ErrorHandlingStage handles and reports errors from the pipeline
type ErrorHandlingStage struct {
	eventBus events.EventBus
	logger   parsers.Logger
}

// NewErrorHandlingStage creates a new error handling stage
func NewErrorHandlingStage(eventBus events.EventBus, logger parsers.Logger) *ErrorHandlingStage {
	if logger == nil {
		logger = &parsers.NoOpLogger{}
	}

	return &ErrorHandlingStage{
		eventBus: eventBus,
		logger:   logger,
	}
}

// Name returns the stage name
func (s *ErrorHandlingStage) Name() string {
	return "error_handling"
}

// Process handles and reports any errors that occurred during processing
func (s *ErrorHandlingStage) Process(ctx context.Context, data *pipeline.ProcessingData) error {
	if data == nil {
		return fmt.Errorf("processing data cannot be nil")
	}

	errorCount := len(data.Errors)

	if errorCount == 0 {
		s.logger.Debug("no errors to handle for %s", data.FilePath)
		return nil
	}

	s.logger.Debug("handling %d errors for %s", errorCount, data.FilePath)

	// Categorize errors
	errorSummary := s.categorizeErrors(data.Errors)

	// Store error metadata
	data.Metadata["error_count"] = errorCount
	data.Metadata["error_summary"] = errorSummary
	data.Metadata["has_errors"] = true

	// Log error summary
	s.logger.Error("processing completed with %d errors for %s:", errorCount, data.FilePath)
	for category, count := range errorSummary {
		s.logger.Error("  - %s: %d", category, count)
	}

	// Log individual errors for debugging
	for i, err := range data.Errors {
		s.logger.Debug("  Error %d: %v", i+1, err)
	}

	// Create error report
	errorReport := s.createErrorReport(data)
	data.Metadata["error_report"] = errorReport

	// Decide whether errors should fail the entire pipeline
	// For now, we continue processing but mark as having errors
	return nil
}

// categorizeErrors categorizes errors by type
func (s *ErrorHandlingStage) categorizeErrors(errors []error) map[string]int {
	categories := map[string]int{
		"parsing":        0,
		"validation":     0,
		"transformation": 0,
		"persistence":    0,
		"file_io":        0,
		"other":          0,
	}

	for _, err := range errors {
		category := s.categorizeError(err)
		categories[category]++
	}

	return categories
}

// categorizeError determines the category of an error
func (s *ErrorHandlingStage) categorizeError(err error) string {
	errorStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errorStr, "parse") || strings.Contains(errorStr, "parsing"):
		return "parsing"
	case strings.Contains(errorStr, "validation") || strings.Contains(errorStr, "validate"):
		return "validation"
	case strings.Contains(errorStr, "transform") || strings.Contains(errorStr, "transformation"):
		return "transformation"
	case strings.Contains(errorStr, "persist") || strings.Contains(errorStr, "upsert") || strings.Contains(errorStr, "database"):
		return "persistence"
	case strings.Contains(errorStr, "file") || strings.Contains(errorStr, "read") || strings.Contains(errorStr, "not found"):
		return "file_io"
	default:
		return "other"
	}
}

// createErrorReport creates a detailed error report
func (s *ErrorHandlingStage) createErrorReport(data *pipeline.ProcessingData) map[string]any {
	report := map[string]any{
		"file_path":   data.FilePath,
		"collection":  data.WorkItem.Collection,
		"error_count": len(data.Errors),
		"timestamp":   time.Now(),
	}

	// Add processing statistics
	if parsedCount, ok := data.Metadata["parsed_count"].(int); ok {
		report["parsed_count"] = parsedCount
	}
	if writtenCount, ok := data.Metadata["written_count"].(int); ok {
		report["written_count"] = writtenCount
	}
	if validationErrors, ok := data.Metadata["validation_errors"].(int); ok {
		report["validation_errors"] = validationErrors
	}

	// Add error details
	var errorDetails []map[string]any
	for i, err := range data.Errors {
		detail := map[string]any{
			"index":    i + 1,
			"message":  err.Error(),
			"category": s.categorizeError(err),
		}
		errorDetails = append(errorDetails, detail)
	}
	report["error_details"] = errorDetails

	return report
}
