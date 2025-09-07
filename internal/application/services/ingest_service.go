package services

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/infrastructure"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/infrastructure/mongodb"
)

// IngestService handles the ingestion business logic
type IngestService struct {
	repository domain.ParserRepository
}

// NewIngestService creates a new ingest service
func NewIngestService(repository domain.ParserRepository) *IngestService {
	return &IngestService{
		repository: repository,
	}
}

// ExecuteIngest runs the ingestion process
func (s *IngestService) ExecuteIngest(baseDir string, workItems []parsers.WorkItem, dryRun bool) ([]*parsers.IngestResult, error) {
	var results []*parsers.IngestResult

	for _, item := range workItems {
		result := s.processWorkItem(baseDir, item, dryRun)
		results = append(results, result)
	}

	return results, nil
}

// FilterWork filters work items by collection names
func (s *IngestService) FilterWork(items []parsers.WorkItem, only []string) []parsers.WorkItem {
	if len(only) == 0 {
		return items
	}

	wanted := make(map[string]bool)
	for _, collection := range only {
		wanted[collection] = true
	}

	var filtered []parsers.WorkItem
	for _, item := range items {
		if wanted[item.Collection] {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// processWorkItem processes a single work item
func (s *IngestService) processWorkItem(baseDir string, item parsers.WorkItem, dryRun bool) *parsers.IngestResult {
	result := parsers.NewIngestResult(item.Collection, item.Filename)

	// Build full path
	fullPath := filepath.Join(baseDir, item.Filename)

	// Check if file exists
	if !infrastructure.FileExists(fullPath) {
		result.SetError(fmt.Errorf("missing file: %s", fullPath))
		return result
	}

	// Read file
	lines, err := infrastructure.ReadLines(fullPath)
	if err != nil {
		result.SetError(fmt.Errorf("failed to read file %s: %w", fullPath, err))
		return result
	}

	// Parse content
	docs, err := item.Parser(lines)
	if err != nil {
		result.SetError(fmt.Errorf("failed to parse %s: %w", item.Filename, err))
		return result
	}

	result.Parsed = len(docs)

	if dryRun || s.repository == nil {
		// Generate preview for dry run
		preview := s.generatePreview(docs)
		result.SetPreview(preview)
		result.Written = 0
	} else {
		// Write to repository
		uniqueFields := mongodb.GetUniqueFieldsForCollection(item.Collection)
		written, err := s.repository.UpsertMany(item.Collection, uniqueFields, docs)
		if err != nil {
			result.SetError(fmt.Errorf("failed to upsert to %s: %w", item.Collection, err))
			return result
		}
		result.Written = written
	}

	return result
}

// generatePreview generates a preview of parsed documents
func (s *IngestService) generatePreview(docs []map[string]any) string {
	previewKeys := []string{"name", "term", "level", "rarity", "type", "school", "nome", "titolo"}

	var preview []map[string]any
	maxPreview := 5
	if len(docs) < maxPreview {
		maxPreview = len(docs)
	}

	for i := 0; i < maxPreview; i++ {
		doc := docs[i]
		previewDoc := make(map[string]any)

		for _, key := range previewKeys {
			if value, exists := doc[key]; exists {
				previewDoc[key] = value
			}
		}

		if len(previewDoc) > 0 {
			preview = append(preview, previewDoc)
		}
	}

	// Convert to JSON
	jsonBytes, err := json.MarshalIndent(preview, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error generating preview: %v", err)
	}

	return string(jsonBytes)
}

// GetCollectionStats returns statistics for collections
func (s *IngestService) GetCollectionStats() (map[string]int64, error) {
	if s.repository == nil {
		return nil, fmt.Errorf("repository not available")
	}

	collections := []string{
		"documenti", "classi", "backgrounds", "armi", "armature",
		"strumenti", "servizi", "equipaggiamento", "oggetti_magici",
		"incantesimi", "talenti", "mostri", "animali",
	}

	stats := make(map[string]int64)

	for _, collection := range collections {
		// Note: We'd need to add Count method to ParserRepository interface
		// For now, return 0 or implement differently
		stats[collection] = 0
	}

	return stats, nil
}
