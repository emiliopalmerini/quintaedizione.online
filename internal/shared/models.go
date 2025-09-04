package shared

import (
	"time"
)

// BaseEntity represents common fields for all entities
type BaseEntity struct {
	ID         string    `json:"id" bson:"_id"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
	Version    string    `json:"version" bson:"version"`
	Source     string    `json:"source" bson:"source"`
}

// NewBaseEntity creates a new base entity with defaults
func NewBaseEntity(id string) BaseEntity {
	now := time.Now()
	return BaseEntity{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   "1.0",
		Source:    "SRD",
	}
}

// UpdateTimestamp updates the UpdatedAt field
func (b *BaseEntity) UpdateTimestamp() {
	b.UpdatedAt = time.Now()
}

// MarkdownContent represents content with markdown formatting
type MarkdownContent struct {
	Raw       string `json:"raw" bson:"raw"`             // Original markdown
	HTML      string `json:"html" bson:"html"`           // Rendered HTML
	PlainText string `json:"plain_text" bson:"plain_text"` // Plain text version
}

// NewMarkdownContent creates markdown content with just the raw content
func NewMarkdownContent(raw string) MarkdownContent {
	return MarkdownContent{
		Raw:       raw,
		HTML:      "", // To be rendered later
		PlainText: "", // To be extracted later
	}
}

// SearchableContent represents content optimized for search
type SearchableContent struct {
	Title       string   `json:"title" bson:"title"`
	Description string   `json:"description" bson:"description"`
	Keywords    []string `json:"keywords" bson:"keywords"`
	Category    string   `json:"category" bson:"category"`
	Tags        []string `json:"tags" bson:"tags"`
}

// AddKeyword adds a keyword if not already present
func (s *SearchableContent) AddKeyword(keyword string) {
	for _, k := range s.Keywords {
		if k == keyword {
			return
		}
	}
	s.Keywords = append(s.Keywords, keyword)
}

// AddTag adds a tag if not already present
func (s *SearchableContent) AddTag(tag string) {
	for _, t := range s.Tags {
		if t == tag {
			return
		}
	}
	s.Tags = append(s.Tags, tag)
}

// ValidationError represents a domain validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Error implements the error interface
func (v ValidationError) Error() string {
	return v.Message
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors"`
}

// NewValidationResult creates a new validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:  true,
		Errors: make([]ValidationError, 0),
	}
}

// AddError adds a validation error
func (v *ValidationResult) AddError(field, message, code string) {
	v.Valid = false
	v.Errors = append(v.Errors, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	})
}

// HasErrors returns true if there are validation errors
func (v *ValidationResult) HasErrors() bool {
	return len(v.Errors) > 0
}

// Repository interface patterns

// Repository represents the base repository interface
type Repository[T any] interface {
	FindByID(id string) (*T, error)
	Save(entity *T) error
	Delete(id string) error
	FindAll() ([]*T, error)
}

// QueryRepository represents read-only repository operations
type QueryRepository[T any] interface {
	FindByID(id string) (*T, error)
	FindAll() ([]*T, error)
	Search(query SearchQuery) ([]*T, error)
	Count(query SearchQuery) (int64, error)
}

// SearchQuery represents a generic search query
type SearchQuery struct {
	Text     string            `json:"text"`
	Filters  map[string]string `json:"filters"`
	Sort     string            `json:"sort"`
	SortDesc bool              `json:"sort_desc"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

// NewSearchQuery creates a new search query with defaults
func NewSearchQuery() *SearchQuery {
	return &SearchQuery{
		Filters:  make(map[string]string),
		Sort:     "name",
		SortDesc: false,
		Limit:    50,
		Offset:   0,
	}
}

// AddFilter adds a filter to the search query
func (s *SearchQuery) AddFilter(key, value string) {
	s.Filters[key] = value
}

// RemoveFilter removes a filter from the search query
func (s *SearchQuery) RemoveFilter(key string) {
	delete(s.Filters, key)
}

// HasFilter checks if a filter exists
func (s *SearchQuery) HasFilter(key string) bool {
	_, exists := s.Filters[key]
	return exists
}

// GetFilter gets a filter value
func (s *SearchQuery) GetFilter(key string) (string, bool) {
	value, exists := s.Filters[key]
	return value, exists
}

// Domain Event patterns

// DomainEvent represents a domain event
type DomainEvent interface {
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
	Version() int
}

// BaseDomainEvent provides common domain event functionality
type BaseDomainEvent struct {
	EventName   string    `json:"event_name"`
	AggregateId string    `json:"aggregate_id"`
	Occurred    time.Time `json:"occurred_at"`
	EventVersion int      `json:"version"`
}

// EventType returns the event type
func (b BaseDomainEvent) EventType() string {
	return b.EventName
}

// AggregateID returns the aggregate ID
func (b BaseDomainEvent) AggregateID() string {
	return b.AggregateId
}

// OccurredAt returns when the event occurred
func (b BaseDomainEvent) OccurredAt() time.Time {
	return b.Occurred
}

// Version returns the event version
func (b BaseDomainEvent) Version() int {
	return b.EventVersion
}

// EventStore interface for persisting domain events
type EventStore interface {
	SaveEvent(event DomainEvent) error
	GetEvents(aggregateID string) ([]DomainEvent, error)
	GetEventsByType(eventType string) ([]DomainEvent, error)
}

// EventPublisher interface for publishing domain events
type EventPublisher interface {
	Publish(event DomainEvent) error
	Subscribe(eventType string, handler EventHandler) error
}

// EventHandler handles domain events
type EventHandler interface {
	Handle(event DomainEvent) error
}

// Common use case interfaces

// UseCase represents a use case interface
type UseCase[TRequest, TResponse any] interface {
	Execute(request TRequest) (TResponse, error)
}

// Command represents a command in CQRS
type Command interface {
	CommandType() string
}

// Query represents a query in CQRS
type Query interface {
	QueryType() string
}

// CommandHandler handles commands
type CommandHandler[TCommand Command, TResult any] interface {
	Handle(command TCommand) (TResult, error)
}

// QueryHandler handles queries
type QueryHandler[TQuery Query, TResult any] interface {
	Handle(query TQuery) (TResult, error)
}