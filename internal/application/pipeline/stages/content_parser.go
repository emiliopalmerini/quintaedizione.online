package stages

import (
	"context"
	"fmt"
	"time"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/events"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/pipeline"
)

// ContentParserStage parses file content using the registry system
type ContentParserStage struct {
	registry *parsers.ParserRegistry
	eventBus events.EventBus
	logger   parsers.Logger
}

// NewContentParserStage creates a new content parser stage
func NewContentParserStage(registry *parsers.ParserRegistry, eventBus events.EventBus, logger parsers.Logger) *ContentParserStage {
	if logger == nil {
		logger = &parsers.NoOpLogger{}
	}

	return &ContentParserStage{
		registry: registry,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Name returns the stage name
func (s *ContentParserStage) Name() string {
	return "content_parser"
}

// Process parses the raw content using the appropriate strategy from the registry
func (s *ContentParserStage) Process(ctx context.Context, data *pipeline.ProcessingData) error {
	if data == nil {
		return fmt.Errorf("processing data cannot be nil")
	}

	if data.WorkItem == nil {
		return fmt.Errorf("work item cannot be nil")
	}

	if len(data.RawContent) == 0 {
		err := fmt.Errorf("no raw content available for parsing")
		s.logger.Error("no raw content available for %s", data.FilePath)

		if s.eventBus != nil {
			s.eventBus.Publish(&events.ParsingErrorEvent{
				BaseEvent:  events.BaseEvent{EventTime: time.Now()},
				FilePath:   data.FilePath,
				Collection: data.WorkItem.Collection,
				Error:      err,
			})
		}

		return err
	}

	// Determine content type from collection name using existing function
	contentType, err := parsers.GetContentTypeFromCollection(data.WorkItem.Collection)
	if err != nil {
		s.logger.Error("unknown content type for collection: %s", data.WorkItem.Collection)

		if s.eventBus != nil {
			s.eventBus.Publish(&events.ParsingErrorEvent{
				BaseEvent:  events.BaseEvent{EventTime: time.Now()},
				FilePath:   data.FilePath,
				Collection: data.WorkItem.Collection,
				Error:      err,
			})
		}

		return err
	}

	s.logger.Debug("parsing content for collection: %s, content type: %s", data.WorkItem.Collection, contentType)

	// Get parsing strategy from registry
	strategy, err := s.registry.GetParser(contentType)
	if err != nil {
		s.logger.Error("no parser found for content type %s: %v", contentType, err)

		if s.eventBus != nil {
			s.eventBus.Publish(&events.ParsingErrorEvent{
				BaseEvent:  events.BaseEvent{EventTime: time.Now()},
				FilePath:   data.FilePath,
				Collection: data.WorkItem.Collection,
				Error:      err,
			})
		}

		return fmt.Errorf("no parser found for content type %s: %w", contentType, err)
	}

	// Create parsing context
	context := parsers.NewParsingContext(data.FilePath, "ita").
		WithMetadata("collection", data.WorkItem.Collection)

	// Parse using strategy
	parsedEntities, err := strategy.Parse(data.RawContent, context)
	if err != nil {
		s.logger.Error("strategy parser failed for %s: %v", data.FilePath, err)

		if s.eventBus != nil {
			s.eventBus.Publish(&events.ParsingErrorEvent{
				BaseEvent:  events.BaseEvent{EventTime: time.Now()},
				FilePath:   data.FilePath,
				Collection: data.WorkItem.Collection,
				Error:      err,
			})
		}

		return fmt.Errorf("failed to parse %s: %w", data.FilePath, err)
	}

	// Store parsed data
	data.ParsedData = parsedEntities
	data.ContentType = contentType

	s.logger.Debug("successfully parsed %d entities from %s", len(parsedEntities), data.FilePath)

	// Store parsing metadata
	data.Metadata["parsed_count"] = len(parsedEntities)
	data.Metadata["parser_type"] = "strategy"
	data.Metadata["content_type"] = string(contentType)

	return nil
}
