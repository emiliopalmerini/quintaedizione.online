package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/events"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/infrastructure"
)

// ProcessingStage represents a single stage in the processing pipeline
type ProcessingStage interface {
	Name() string
	Process(ctx context.Context, data *ProcessingData) error
}

// Pipeline orchestrates the execution of multiple processing stages
type Pipeline struct {
	stages   []ProcessingStage
	eventBus events.EventBus
	logger   parsers.Logger
	config   infrastructure.PipelineConfig
}

// ProcessingData holds data that flows through the pipeline stages
type ProcessingData struct {
	WorkItem    *parsers.WorkItem
	FilePath    string
	RawContent  []string
	ContentType parsers.ContentType
	ParsedData  []domain.ParsedEntity
	Documents   []map[string]any // for backward compatibility
	Metadata    map[string]any
	Errors      []error
}

// NewPipeline creates a new processing pipeline
func NewPipeline(stages []ProcessingStage, eventBus events.EventBus, logger parsers.Logger, config infrastructure.PipelineConfig) *Pipeline {
	if logger == nil {
		logger = &parsers.NoOpLogger{}
	}

	return &Pipeline{
		stages:   stages,
		eventBus: eventBus,
		logger:   logger,
		config:   config,
	}
}

// Execute runs all pipeline stages on the provided data
func (p *Pipeline) Execute(ctx context.Context, data *ProcessingData) error {
	if data == nil {
		return fmt.Errorf("processing data cannot be nil")
	}

	// Initialize metadata if nil
	if data.Metadata == nil {
		data.Metadata = make(map[string]any)
	}

	// Publish pipeline started event
	if p.eventBus != nil {
		p.eventBus.Publish(&events.PipelineStartedEvent{
			BaseEvent:  events.BaseEvent{EventTime: time.Now()},
			FilePath:   data.FilePath,
			Collection: data.WorkItem.Collection,
		})
	}

	// Execute stages sequentially
	for i, stage := range p.stages {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			p.logger.Error("pipeline execution cancelled at stage %s: %v", stage.Name(), err)
			if p.eventBus != nil {
				p.eventBus.Publish(&events.PipelineFailedEvent{
					BaseEvent:  events.BaseEvent{EventTime: time.Now()},
					FilePath:   data.FilePath,
					Collection: data.WorkItem.Collection,
					Error:      err,
					Stage:      stage.Name(),
				})
			}
			return err
		default:
		}

		p.logger.Debug("executing stage %d/%d: %s", i+1, len(p.stages), stage.Name())

		// Publish stage started event
		if p.eventBus != nil {
			p.eventBus.Publish(&events.StageStartedEvent{
				BaseEvent:  events.BaseEvent{EventTime: time.Now()},
				StageName:  stage.Name(),
				FilePath:   data.FilePath,
				Collection: data.WorkItem.Collection,
			})
		}

		if err := stage.Process(ctx, data); err != nil {
			data.Errors = append(data.Errors, err)
			p.logger.Error("stage %s failed: %v", stage.Name(), err)

			// Publish stage failed event
			if p.eventBus != nil {
				p.eventBus.Publish(&events.StageFailedEvent{
					BaseEvent:  events.BaseEvent{EventTime: time.Now()},
					StageName:  stage.Name(),
					FilePath:   data.FilePath,
					Collection: data.WorkItem.Collection,
					Error:      err,
				})
			}

			// Check if we should continue or stop on error
			if p.config.ShouldStopOnError() {
				if p.eventBus != nil {
					p.eventBus.Publish(&events.PipelineFailedEvent{
						BaseEvent:  events.BaseEvent{EventTime: time.Now()},
						FilePath:   data.FilePath,
						Collection: data.WorkItem.Collection,
						Error:      err,
						Stage:      stage.Name(),
					})
				}
				return fmt.Errorf("pipeline failed at stage %s: %w", stage.Name(), err)
			}
		} else {
			p.logger.Debug("stage %s completed successfully", stage.Name())

			// Publish stage completed event
			if p.eventBus != nil {
				p.eventBus.Publish(&events.StageCompletedEvent{
					BaseEvent:  events.BaseEvent{EventTime: time.Now()},
					StageName:  stage.Name(),
					FilePath:   data.FilePath,
					Collection: data.WorkItem.Collection,
				})
			}
		}
	}

	// Publish pipeline completed event
	if p.eventBus != nil {
		p.eventBus.Publish(&events.PipelineCompletedEvent{
			BaseEvent:   events.BaseEvent{EventTime: time.Now()},
			FilePath:    data.FilePath,
			Collection:  data.WorkItem.Collection,
			ParsedCount: len(data.ParsedData),
			HasErrors:   len(data.Errors) > 0,
		})
	}

	if len(data.Errors) > 0 {
		p.logger.Info("pipeline completed with %d errors for %s", len(data.Errors), data.FilePath)
	} else {
		p.logger.Info("pipeline completed successfully for %s", data.FilePath)
	}

	return nil
}

// AddStage adds a new stage to the pipeline
func (p *Pipeline) AddStage(stage ProcessingStage) {
	p.stages = append(p.stages, stage)
}

// GetStages returns all stages in the pipeline
func (p *Pipeline) GetStages() []ProcessingStage {
	return p.stages
}

// GetStageCount returns the number of stages in the pipeline
func (p *Pipeline) GetStageCount() int {
	return len(p.stages)
}

// GetConfig returns the pipeline configuration
func (p *Pipeline) GetConfig() infrastructure.PipelineConfig {
	return p.config
}

// NewProcessingData creates a new ProcessingData instance
func NewProcessingData(workItem *parsers.WorkItem, filePath string) *ProcessingData {
	return &ProcessingData{
		WorkItem:   workItem,
		FilePath:   filePath,
		RawContent: nil,
		ParsedData: nil,
		Documents:  nil,
		Metadata:   make(map[string]any),
		Errors:     nil,
	}
}
