package stages

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/events"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/pipeline"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/infrastructure"
)

// FileReaderStage reads files from the filesystem using existing file utilities
type FileReaderStage struct {
	baseDir  string
	eventBus events.EventBus
	logger   parsers.Logger
}

// NewFileReaderStage creates a new file reader stage
func NewFileReaderStage(baseDir string, eventBus events.EventBus, logger parsers.Logger) *FileReaderStage {
	if logger == nil {
		logger = &parsers.NoOpLogger{}
	}

	return &FileReaderStage{
		baseDir:  baseDir,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Name returns the stage name
func (s *FileReaderStage) Name() string {
	return "file_reader"
}

// Process reads the file content and populates RawContent in ProcessingData
func (s *FileReaderStage) Process(ctx context.Context, data *pipeline.ProcessingData) error {
	if data == nil {
		return fmt.Errorf("processing data cannot be nil")
	}

	if data.WorkItem == nil {
		return fmt.Errorf("work item cannot be nil")
	}

	// Build full file path
	fullPath := filepath.Join(s.baseDir, data.WorkItem.Filename)
	data.FilePath = fullPath

	s.logger.Debug("reading file: %s", fullPath)

	// Check if file exists using existing infrastructure
	if !infrastructure.FileExists(fullPath) {
		err := fmt.Errorf("file not found: %s", fullPath)
		s.logger.Error("file not found: %s", fullPath)
		
		// Publish file processing error event
		if s.eventBus != nil {
			s.eventBus.Publish(&events.ParsingErrorEvent{
				BaseEvent:  events.BaseEvent{EventTime: time.Now()},
				FilePath:   fullPath,
				Collection: data.WorkItem.Collection,
				Error:      err,
			})
		}
		
		return err
	}

	// Read file lines using existing infrastructure
	lines, err := infrastructure.ReadLines(fullPath)
	if err != nil {
		s.logger.Error("failed to read file %s: %v", fullPath, err)
		
		// Publish file processing error event
		if s.eventBus != nil {
			s.eventBus.Publish(&events.ParsingErrorEvent{
				BaseEvent:  events.BaseEvent{EventTime: time.Now()},
				FilePath:   fullPath,
				Collection: data.WorkItem.Collection,
				Error:      err,
			})
		}
		
		return fmt.Errorf("failed to read file %s: %w", fullPath, err)
	}

	// Store raw content in processing data
	data.RawContent = lines
	
	// Store metadata
	data.Metadata["file_size"] = len(lines)
	data.Metadata["file_path"] = fullPath
	
	s.logger.Debug("successfully read %d lines from file: %s", len(lines), fullPath)

	// Publish file processing started event
	if s.eventBus != nil {
		s.eventBus.Publish(&events.FileProcessingStartedEvent{
			BaseEvent:  events.BaseEvent{EventTime: time.Now()},
			FilePath:   fullPath,
			Collection: data.WorkItem.Collection,
		})
	}

	return nil
}