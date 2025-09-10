package stages

import (
	"context"
	"encoding/json"
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

	// Handle case where transformation stage was skipped
	var documents []map[string]any
	if len(data.Documents) > 0 {
		documents = data.Documents
	} else if len(data.ParsedData) > 0 {
		// Convert parsed entities directly to documents with separated structure
		for _, entity := range data.ParsedData {
			// Convert entity to map first
			entityBytes, err := json.Marshal(entity)
			if err != nil {
				s.logger.Error("failed to marshal entity: %v", err)
				continue
			}
			
			var entityMap map[string]any
			if err := json.Unmarshal(entityBytes, &entityMap); err != nil {
				s.logger.Error("failed to unmarshal entity: %v", err)
				continue
			}
			
			// Extract contenuto for metadata level
			contenuto := ""
			if contenutoValue, exists := entityMap["contenuto"]; exists {
				if contenutoStr, ok := contenutoValue.(string); ok {
					contenuto = contenutoStr
				}
				// Remove contenuto from domain data
				delete(entityMap, "contenuto")
			}
			
			// Create document with separated structure
			doc := map[string]any{
				// Metadata at root level
				"collection":  data.WorkItem.Collection,
				"source_file": data.FilePath,
				"language":    "ita", 
				"created_at":  time.Now(),
				"contenuto":   contenuto,
				
				// All domain data in value object
				"value": entityMap,
			}
			
			documents = append(documents, doc)
		}
		s.logger.Debug("converted %d parsed entities to documents for %s", len(documents), data.FilePath)
	} else {
		s.logger.Debug("no documents or parsed data to persist for %s", data.FilePath)
		data.Metadata["written_count"] = 0
		return nil
	}

	collection := data.WorkItem.Collection
	documentCount := len(documents)

	s.logger.Debug("persisting %d documents to collection %s (dry_run: %v)", documentCount, collection, s.dryRun)

	if s.dryRun {
		// In dry run mode, just log what would be done
		s.logger.Info("dry run: would persist %d documents to %s", documentCount, collection)
		data.Metadata["written_count"] = 0
		data.Metadata["dry_run"] = true

		// Generate preview for dry run
		preview := s.generatePreview(documents)
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
	writtenCount, err := s.repositoryWrapper.UpsertMany(collection, uniqueFields, documents)
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
	maxPreview := min(len(documents), 5)

	for i := range maxPreview {
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
