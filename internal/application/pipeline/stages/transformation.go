package stages

import (
	"context"
	"fmt"
	"time"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/events"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/pipeline"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// TransformationStage transforms parsed entities using the document builder
type TransformationStage struct {
	builder  *parsers.DocumentBuilder
	eventBus events.EventBus
	logger   parsers.Logger
}

// NewTransformationStage creates a new transformation stage
func NewTransformationStage(builder *parsers.DocumentBuilder, eventBus events.EventBus, logger parsers.Logger) *TransformationStage {
	if logger == nil {
		logger = &parsers.NoOpLogger{}
	}

	return &TransformationStage{
		builder:  builder,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Name returns the stage name
func (s *TransformationStage) Name() string {
	return "transformation"
}

// Process transforms parsed entities into documents suitable for persistence
func (s *TransformationStage) Process(ctx context.Context, data *pipeline.ProcessingData) error {
	if data == nil {
		return fmt.Errorf("processing data cannot be nil")
	}

	if len(data.ParsedData) == 0 {
		s.logger.Debug("no parsed entities to transform for %s", data.FilePath)
		data.Documents = []map[string]any{}
		return nil
	}

	s.logger.Debug("transforming %d parsed entities from %s", len(data.ParsedData), data.FilePath)

	var documents []map[string]any
	var transformationErrors []error

	// Transform each parsed entity into a document
	for i, entity := range data.ParsedData {
		doc, err := s.transformEntity(entity, data, i)
		if err != nil {
			transformationErrors = append(transformationErrors, err)
			s.logger.Error("transformation failed for entity %d in %s: %v", i, data.FilePath, err)
		} else {
			documents = append(documents, doc)
		}
	}

	// Store transformed documents
	data.Documents = documents

	// Store transformation metadata
	data.Metadata["transformed_count"] = len(documents)
	data.Metadata["transformation_errors"] = len(transformationErrors)

	if len(transformationErrors) > 0 {
		s.logger.Info("transformation completed for %s: %d successful, %d failed", 
			data.FilePath, len(documents), len(transformationErrors))
		
		// Add errors to processing data but continue
		data.Errors = append(data.Errors, transformationErrors...)
	} else {
		s.logger.Debug("all %d entities transformed successfully for %s", len(documents), data.FilePath)
	}

	return nil
}

// transformEntity transforms a single parsed entity into a document
func (s *TransformationStage) transformEntity(entity domain.ParsedEntity, data *pipeline.ProcessingData, index int) (map[string]any, error) {
	if entity == nil {
		return nil, fmt.Errorf("entity at index %d is nil", index)
	}

	// If we have a document builder, use it for transformation
	if s.builder != nil {
		// Use document builder to create the document
		doc, err := s.builder.BuildDocument(entity, data.WorkItem.Collection)
		if err != nil {
			return nil, fmt.Errorf("document builder failed for entity at index %d: %w", index, err)
		}
		return doc, nil
	}

	// Fallback: create a basic document structure
	doc := map[string]any{
		"entity_type": entity.EntityType(),
		"entity":      entity,
		"collection":  data.WorkItem.Collection,
		"source_file": data.FilePath,
		"created_at":  time.Now(),
	}

	// Add common metadata fields
	if data.Metadata != nil {
		doc["language"] = data.Metadata["language"] 
		if doc["language"] == nil {
			doc["language"] = "ita" // default to Italian
		}
	}

	return doc, nil
}