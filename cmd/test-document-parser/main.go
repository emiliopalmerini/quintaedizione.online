package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
)

var (
	inputFile = flag.String("file", "data/ita/lists/regole.md", "Input markdown file to parse")
	verbose   = flag.Bool("verbose", false, "Enable verbose output")
)

func main() {
	flag.Parse()

	fmt.Println("🧪 Document Parser Test")
	fmt.Println("=======================\n")

	// Create Document registry
	registry, err := parsers.CreateDocumentRegistry()
	if err != nil {
		log.Fatalf("❌ Failed to create registry: %v", err)
	}

	fmt.Printf("✅ Loaded %d Document parsers\n\n", registry.Count())

	// Read file
	content, err := os.ReadFile(*inputFile)
	if err != nil {
		log.Fatalf("❌ Failed to read file: %v", err)
	}

	// Determine content type from filename
	filename := *inputFile
	var contentType parsers.ContentType

	switch {
	case strings.Contains(filename, "regole"):
		contentType = parsers.ContentTypeRegole
	case strings.Contains(filename, "incantesimi"):
		contentType = parsers.ContentTypeIncantesimi
	case strings.Contains(filename, "mostri"):
		contentType = parsers.ContentTypeMostri
	case strings.Contains(filename, "animali"):
		contentType = parsers.ContentTypeAnimali
	case strings.Contains(filename, "classi"):
		contentType = parsers.ContentTypeClassi
	case strings.Contains(filename, "backgrounds"):
		contentType = parsers.ContentTypeBackgrounds
	case strings.Contains(filename, "armi"):
		contentType = parsers.ContentTypeArmi
	case strings.Contains(filename, "armature"):
		contentType = parsers.ContentTypeArmature
	case strings.Contains(filename, "equipaggiamenti"):
		contentType = parsers.ContentTypeEquipaggiamenti
	case strings.Contains(filename, "servizi"):
		contentType = parsers.ContentTypeServizi
	case strings.Contains(filename, "strumenti"):
		contentType = parsers.ContentTypeStrumenti
	case strings.Contains(filename, "talenti"):
		contentType = parsers.ContentTypeTalenti
	case strings.Contains(filename, "oggetti_magici"):
		contentType = parsers.ContentTypeOggettiMagici
	case strings.Contains(filename, "cavalcature"):
		contentType = parsers.ContentTypeCavalcatureVeicoli
	default:
		log.Fatalf("❌ Unknown content type for file: %s", filename)
	}

	// Get strategy
	strategy, err := registry.GetStrategy(contentType, parsers.Italian)
	if err != nil {
		log.Fatalf("❌ Failed to get strategy: %v", err)
	}

	fmt.Printf("📄 Parsing: %s\n", filename)
	fmt.Printf("🏷️  Content Type: %s\n", contentType)
	fmt.Printf("🔧 Strategy: %s\n\n", strategy.Name())

	// Parse content
	lines := strings.Split(string(content), "\n")
	ctx := parsers.NewParsingContext(filename, string(parsers.Italian))
	ctx.WithLogger(parsers.NewConsoleLogger("info"))

	documents, err := strategy.ParseDocument(lines, ctx)
	if err != nil {
		log.Fatalf("❌ Parsing failed: %v", err)
	}

	fmt.Printf("✅ Successfully parsed %d documents\n\n", len(documents))

	// Display results
	for i, doc := range documents {
		fmt.Printf("─────────────────────────────────────────────────\n")
		fmt.Printf("Document #%d\n", i+1)
		fmt.Printf("─────────────────────────────────────────────────\n")
		fmt.Printf("ID:      %s\n", doc.ID)
		fmt.Printf("Title:   %s\n", doc.Title)
		fmt.Printf("Collection: %s\n", doc.GetCollection())
		fmt.Printf("Filters: %v\n", doc.Filters)

		if *verbose {
			fmt.Printf("\nHTML Content:\n%s\n", doc.Content)
		} else {
			contentPreview := string(doc.Content)
			if len(contentPreview) > 200 {
				contentPreview = contentPreview[:200] + "..."
			}
			fmt.Printf("\nHTML Preview:\n%s\n", contentPreview)
		}
		fmt.Println()
	}

	fmt.Printf("─────────────────────────────────────────────────\n")
	fmt.Printf("✅ Test completed successfully!\n")
}
