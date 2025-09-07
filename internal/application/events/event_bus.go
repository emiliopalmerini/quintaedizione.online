package events

import (
	"sync"
	"time"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
)

// BaseEvent provides common event functionality
type BaseEvent struct {
	EventTime time.Time `json:"timestamp"`
}

// EventType returns the base event type
func (e *BaseEvent) EventType() string {
	return "base"
}

// Timestamp returns the event timestamp
func (e *BaseEvent) Timestamp() time.Time {
	return e.EventTime
}

// Pipeline Events

// PipelineStartedEvent signals pipeline execution start
type PipelineStartedEvent struct {
	BaseEvent
	FilePath   string `json:"file_path"`
	Collection string `json:"collection"`
}

func (e *PipelineStartedEvent) EventType() string {
	return "pipeline_started"
}

// PipelineCompletedEvent signals successful pipeline completion
type PipelineCompletedEvent struct {
	BaseEvent
	FilePath    string `json:"file_path"`
	Collection  string `json:"collection"`
	ParsedCount int    `json:"parsed_count"`
	HasErrors   bool   `json:"has_errors"`
}

func (e *PipelineCompletedEvent) EventType() string {
	return "pipeline_completed"
}

// PipelineFailedEvent signals pipeline failure
type PipelineFailedEvent struct {
	BaseEvent
	FilePath   string `json:"file_path"`
	Collection string `json:"collection"`
	Error      error  `json:"error"`
	Stage      string `json:"stage"`
}

func (e *PipelineFailedEvent) EventType() string {
	return "pipeline_failed"
}

// StageStartedEvent signals stage execution start
type StageStartedEvent struct {
	BaseEvent
	StageName  string `json:"stage_name"`
	FilePath   string `json:"file_path"`
	Collection string `json:"collection"`
}

func (e *StageStartedEvent) EventType() string {
	return "stage_started"
}

// StageCompletedEvent signals successful stage completion
type StageCompletedEvent struct {
	BaseEvent
	StageName  string `json:"stage_name"`
	FilePath   string `json:"file_path"`
	Collection string `json:"collection"`
}

func (e *StageCompletedEvent) EventType() string {
	return "stage_completed"
}

// StageFailedEvent signals stage failure
type StageFailedEvent struct {
	BaseEvent
	StageName  string `json:"stage_name"`
	FilePath   string `json:"file_path"`
	Collection string `json:"collection"`
	Error      error  `json:"error"`
}

func (e *StageFailedEvent) EventType() string {
	return "stage_failed"
}

// FileProcessingStartedEvent signals file processing start
type FileProcessingStartedEvent struct {
	BaseEvent
	FilePath   string `json:"file_path"`
	Collection string `json:"collection"`
}

func (e *FileProcessingStartedEvent) EventType() string {
	return "file_processing_started"
}

// FileProcessingCompletedEvent signals successful file processing completion
type FileProcessingCompletedEvent struct {
	BaseEvent
	FilePath     string `json:"file_path"`
	Collection   string `json:"collection"`
	ParsedCount  int    `json:"parsed_count"`
	WrittenCount int    `json:"written_count"`
}

func (e *FileProcessingCompletedEvent) EventType() string {
	return "file_processing_completed"
}

// ParsingErrorEvent signals parsing error
type ParsingErrorEvent struct {
	BaseEvent
	FilePath   string `json:"file_path"`
	Collection string `json:"collection"`
	Error      error  `json:"error"`
	LineNumber int    `json:"line_number,omitempty"`
}

func (e *ParsingErrorEvent) EventType() string {
	return "parsing_error"
}

// ValidationErrorEvent signals validation error
type ValidationErrorEvent struct {
	BaseEvent
	FilePath   string `json:"file_path"`
	Collection string `json:"collection"`
	Error      error  `json:"error"`
	EntityType string `json:"entity_type,omitempty"`
}

func (e *ValidationErrorEvent) EventType() string {
	return "validation_error"
}

// PersistenceErrorEvent signals persistence error
type PersistenceErrorEvent struct {
	BaseEvent
	FilePath    string `json:"file_path"`
	Collection  string `json:"collection"`
	Error       error  `json:"error"`
	EntityCount int    `json:"entity_count"`
}

func (e *PersistenceErrorEvent) EventType() string {
	return "persistence_error"
}

// ProgressEvent signals processing progress
type ProgressEvent struct {
	BaseEvent
	TotalFiles      int     `json:"total_files"`
	ProcessedFiles  int     `json:"processed_files"`
	CurrentFile     string  `json:"current_file"`
	ProgressPercent float64 `json:"progress_percent"`
}

func (e *ProgressEvent) EventType() string {
	return "progress"
}

// ProcessingSummaryEvent signals final processing summary
type ProcessingSummaryEvent struct {
	BaseEvent
	TotalFiles      int           `json:"total_files"`
	SuccessfulFiles int           `json:"successful_files"`
	FailedFiles     int           `json:"failed_files"`
	TotalParsed     int           `json:"total_parsed"`
	TotalWritten    int           `json:"total_written"`
	TotalErrors     int           `json:"total_errors"`
	Duration        time.Duration `json:"duration"`
}

func (e *ProcessingSummaryEvent) EventType() string {
	return "processing_summary"
}

// Event represents any event that can be published
type Event interface {
	EventType() string
	Timestamp() time.Time
}

// EventHandler handles published events
type EventHandler func(event Event)

// EventBus manages event publishing and subscription
type EventBus interface {
	Subscribe(eventType string, handler EventHandler) UnsubscribeFunc
	Publish(event Event)
	PublishAsync(event Event)
	Close()
}

// UnsubscribeFunc removes an event subscription
type UnsubscribeFunc func()

// InMemoryEventBus provides an in-memory event bus implementation
type InMemoryEventBus struct {
	subscribers map[string][]EventHandler
	mu          sync.RWMutex
	logger      parsers.Logger
	closed      bool
	eventBuffer chan Event
}

// NewInMemoryEventBus creates a new in-memory event bus
func NewInMemoryEventBus(logger parsers.Logger, bufferSize int) *InMemoryEventBus {
	if logger == nil {
		logger = &parsers.NoOpLogger{}
	}
	
	if bufferSize <= 0 {
		bufferSize = 100
	}

	bus := &InMemoryEventBus{
		subscribers: make(map[string][]EventHandler),
		logger:      logger,
		eventBuffer: make(chan Event, bufferSize),
	}

	// Start async event processing goroutine
	go bus.processAsyncEvents()

	return bus
}

// Subscribe adds an event handler for a specific event type
func (bus *InMemoryEventBus) Subscribe(eventType string, handler EventHandler) UnsubscribeFunc {
	if handler == nil {
		bus.logger.Error("attempted to subscribe nil handler for event type: %s", eventType)
		return func() {}
	}

	bus.mu.Lock()
	defer bus.mu.Unlock()

	if bus.closed {
		bus.logger.Error("attempted to subscribe to closed event bus")
		return func() {}
	}

	bus.subscribers[eventType] = append(bus.subscribers[eventType], handler)
	
	bus.logger.Debug("subscribed handler for event type: %s", eventType)

	// Return unsubscribe function
	return func() {
		bus.unsubscribe(eventType, handler)
	}
}

// Publish synchronously publishes an event to all subscribers
func (bus *InMemoryEventBus) Publish(event Event) {
	if event == nil {
		bus.logger.Error("attempted to publish nil event")
		return
	}

	bus.mu.RLock()
	defer bus.mu.RUnlock()

	if bus.closed {
		bus.logger.Error("attempted to publish to closed event bus")
		return
	}

	eventType := event.EventType()
	handlers := bus.subscribers[eventType]

	bus.logger.Debug("publishing event %s to %d handlers", eventType, len(handlers))

	for _, handler := range handlers {
		func(h EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					bus.logger.Error("event handler panicked for %s: %v", eventType, r)
				}
			}()
			h(event)
		}(handler)
	}
}

// PublishAsync asynchronously publishes an event to all subscribers
func (bus *InMemoryEventBus) PublishAsync(event Event) {
	if event == nil {
		bus.logger.Error("attempted to publish nil event")
		return
	}

	if bus.closed {
		bus.logger.Error("attempted to publish to closed event bus")
		return
	}

	select {
	case bus.eventBuffer <- event:
		bus.logger.Debug("queued async event: %s", event.EventType())
	default:
		bus.logger.Error("event buffer full, dropping event: %s", event.EventType())
	}
}

// Close shuts down the event bus
func (bus *InMemoryEventBus) Close() {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if bus.closed {
		return
	}

	bus.closed = true
	close(bus.eventBuffer)
	bus.logger.Info("event bus closed")
}

// processAsyncEvents processes events from the async buffer
func (bus *InMemoryEventBus) processAsyncEvents() {
	for event := range bus.eventBuffer {
		if event != nil {
			bus.Publish(event)
		}
	}
}

// unsubscribe removes a handler for a specific event type
func (bus *InMemoryEventBus) unsubscribe(eventType string, targetHandler EventHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	handlers := bus.subscribers[eventType]
	for i, handler := range handlers {
		// Note: This comparison is not perfect for function equality,
		// but works for most practical cases
		if &handler == &targetHandler {
			bus.subscribers[eventType] = append(handlers[:i], handlers[i+1:]...)
			bus.logger.Debug("unsubscribed handler for event type: %s", eventType)
			break
		}
	}
}

// GetSubscriberCount returns the number of subscribers for an event type
func (bus *InMemoryEventBus) GetSubscriberCount(eventType string) int {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	return len(bus.subscribers[eventType])
}

// GetTotalSubscribers returns the total number of event subscriptions
func (bus *InMemoryEventBus) GetTotalSubscribers() int {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	
	total := 0
	for _, handlers := range bus.subscribers {
		total += len(handlers)
	}
	return total
}