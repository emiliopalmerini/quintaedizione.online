package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/repositories"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
)

type ParserCLI struct {
	registry         *parsers.Registry
	repositoryFactory *repositories.RepositoryFactory
	repositoryWrapper *repositories.ParserRepositoryWrapper
	context          context.Context
	workItems        []parsers.WorkItem
}

func NewParserCLI(mongoURI, dbName string) (*ParserCLI, error) {
	ctx := context.Background()

	// Initialize MongoDB connection
	mongoClient, err := mongodb.NewClient(mongodb.Config{
		URI:         mongoURI,
		Database:    dbName,
		Timeout:     30 * time.Second,
		MaxPoolSize: 10,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Initialize repository factory
	repositoryFactory := repositories.NewRepositoryFactory(mongoClient)
	repositoryWrapper := repositories.NewParserRepositoryWrapper(repositoryFactory)

	// Create parser registry with all strategies
	registry, err := parsers.CreateDefaultRegistry()
	if err != nil {
		return nil, fmt.Errorf("failed to create parser registry: %w", err)
	}

	// Load default work items
	workItems := parsers.CreateDefaultWorkItems()

	return &ParserCLI{
		registry:          registry,
		repositoryFactory: repositoryFactory,
		repositoryWrapper: repositoryWrapper,
		context:          ctx,
		workItems:        workItems,
	}, nil
}

func (p *ParserCLI) ParseFile(inputDir, filename string) error {
	filePath := filepath.Join(inputDir, filename)

	// Find matching work item
	workItem, err := p.findWorkItem(filename)
	if err != nil {
		return fmt.Errorf("work item not found for %s: %w", filename, err)
	}

	// Get content type from collection
	contentType, err := parsers.GetContentTypeFromCollection(workItem.Collection)
	if err != nil {
		return fmt.Errorf("invalid content type for collection %s: %w", workItem.Collection, err)
	}

	// Get parsing strategy
	strategy, err := p.registry.GetStrategy(contentType, parsers.Italian)
	if err != nil {
		return fmt.Errorf("parser not found for %s: %w", contentType, err)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Split content into lines
	lines := strings.Split(string(content), "\n")

	// Create parsing context
	parsingContext := parsers.NewParsingContext(filename, string(parsers.Italian))
	parsingContext.WithLogger(parsers.NewConsoleLogger("info"))

	// Parse content
	entities, err := strategy.Parse(lines, parsingContext)
	if err != nil {
		return fmt.Errorf("parsing failed for %s: %w", filename, err)
	}

	if *verbose {
		fmt.Printf("📄 Parsed %d entities from %s\n", len(entities), filename)
		for i, entity := range entities {
			fmt.Printf("  [%d] %s (type: %s)\n", i+1, "Entity", entity.EntityType())
		}
	}

	// Save entities to MongoDB if not in dry-run mode
	if !*dryRun {
		err = p.saveEntities(entities, workItem.Collection)
		if err != nil {
			return fmt.Errorf("failed to save entities: %w", err)
		}
		fmt.Printf("💾 Saved %d entities to collection '%s'\n", len(entities), workItem.Collection)
	} else {
		fmt.Printf("🔍 Dry run: would save %d entities to collection '%s'\n", len(entities), workItem.Collection)
	}

	return nil
}

func (p *ParserCLI) ParseAllFiles(inputDir string) error {
	successCount := 0
	errorCount := 0

	for _, workItem := range p.workItems {
		filename := filepath.Base(workItem.Filename)
		fmt.Printf("🔄 Processing: %s -> %s\n", filename, workItem.Collection)

		err := p.ParseFile(inputDir, filename)
		if err != nil {
			fmt.Printf("❌ Error parsing %s: %v\n", filename, err)
			errorCount++
			continue
		}

		successCount++
	}

	fmt.Printf("\n📊 Summary: %d successful, %d failed\n", successCount, errorCount)
	return nil
}

func (p *ParserCLI) ListAvailableParsers() {
	parserInfos := parsers.GetAvailableParsers(p.registry)
	
	fmt.Println("📋 Available Parsers:")
	fmt.Println("===================")
	
	for _, parser := range parserInfos {
		fmt.Printf("• %s (%s)\n", parser.Name, parser.Key)
		if parser.Description != "" {
			fmt.Printf("  %s\n", parser.Description)
		}
		fmt.Println()
	}
}

func (p *ParserCLI) findWorkItem(filename string) (parsers.WorkItem, error) {
	// Remove extension for comparison
	baseName := strings.TrimSuffix(filename, filepath.Ext(filename))
	
	for _, item := range p.workItems {
		itemBaseName := strings.TrimSuffix(filepath.Base(item.Filename), filepath.Ext(item.Filename))
		if itemBaseName == baseName {
			return item, nil
		}
	}
	
	return parsers.WorkItem{}, fmt.Errorf("no work item found for file: %s", filename)
}

func (p *ParserCLI) saveEntities(entities []domain.ParsedEntity, collection string) error {
	if len(entities) == 0 {
		return nil
	}
	
	if *verbose {
		fmt.Printf("🗃️  Saving to collection: %s\n", collection)
	}
	
	// Convert entities to flattened maps without wrapper
	docs := make([]map[string]any, len(entities))
	for i, entity := range entities {
		// Convert entity to map via JSON marshaling/unmarshaling
		jsonData, err := json.Marshal(entity)
		if err != nil {
			return fmt.Errorf("failed to marshal entity %d: %w", i, err)
		}
		
		var entityMap map[string]any
		if err := json.Unmarshal(jsonData, &entityMap); err != nil {
			return fmt.Errorf("failed to unmarshal entity %d: %w", i, err)
		}
		
		// Add metadata fields directly to the flattened document
		entityMap["collection"] = collection
		entityMap["source_file"] = fmt.Sprintf("ita/lists/%s.md", collection)
		entityMap["language"] = "ita"
		entityMap["created_at"] = time.Now()
		
		docs[i] = entityMap
	}
	
	// Use repository wrapper to save with upsert semantics
	uniqueFields := []string{"slug"} // Use slug as unique field
	saved, err := p.repositoryWrapper.UpsertMany(collection, uniqueFields, docs)
	if err != nil {
		return fmt.Errorf("failed to save entities to collection %s: %w", collection, err)
	}
	
	if *verbose {
		fmt.Printf("  Saved/updated %d entities in MongoDB\n", saved)
	}
	
	return nil
}

func (p *ParserCLI) Close() error {
	// Close MongoDB connection
	if p.repositoryFactory != nil {
		// Repository factory doesn't have a close method, but we should close the underlying client
		// This would need to be implemented in the factory or client
	}
	return nil
}

func getMongoURIFromEnv() string {
	if uri := os.Getenv("MONGO_URI"); uri != "" {
		return uri
	}
	return "mongodb://admin:password@localhost:27017/?authSource=admin"
}

func getDBNameFromEnv() string {
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		return dbName
	}
	return "dnd"
}