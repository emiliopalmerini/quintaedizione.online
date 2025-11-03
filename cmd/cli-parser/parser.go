package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/repositories"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	domainRepos "github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain/repositories"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
)

type ParserCLI struct {
	documentRegistry  *parsers.DocumentRegistry
	repositoryFactory *repositories.RepositoryFactory
	documentRepo      domainRepos.DocumentRepository
	context           context.Context
	workItems         []parsers.WorkItem
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
	documentRepo := repositoryFactory.DocumentRepository()

	// Create Document parser registry
	documentRegistry, err := parsers.CreateDocumentRegistry()
	if err != nil {
		return nil, fmt.Errorf("failed to create document registry: %w", err)
	}

	// Load default work items
	workItems := parsers.CreateDefaultWorkItems()

	return &ParserCLI{
		documentRegistry:  documentRegistry,
		repositoryFactory: repositoryFactory,
		documentRepo:      documentRepo,
		context:           ctx,
		workItems:         workItems,
	}, nil
}

func (p *ParserCLI) ParseFile(inputDir, filename string) error {
	return p.parseFileWithDocuments(inputDir, filename)
}

func (p *ParserCLI) parseFileWithDocuments(inputDir, filename string) error {
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

	// Get Document parsing strategy
	strategy, err := p.documentRegistry.GetStrategy(contentType, parsers.Italian)
	if err != nil {
		return fmt.Errorf("document parser not found for %s: %w", contentType, err)
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
	documents, err := strategy.ParseDocument(lines, parsingContext)
	if err != nil {
		return fmt.Errorf("parsing failed for %s: %w", filename, err)
	}

	if *verbose {
		fmt.Printf("📄 Parsed %d documents from %s\n", len(documents), filename)
		for i, doc := range documents {
			fmt.Printf("  [%d] %s (ID: %s)\n", i+1, doc.Title, doc.ID)
		}
	}

	// Save documents to MongoDB if not in dry-run mode
	if !*dryRun {
		err = p.saveDocuments(documents, workItem.Collection)
		if err != nil {
			return fmt.Errorf("failed to save documents: %w", err)
		}
		fmt.Printf("💾 Saved %d documents to collection '%s'\n", len(documents), workItem.Collection)
	} else {
		fmt.Printf("🔍 Dry run: would save %d documents to collection '%s'\n", len(documents), workItem.Collection)
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
	fmt.Println("📋 Available Document Parsers:")
	fmt.Println("==============================")
	fmt.Printf("\n✅ %d Document-based parsers loaded\n", p.documentRegistry.Count())
	fmt.Println("\nAll parsers use the unified Document model with HTML rendering.")
	fmt.Println("Supported content types: regole, incantesimi, mostri, animali, classi,")
	fmt.Println("backgrounds, armi, armature, equipaggiamenti, servizi, strumenti,")
	fmt.Println("talenti, oggetti_magici, cavalcature_veicoli")
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

func (p *ParserCLI) saveDocuments(documents []*domain.Document, collection string) error {
	if len(documents) == 0 {
		return nil
	}

	if *verbose {
		fmt.Printf("🗃️  Saving to collection: %s\n", collection)
	}

	// Use DocumentRepository to save documents
	saved, err := p.documentRepo.UpsertMany(p.context, collection, documents)
	if err != nil {
		return fmt.Errorf("failed to save documents to collection %s: %w", collection, err)
	}

	if *verbose {
		fmt.Printf("  Saved/updated %d documents in MongoDB\n", saved)
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
