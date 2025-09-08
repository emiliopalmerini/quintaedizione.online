package stages

import (
	"context"
	"fmt"
	"time"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/repositories"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/events"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/pipeline"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/infrastructure/mongodb"
)

// PersistenceStage persists transformed documents to the repository
type PersistenceStage struct {
	repositoryWrapper *repositories.ParserRepositoryWrapper
	dryRun            bool
	eventBus          events.EventBus
	logger            parsers.Logger
}

// NewPersistenceStage creates a new persistence stage
func NewPersistenceStage(repositoryWrapper *repositories.ParserRepositoryWrapper, dryRun bool, eventBus events.EventBus, logger parsers.Logger) *PersistenceStage {
	if logger == nil {
		logger = &parsers.NoOpLogger{}
	}

	return &PersistenceStage{
		repositoryWrapper: repositoryWrapper,
		dryRun:            dryRun,
		eventBus:          eventBus,
		logger:            logger,
	}
}

// Name returns the stage name
func (s *PersistenceStage) Name() string {
	return "persistence"
}

// Process persists the transformed documents to the repository
func (s *PersistenceStage) Process(ctx context.Context, data *pipeline.ProcessingData) error {
	if data == nil {
		return fmt.Errorf("processing data cannot be nil")
	}

	if len(data.Documents) == 0 {
		s.logger.Debug("no documents to persist for %s", data.FilePath)
		data.Metadata["written_count"] = 0
		return nil
	}

	collection := data.WorkItem.Collection
	documentCount := len(data.Documents)

	s.logger.Debug("persisting %d documents to collection %s (dry_run: %v)", documentCount, collection, s.dryRun)

	if s.dryRun {
		// In dry run mode, just log what would be done
		s.logger.Info("dry run: would persist %d documents to %s", documentCount, collection)
		data.Metadata["written_count"] = 0
		data.Metadata["dry_run"] = true

		// Generate preview for dry run
		preview := s.generatePreview(data.Documents)
		data.Metadata["preview"] = preview

		return nil
	}

	if s.repositoryWrapper == nil {
		err := fmt.Errorf("repository not available for persistence")
		s.logger.Error("repository not available for persistence")

		if s.eventBus != nil {
			s.eventBus.Publish(&events.PersistenceErrorEvent{
				BaseEvent:   events.BaseEvent{EventTime: time.Now()},
				FilePath:    data.FilePath,
				Collection:  collection,
				Error:       err,
				EntityCount: documentCount,
			})
		}

		return err
	}

	// Get unique fields for this collection
	uniqueFields := mongodb.GetUniqueFieldsForCollection(collection)

	// Persist documents using bulk upsert
	writtenCount, err := s.repositoryWrapper.UpsertMany(collection, uniqueFields, data.Documents)
	if err != nil {
		s.logger.Error("failed to persist %d documents to %s: %v", documentCount, collection, err)

		if s.eventBus != nil {
			s.eventBus.Publish(&events.PersistenceErrorEvent{
				BaseEvent:   events.BaseEvent{EventTime: time.Now()},
				FilePath:    data.FilePath,
				Collection:  collection,
				Error:       err,
				EntityCount: documentCount,
			})
		}

		return fmt.Errorf("failed to persist documents to %s: %w", collection, err)
	}

	// Store persistence metadata
	data.Metadata["written_count"] = writtenCount
	data.Metadata["dry_run"] = false

	s.logger.Debug("successfully persisted %d documents to %s (written: %d)", documentCount, collection, writtenCount)

	// Publish file processing completed event
	if s.eventBus != nil {
		s.eventBus.Publish(&events.FileProcessingCompletedEvent{
			BaseEvent:    events.BaseEvent{EventTime: time.Now()},
			FilePath:     data.FilePath,
			Collection:   collection,
			ParsedCount:  len(data.ParsedData),
			WrittenCount: writtenCount,
		})
	}

	return nil
}

// generatePreview generates a preview of documents for dry run mode
func (s *PersistenceStage) generatePreview(documents []map[string]any) []map[string]any {
	previewKeys := []string{"name", "term", "level", "rarity", "type", "school", "nome", "titolo", "entity_type"}

	var preview []map[string]any
	maxPreview := 5
	if len(documents) < maxPreview {
		maxPreview = len(documents)
	}

	for i := 0; i < maxPreview; i++ {
		doc := documents[i]
		previewDoc := make(map[string]any)

		for _, key := range previewKeys {
			if value, exists := doc[key]; exists {
				previewDoc[key] = value
			}
		}

		// Also try to extract fields from nested entity
		if entity, exists := doc["entity"]; exists && entity != nil {
			if entityMap, ok := entity.(map[string]any); ok {
				for _, key := range previewKeys {
					if value, exists := entityMap[key]; exists && previewDoc[key] == nil {
						previewDoc[key] = value
					}
				}
			}
		}

		if len(previewDoc) > 0 {
			preview = append(preview, previewDoc)
		}
	}

	return preview
}
