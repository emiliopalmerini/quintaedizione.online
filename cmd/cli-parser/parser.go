package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/repositories"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/services"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
)

type ParserCLI struct {
	parserService *services.ParserService
	mongoClient   *mongodb.Client
	context       context.Context
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

	// Create Document parser registry
	documentRegistry, err := parsers.CreateDocumentRegistry()
	if err != nil {
		return nil, fmt.Errorf("failed to create document registry: %w", err)
	}

	// Create parser service
	parserService := services.NewParserService(services.ParserServiceConfig{
		DocumentRegistry: documentRegistry,
		DocumentRepo:     repositoryFactory.DocumentRepository(),
		WorkItems:        nil, // Use default work items
		Logger:           parsers.NewConsoleLogger("parser"),
		DryRun:           *dryRun, // Use the global flag
	})

	return &ParserCLI{
		parserService: parserService,
		mongoClient:   mongoClient,
		context:       ctx,
	}, nil
}

func (p *ParserCLI) ParseFile(inputDir, filename string) error {
	result, err := p.parserService.ParseFile(p.context, inputDir, filename)
	if err != nil {
		return err
	}

	if *verbose {
		fmt.Printf("📄 Parsed %d documents from %s\n", result.DocumentCount, result.Filename)
	}

	if !*dryRun {
		fmt.Printf("💾 Saved %d documents to collection '%s'\n", result.DocumentCount, result.Collection)
	} else {
		fmt.Printf("🔍 Dry run: would save %d documents to collection '%s'\n", result.DocumentCount, result.Collection)
	}

	return nil
}

func (p *ParserCLI) ParseAllFiles(inputDir string) error {
	result, err := p.parserService.ParseAllFiles(p.context, inputDir)
	if err != nil {
		return err
	}

	// Print detailed results if verbose
	if *verbose {
		for _, fileResult := range result.FileResults {
			if fileResult.Error != nil {
				fmt.Printf("❌ %s: %v\n", fileResult.Filename, fileResult.Error)
			} else {
				fmt.Printf("✅ %s: %d documents -> %s\n", fileResult.Filename, fileResult.DocumentCount, fileResult.Collection)
			}
		}
	}

	fmt.Printf("\n📊 Summary: %d successful, %d failed, %d total documents in %.2fs\n",
		result.SuccessCount, result.ErrorCount, result.TotalDocuments, result.Duration.Seconds())

	return nil
}

func (p *ParserCLI) ListAvailableParsers() {
	fmt.Println("📋 Available Document Parsers:")
	fmt.Println("==============================")

	workItems := p.parserService.GetWorkItems()
	fmt.Printf("\n✅ %d Document-based parsers configured\n", len(workItems))
	fmt.Println("\nAll parsers use the unified Document model with HTML rendering.")
	fmt.Println("\nConfigured collections:")
	for _, item := range workItems {
		fmt.Printf("  • %s -> %s\n", filepath.Base(item.Filename), item.Collection)
	}
}

func (p *ParserCLI) Close() error {
	// Close MongoDB connection
	if p.mongoClient != nil {
		return p.mongoClient.Close()
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
