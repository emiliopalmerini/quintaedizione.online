package observers

import (
	"fmt"
	"time"

	"github.com/emiliopalmerini/quintaedizione.online/internal/application/events"
	"github.com/emiliopalmerini/quintaedizione.online/internal/application/parsers"
)

// LoggingObserver logs all pipeline events for debugging and monitoring
type LoggingObserver struct {
	logger     parsers.Logger
	logLevel   LogLevel
	startTimes map[string]time.Time // Track stage start times
}

// LogLevel represents different logging levels
type LogLevel int

const (
	LogLevelError LogLevel = iota
	LogLevelInfo
	LogLevelDebug
)

// NewLoggingObserver creates a new logging observer
func NewLoggingObserver(logger parsers.Logger, logLevel LogLevel) *LoggingObserver {
	if logger == nil {
		logger = &parsers.NoOpLogger{}
	}

	return &LoggingObserver{
		logger:     logger,
		logLevel:   logLevel,
		startTimes: make(map[string]time.Time),
	}
}

// HandleEvent processes events for logging
func (lo *LoggingObserver) HandleEvent(event events.Event) {
	switch e := event.(type) {
	case *events.PipelineStartedEvent:
		lo.logPipelineStarted(e)
	case *events.PipelineCompletedEvent:
		lo.logPipelineCompleted(e)
	case *events.PipelineFailedEvent:
		lo.logPipelineFailed(e)
	case *events.StageStartedEvent:
		lo.logStageStarted(e)
	case *events.StageCompletedEvent:
		lo.logStageCompleted(e)
	case *events.StageFailedEvent:
		lo.logStageFailed(e)
	case *events.FileProcessingStartedEvent:
		lo.logFileProcessingStarted(e)
	case *events.FileProcessingCompletedEvent:
		lo.logFileProcessingCompleted(e)
	case *events.ParsingErrorEvent:
		lo.logParsingError(e)
	case *events.ValidationErrorEvent:
		lo.logValidationError(e)
	case *events.PersistenceErrorEvent:
		lo.logPersistenceError(e)
	case *events.ProgressEvent:
		lo.logProgress(e)
	case *events.ProcessingSummaryEvent:
		lo.logProcessingSummary(e)
	}
}

// logPipelineStarted logs pipeline started events
func (lo *LoggingObserver) logPipelineStarted(event *events.PipelineStartedEvent) {
	if lo.logLevel >= LogLevelInfo {
		lo.logger.Info("🚀 Started processing: %s → %s", event.FilePath, event.Collection)
	}

	// Store start time for duration calculation
	lo.startTimes[event.FilePath] = event.Timestamp()
}

// logPipelineCompleted logs pipeline completed events
func (lo *LoggingObserver) logPipelineCompleted(event *events.PipelineCompletedEvent) {
	duration := ""
	if startTime, exists := lo.startTimes[event.FilePath]; exists {
		duration = fmt.Sprintf(" (took %v)", event.Timestamp().Sub(startTime))
		delete(lo.startTimes, event.FilePath) // Clean up
	}

	if event.HasErrors {
		lo.logger.Info("⚠️  Completed with errors: %s → %s, parsed: %d%s",
			event.FilePath, event.Collection, event.ParsedCount, duration)
	} else if lo.logLevel >= LogLevelInfo {
		lo.logger.Info("✅ Completed successfully: %s → %s, parsed: %d%s",
			event.FilePath, event.Collection, event.ParsedCount, duration)
	}
}

// logPipelineFailed logs pipeline failed events
func (lo *LoggingObserver) logPipelineFailed(event *events.PipelineFailedEvent) {
	duration := ""
	if startTime, exists := lo.startTimes[event.FilePath]; exists {
		duration = fmt.Sprintf(" (after %v)", event.Timestamp().Sub(startTime))
		delete(lo.startTimes, event.FilePath) // Clean up
	}

	lo.logger.Error("❌ Pipeline failed: %s at stage %s: %v%s",
		event.FilePath, event.Stage, event.Error, duration)
}

// logStageStarted logs stage started events
func (lo *LoggingObserver) logStageStarted(event *events.StageStartedEvent) {
	if lo.logLevel >= LogLevelDebug {
		lo.logger.Debug("🔄 Stage started: %s for %s", event.StageName, event.FilePath)
	}

	// Store stage start time
	stageKey := fmt.Sprintf("%s:%s", event.FilePath, event.StageName)
	lo.startTimes[stageKey] = event.Timestamp()
}

// logStageCompleted logs stage completed events
func (lo *LoggingObserver) logStageCompleted(event *events.StageCompletedEvent) {
	duration := ""
	stageKey := fmt.Sprintf("%s:%s", event.FilePath, event.StageName)
	if startTime, exists := lo.startTimes[stageKey]; exists {
		duration = fmt.Sprintf(" (took %v)", event.Timestamp().Sub(startTime))
		delete(lo.startTimes, stageKey) // Clean up
	}

	if lo.logLevel >= LogLevelDebug {
		lo.logger.Debug("✅ Stage completed: %s for %s%s", event.StageName, event.FilePath, duration)
	}
}

// logStageFailed logs stage failed events
func (lo *LoggingObserver) logStageFailed(event *events.StageFailedEvent) {
	duration := ""
	stageKey := fmt.Sprintf("%s:%s", event.FilePath, event.StageName)
	if startTime, exists := lo.startTimes[stageKey]; exists {
		duration = fmt.Sprintf(" (after %v)", event.Timestamp().Sub(startTime))
		delete(lo.startTimes, stageKey) // Clean up
	}

	lo.logger.Error("❌ Stage failed: %s for %s: %v%s",
		event.StageName, event.FilePath, event.Error, duration)
}

// logFileProcessingStarted logs file processing started events
func (lo *LoggingObserver) logFileProcessingStarted(event *events.FileProcessingStartedEvent) {
	if lo.logLevel >= LogLevelDebug {
		lo.logger.Debug("📖 File processing started: %s → %s", event.FilePath, event.Collection)
	}
}

// logFileProcessingCompleted logs file processing completed events
func (lo *LoggingObserver) logFileProcessingCompleted(event *events.FileProcessingCompletedEvent) {
	if lo.logLevel >= LogLevelInfo {
		lo.logger.Info("📝 File processed: %s → %s, parsed: %d, written: %d",
			event.FilePath, event.Collection, event.ParsedCount, event.WrittenCount)
	}
}

// logParsingError logs parsing error events
func (lo *LoggingObserver) logParsingError(event *events.ParsingErrorEvent) {
	lineInfo := ""
	if event.LineNumber > 0 {
		lineInfo = fmt.Sprintf(" (line %d)", event.LineNumber)
	}

	lo.logger.Error("🔍 Parsing error in %s%s: %v", event.FilePath, lineInfo, event.Error)
}

// logValidationError logs validation error events
func (lo *LoggingObserver) logValidationError(event *events.ValidationErrorEvent) {
	entityInfo := ""
	if event.EntityType != "" {
		entityInfo = fmt.Sprintf(" (entity: %s)", event.EntityType)
	}

	lo.logger.Error("✓ Validation error in %s%s: %v", event.FilePath, entityInfo, event.Error)
}

// logPersistenceError logs persistence error events
func (lo *LoggingObserver) logPersistenceError(event *events.PersistenceErrorEvent) {
	lo.logger.Error("💾 Persistence error in %s → %s (%d entities): %v",
		event.FilePath, event.Collection, event.EntityCount, event.Error)
}

// logProgress logs progress events
func (lo *LoggingObserver) logProgress(event *events.ProgressEvent) {
	if lo.logLevel >= LogLevelInfo {
		lo.logger.Info("📊 Progress: %d/%d files (%.1f%%) - current: %s",
			event.ProcessedFiles, event.TotalFiles, event.ProgressPercent, event.CurrentFile)
	}
}

// logProcessingSummary logs processing summary events
func (lo *LoggingObserver) logProcessingSummary(event *events.ProcessingSummaryEvent) {
	lo.logger.Info("📋 Processing Summary:")
	lo.logger.Info("  Duration: %v", event.Duration)
	lo.logger.Info("  Total files: %d", event.TotalFiles)
	lo.logger.Info("  Successful: %d", event.SuccessfulFiles)
	lo.logger.Info("  Failed: %d", event.FailedFiles)
	lo.logger.Info("  Total parsed: %d", event.TotalParsed)
	lo.logger.Info("  Total written: %d", event.TotalWritten)
	lo.logger.Info("  Total errors: %d", event.TotalErrors)

	if event.FailedFiles == 0 {
		lo.logger.Info("🎉 All files processed successfully!")
	} else if event.SuccessfulFiles > 0 {
		lo.logger.Info("⚠️  Processing completed with some failures")
	} else {
		lo.logger.Error("❌ All files failed to process")
	}
}

// SetLogLevel changes the logging level
func (lo *LoggingObserver) SetLogLevel(level LogLevel) {
	lo.logLevel = level
}

// GetLogLevel returns the current logging level
func (lo *LoggingObserver) GetLogLevel() LogLevel {
	return lo.logLevel
}
