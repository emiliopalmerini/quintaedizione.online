package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
	domainRepos "github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain/repositories"
)

// ParserService handles parsing markdown files into documents
type ParserService struct {
	documentRegistry *parsers.DocumentRegistry
	documentRepo     domainRepos.DocumentRepository
	workItems        []parsers.WorkItem
	logger           parsers.Logger
	dryRun           bool
}

// ParserServiceConfig contains configuration for ParserService
type ParserServiceConfig struct {
	DocumentRegistry *parsers.DocumentRegistry
	DocumentRepo     domainRepos.DocumentRepository
	WorkItems        []parsers.WorkItem
	Logger           parsers.Logger
	DryRun           bool
}

// NewParserService creates a new ParserService
func NewParserService(config ParserServiceConfig) *ParserService {
	logger := config.Logger
	if logger == nil {
		logger = parsers.NewConsoleLogger("info")
	}

	workItems := config.WorkItems
	if workItems == nil {
		workItems = parsers.CreateDefaultWorkItems()
	}

	return &ParserService{
		documentRegistry: config.DocumentRegistry,
		documentRepo:     config.DocumentRepo,
		workItems:        workItems,
		logger:           logger,
		dryRun:           config.DryRun,
	}
}

// ParseResult contains the result of a parsing operation
type ParseResult struct {
	TotalFiles      int
	SuccessCount    int
	ErrorCount      int
	TotalDocuments  int
	Duration        time.Duration
	FileResults     []FileResult
}

// FileResult contains the result of parsing a single file
type FileResult struct {
	Filename       string
	Collection     string
	DocumentCount  int
	Error          error
}

// ParseAllFiles parses all markdown files in the input directory
func (s *ParserService) ParseAllFiles(ctx context.Context, inputDir string) (*ParseResult, error) {
	startTime := time.Now()

	s.logger.Info(fmt.Sprintf("Starting parsing of %d files from %s", len(s.workItems), inputDir))

	result := &ParseResult{
		TotalFiles:   len(s.workItems),
		FileResults:  make([]FileResult, 0, len(s.workItems)),
	}

	for _, workItem := range s.workItems {
		filename := filepath.Base(workItem.Filename)
		s.logger.Info(fmt.Sprintf("Processing: %s -> %s", filename, workItem.Collection))

		fileResult := s.parseFile(ctx, inputDir, workItem)
		result.FileResults = append(result.FileResults, fileResult)

		if fileResult.Error != nil {
			s.logger.Error(fmt.Sprintf("Failed to parse %s: %v", filename, fileResult.Error))
			result.ErrorCount++
			// Return error immediately to block startup
			return result, fmt.Errorf("parsing failed for %s: %w", filename, fileResult.Error)
		}

		result.SuccessCount++
		result.TotalDocuments += fileResult.DocumentCount
		s.logger.Info(fmt.Sprintf("Saved %d documents to collection '%s'", fileResult.DocumentCount, workItem.Collection))
	}

	result.Duration = time.Since(startTime)
	s.logger.Info(fmt.Sprintf("Parsing completed: %d files, %d documents in %.2fs",
		result.SuccessCount, result.TotalDocuments, result.Duration.Seconds()))

	return result, nil
}

// ParseFile parses a specific markdown file
func (s *ParserService) ParseFile(ctx context.Context, inputDir, filename string) (*FileResult, error) {
	// Find matching work item
	workItem, err := s.findWorkItem(filename)
	if err != nil {
		return nil, fmt.Errorf("work item not found for %s: %w", filename, err)
	}

	result := s.parseFile(ctx, inputDir, workItem)
	return &result, nil
}

// parseFile is the internal implementation that parses a file
func (s *ParserService) parseFile(ctx context.Context, inputDir string, workItem parsers.WorkItem) FileResult {
	result := FileResult{
		Filename:   filepath.Base(workItem.Filename),
		Collection: workItem.Collection,
	}

	// Build full file path
	filePath := filepath.Join(inputDir, filepath.Base(workItem.Filename))

	// Get content type from collection
	contentType, err := parsers.GetContentTypeFromCollection(workItem.Collection)
	if err != nil {
		result.Error = fmt.Errorf("invalid content type for collection %s: %w", workItem.Collection, err)
		return result
	}

	// Get Document parsing strategy
	strategy, err := s.documentRegistry.GetStrategy(contentType, workItem.Language)
	if err != nil {
		result.Error = fmt.Errorf("document parser not found for %s: %w", contentType, err)
		return result
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		result.Error = fmt.Errorf("failed to read file %s: %w", filePath, err)
		return result
	}

	// Split content into lines
	lines := strings.Split(string(content), "\n")

	// Create parsing context
	parsingContext := parsers.NewParsingContext(result.Filename, string(workItem.Language))
	parsingContext.WithLogger(s.logger)

	// Parse content
	documents, err := strategy.ParseDocument(lines, parsingContext)
	if err != nil {
		result.Error = fmt.Errorf("parsing failed for %s: %w", result.Filename, err)
		return result
	}

	result.DocumentCount = len(documents)

	// Save documents to MongoDB if not in dry-run mode
	if !s.dryRun {
		saved, err := s.documentRepo.UpsertMany(ctx, workItem.Collection, documents)
		if err != nil {
			result.Error = fmt.Errorf("failed to save documents: %w", err)
			return result
		}

		// Update document count with actual saved count
		result.DocumentCount = saved
	}

	return result
}

// findWorkItem finds a work item by filename
func (s *ParserService) findWorkItem(filename string) (parsers.WorkItem, error) {
	// Remove extension for comparison
	baseName := strings.TrimSuffix(filename, filepath.Ext(filename))

	for _, item := range s.workItems {
		itemBaseName := strings.TrimSuffix(filepath.Base(item.Filename), filepath.Ext(item.Filename))
		if itemBaseName == baseName {
			return item, nil
		}
	}

	return parsers.WorkItem{}, fmt.Errorf("no work item found for file: %s", filename)
}

// GetWorkItems returns the configured work items
func (s *ParserService) GetWorkItems() []parsers.WorkItem {
	return s.workItems
}

// SetDryRun sets the dry-run mode
func (s *ParserService) SetDryRun(dryRun bool) {
	s.dryRun = dryRun
}
