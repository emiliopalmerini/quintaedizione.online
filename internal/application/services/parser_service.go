package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emiliopalmerini/quintaedizione.online/internal/application/parsers"
	domainRepos "github.com/emiliopalmerini/quintaedizione.online/internal/domain/repositories"
)

type ParserService struct {
	documentRegistry *parsers.DocumentRegistry
	documentRepo     domainRepos.DocumentRepository
	workItems        []parsers.WorkItem
	logger           parsers.Logger
	dryRun           bool
}

type ParserServiceConfig struct {
	DocumentRegistry *parsers.DocumentRegistry
	DocumentRepo     domainRepos.DocumentRepository
	WorkItems        []parsers.WorkItem
	Logger           parsers.Logger
	DryRun           bool
}

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

type ParseResult struct {
	TotalFiles     int
	SuccessCount   int
	ErrorCount     int
	TotalDocuments int
	Duration       time.Duration
	FileResults    []FileResult
}

type FileResult struct {
	Filename      string
	Collection    string
	DocumentCount int
	Error         error
}

func (s *ParserService) ParseAllFiles(ctx context.Context, inputDir string) (*ParseResult, error) {
	startTime := time.Now()

	s.logger.Info(fmt.Sprintf("Starting parsing of %d files from %s", len(s.workItems), inputDir))

	result := &ParseResult{
		TotalFiles:  len(s.workItems),
		FileResults: make([]FileResult, 0, len(s.workItems)),
	}

	for _, workItem := range s.workItems {
		filename := filepath.Base(workItem.Filename)
		s.logger.Info(fmt.Sprintf("Processing: %s -> %s", filename, workItem.Collection))

		fileResult := s.parseFile(ctx, inputDir, workItem)
		result.FileResults = append(result.FileResults, fileResult)

		if fileResult.Error != nil {
			s.logger.Error(fmt.Sprintf("Failed to parse %s: %v", filename, fileResult.Error))
			result.ErrorCount++

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

func (s *ParserService) ParseFile(ctx context.Context, inputDir, filename string) (*FileResult, error) {

	workItem, err := s.findWorkItem(filename)
	if err != nil {
		return nil, fmt.Errorf("work item not found for %s: %w", filename, err)
	}

	result := s.parseFile(ctx, inputDir, workItem)
	return &result, nil
}

func (s *ParserService) parseFile(ctx context.Context, inputDir string, workItem parsers.WorkItem) FileResult {
	result := FileResult{
		Filename:   filepath.Base(workItem.Filename),
		Collection: workItem.Collection,
	}

	filePath := filepath.Join(inputDir, filepath.Base(workItem.Filename))

	contentType, err := parsers.GetContentTypeFromCollection(workItem.Collection)
	if err != nil {
		result.Error = fmt.Errorf("invalid content type for collection %s: %w", workItem.Collection, err)
		return result
	}

	strategy, err := s.documentRegistry.GetStrategy(contentType, workItem.Language)
	if err != nil {
		result.Error = fmt.Errorf("document parser not found for %s: %w", contentType, err)
		return result
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		result.Error = fmt.Errorf("failed to read file %s: %w", filePath, err)
		return result
	}

	lines := strings.Split(string(content), "\n")

	parsingContext := parsers.NewParsingContext(result.Filename, string(workItem.Language))
	parsingContext.WithLogger(s.logger)

	documents, err := strategy.ParseDocument(lines, parsingContext)
	if err != nil {
		result.Error = fmt.Errorf("parsing failed for %s: %w", result.Filename, err)
		return result
	}

	result.DocumentCount = len(documents)

	if !s.dryRun {
		saved, err := s.documentRepo.UpsertMany(ctx, workItem.Collection, documents)
		if err != nil {
			result.Error = fmt.Errorf("failed to save documents: %w", err)
			return result
		}

		result.DocumentCount = saved
	}

	return result
}

func (s *ParserService) findWorkItem(filename string) (parsers.WorkItem, error) {

	baseName := strings.TrimSuffix(filename, filepath.Ext(filename))

	for _, item := range s.workItems {
		itemBaseName := strings.TrimSuffix(filepath.Base(item.Filename), filepath.Ext(item.Filename))
		if itemBaseName == baseName {
			return item, nil
		}
	}

	return parsers.WorkItem{}, fmt.Errorf("no work item found for file: %s", filename)
}

func (s *ParserService) GetWorkItems() []parsers.WorkItem {
	return s.workItems
}

func (s *ParserService) SetDryRun(dryRun bool) {
	s.dryRun = dryRun
}
