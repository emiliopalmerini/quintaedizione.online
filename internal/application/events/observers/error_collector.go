package observers

import (
	"sync"
	"time"

	"github.com/emiliopalmerini/quintaedizione.online/internal/application/events"
	"github.com/emiliopalmerini/quintaedizione.online/internal/application/parsers"
)

type ErrorCollector struct {
	errors   []ErrorInfo
	mu       sync.RWMutex
	logger   parsers.Logger
	eventBus events.EventBus

	totalErrors       int
	parsingErrors     int
	validationErrors  int
	persistenceErrors int
	otherErrors       int
}

type ErrorInfo struct {
	Timestamp  time.Time `json:"timestamp"`
	FilePath   string    `json:"file_path"`
	Collection string    `json:"collection"`
	ErrorType  string    `json:"error_type"`
	Message    string    `json:"message"`
	Stage      string    `json:"stage,omitempty"`
	EntityType string    `json:"entity_type,omitempty"`
	LineNumber int       `json:"line_number,omitempty"`
}

func NewErrorCollector(eventBus events.EventBus, logger parsers.Logger) *ErrorCollector {
	if logger == nil {
		logger = &parsers.NoOpLogger{}
	}

	return &ErrorCollector{
		errors:   make([]ErrorInfo, 0),
		logger:   logger,
		eventBus: eventBus,
	}
}

func (ec *ErrorCollector) HandleEvent(event events.Event) {
	switch e := event.(type) {
	case *events.ParsingErrorEvent:
		ec.handleParsingError(e)
	case *events.ValidationErrorEvent:
		ec.handleValidationError(e)
	case *events.PersistenceErrorEvent:
		ec.handlePersistenceError(e)
	case *events.PipelineFailedEvent:
		ec.handlePipelineFailed(e)
	case *events.StageFailedEvent:
		ec.handleStageFailed(e)
	}
}

func (ec *ErrorCollector) handleParsingError(event *events.ParsingErrorEvent) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	errorInfo := ErrorInfo{
		Timestamp:  event.Timestamp(),
		FilePath:   event.FilePath,
		Collection: event.Collection,
		ErrorType:  "parsing",
		Message:    event.Error.Error(),
		LineNumber: event.LineNumber,
	}

	ec.errors = append(ec.errors, errorInfo)
	ec.totalErrors++
	ec.parsingErrors++

	ec.logger.Error("parsing error in %s: %s", event.FilePath, event.Error.Error())
}

func (ec *ErrorCollector) handleValidationError(event *events.ValidationErrorEvent) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	errorInfo := ErrorInfo{
		Timestamp:  event.Timestamp(),
		FilePath:   event.FilePath,
		Collection: event.Collection,
		ErrorType:  "validation",
		Message:    event.Error.Error(),
		EntityType: event.EntityType,
	}

	ec.errors = append(ec.errors, errorInfo)
	ec.totalErrors++
	ec.validationErrors++

	ec.logger.Error("validation error in %s: %s", event.FilePath, event.Error.Error())
}

func (ec *ErrorCollector) handlePersistenceError(event *events.PersistenceErrorEvent) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	errorInfo := ErrorInfo{
		Timestamp:  event.Timestamp(),
		FilePath:   event.FilePath,
		Collection: event.Collection,
		ErrorType:  "persistence",
		Message:    event.Error.Error(),
	}

	ec.errors = append(ec.errors, errorInfo)
	ec.totalErrors++
	ec.persistenceErrors++

	ec.logger.Error("persistence error in %s: %s", event.FilePath, event.Error.Error())
}

func (ec *ErrorCollector) handlePipelineFailed(event *events.PipelineFailedEvent) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	errorInfo := ErrorInfo{
		Timestamp:  event.Timestamp(),
		FilePath:   event.FilePath,
		Collection: event.Collection,
		ErrorType:  "pipeline",
		Message:    event.Error.Error(),
		Stage:      event.Stage,
	}

	ec.errors = append(ec.errors, errorInfo)
	ec.totalErrors++
	ec.otherErrors++

	ec.logger.Error("pipeline failed for %s at stage %s: %s", event.FilePath, event.Stage, event.Error.Error())
}

func (ec *ErrorCollector) handleStageFailed(event *events.StageFailedEvent) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	errorInfo := ErrorInfo{
		Timestamp:  event.Timestamp(),
		FilePath:   event.FilePath,
		Collection: event.Collection,
		ErrorType:  "stage",
		Message:    event.Error.Error(),
		Stage:      event.StageName,
	}

	ec.errors = append(ec.errors, errorInfo)
	ec.totalErrors++
	ec.otherErrors++

	ec.logger.Error("stage %s failed for %s: %s", event.StageName, event.FilePath, event.Error.Error())
}

func (ec *ErrorCollector) GetErrors() []ErrorInfo {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	errorsCopy := make([]ErrorInfo, len(ec.errors))
	copy(errorsCopy, ec.errors)
	return errorsCopy
}

func (ec *ErrorCollector) GetErrorStatistics() ErrorStatistics {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	return ErrorStatistics{
		TotalErrors:       ec.totalErrors,
		ParsingErrors:     ec.parsingErrors,
		ValidationErrors:  ec.validationErrors,
		PersistenceErrors: ec.persistenceErrors,
		OtherErrors:       ec.otherErrors,
	}
}

func (ec *ErrorCollector) GetErrorsByType() map[string][]ErrorInfo {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	errorsByType := make(map[string][]ErrorInfo)

	for _, errorInfo := range ec.errors {
		errorsByType[errorInfo.ErrorType] = append(errorsByType[errorInfo.ErrorType], errorInfo)
	}

	return errorsByType
}

func (ec *ErrorCollector) GetErrorsByFile() map[string][]ErrorInfo {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	errorsByFile := make(map[string][]ErrorInfo)

	for _, errorInfo := range ec.errors {
		errorsByFile[errorInfo.FilePath] = append(errorsByFile[errorInfo.FilePath], errorInfo)
	}

	return errorsByFile
}

func (ec *ErrorCollector) GenerateErrorReport() ErrorReport {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	report := ErrorReport{
		Timestamp:    time.Now(),
		TotalErrors:  ec.totalErrors,
		Statistics:   ec.GetErrorStatistics(),
		ErrorsByType: make(map[string]int),
		ErrorsByFile: make(map[string]int),
		RecentErrors: make([]ErrorInfo, 0),
	}

	for _, errorInfo := range ec.errors {
		report.ErrorsByType[errorInfo.ErrorType]++
		report.ErrorsByFile[errorInfo.FilePath]++
	}

	recentCount := 10
	if len(ec.errors) < recentCount {
		recentCount = len(ec.errors)
	}

	if recentCount > 0 {
		startIndex := len(ec.errors) - recentCount
		report.RecentErrors = make([]ErrorInfo, recentCount)
		copy(report.RecentErrors, ec.errors[startIndex:])
	}

	return report
}

func (ec *ErrorCollector) Clear() {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.errors = make([]ErrorInfo, 0)
	ec.totalErrors = 0
	ec.parsingErrors = 0
	ec.validationErrors = 0
	ec.persistenceErrors = 0
	ec.otherErrors = 0

	ec.logger.Debug("error collector cleared")
}

type ErrorStatistics struct {
	TotalErrors       int `json:"total_errors"`
	ParsingErrors     int `json:"parsing_errors"`
	ValidationErrors  int `json:"validation_errors"`
	PersistenceErrors int `json:"persistence_errors"`
	OtherErrors       int `json:"other_errors"`
}

type ErrorReport struct {
	Timestamp    time.Time       `json:"timestamp"`
	TotalErrors  int             `json:"total_errors"`
	Statistics   ErrorStatistics `json:"statistics"`
	ErrorsByType map[string]int  `json:"errors_by_type"`
	ErrorsByFile map[string]int  `json:"errors_by_file"`
	RecentErrors []ErrorInfo     `json:"recent_errors"`
}
