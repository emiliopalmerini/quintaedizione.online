package stages

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/events"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/pipeline"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// ValidationStage validates parsed entities for data integrity
type ValidationStage struct {
	eventBus events.EventBus
	logger   parsers.Logger
}

// NewValidationStage creates a new validation stage
func NewValidationStage(eventBus events.EventBus, logger parsers.Logger) *ValidationStage {
	if logger == nil {
		logger = &parsers.NoOpLogger{}
	}

	return &ValidationStage{
		eventBus: eventBus,
		logger:   logger,
	}
}

// Name returns the stage name
func (s *ValidationStage) Name() string {
	return "validation"
}

// Process validates the parsed entities for data integrity
func (s *ValidationStage) Process(ctx context.Context, data *pipeline.ProcessingData) error {
	if data == nil {
		return fmt.Errorf("processing data cannot be nil")
	}

	if len(data.ParsedData) == 0 {
		s.logger.Debug("no parsed entities to validate for %s", data.FilePath)
		return nil
	}

	s.logger.Debug("validating %d parsed entities from %s", len(data.ParsedData), data.FilePath)

	var validationErrors []error
	validEntities := make([]domain.ParsedEntity, 0, len(data.ParsedData))

	// Validate each parsed entity
	for i, entity := range data.ParsedData {
		if err := s.validateEntity(entity, i); err != nil {
			validationErrors = append(validationErrors, err)
			s.logger.Error("validation failed for entity %d in %s: %v", i, data.FilePath, err)
			
			// Publish validation error event
			if s.eventBus != nil {
				s.eventBus.Publish(&events.ValidationErrorEvent{
					BaseEvent:  events.BaseEvent{EventTime: time.Now()},
					FilePath:   data.FilePath,
					Collection: data.WorkItem.Collection,
					Error:      err,
					EntityType: entity.EntityType(),
				})
			}
		} else {
			// Entity is valid, add to valid entities
			validEntities = append(validEntities, entity)
		}
	}

	// Update parsed data with only valid entities
	data.ParsedData = validEntities

	// Store validation metadata
	originalCount := len(data.ParsedData) + len(validationErrors)
	data.Metadata["validation_errors"] = len(validationErrors)
	data.Metadata["valid_entities"] = len(validEntities)
	data.Metadata["original_entities"] = originalCount

	if len(validationErrors) > 0 {
		s.logger.Info("validation completed for %s: %d valid, %d invalid entities", 
			data.FilePath, len(validEntities), len(validationErrors))
		
		// For now, we continue with valid entities rather than failing completely
		// This allows partial processing of files with some invalid entities
		data.Errors = append(data.Errors, validationErrors...)
	} else {
		s.logger.Debug("all %d entities validated successfully for %s", len(validEntities), data.FilePath)
	}

	return nil
}

// validateEntity validates a single parsed entity
func (s *ValidationStage) validateEntity(entity domain.ParsedEntity, index int) error {
	if entity == nil {
		return fmt.Errorf("entity at index %d is nil", index)
	}

	// Check if entity type is valid
	entityType := entity.EntityType()
	if entityType == "" {
		return fmt.Errorf("entity at index %d has empty entity type", index)
	}

	// Validate entity type against known types
	if !s.isValidEntityType(entityType) {
		return fmt.Errorf("entity at index %d has unknown entity type: %s", index, entityType)
	}

	// TODO: Add more specific validation based on entity type
	// This could include checking required fields, data formats, etc.
	
	return nil
}

// isValidEntityType checks if the entity type is among known valid types
func (s *ValidationStage) isValidEntityType(entityType string) bool {
	validTypes := []string{
		"incantesimo",
		"mostro", 
		"classe",
		"background",
		"arma",
		"armatura",
		"equipaggiamento",
		"oggetto_magico",
		"talento",
		"animale",
		"documento",
		"strumento",
		"servizio",
	}

	return slices.Contains(validTypes, entityType)
}