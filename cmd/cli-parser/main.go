package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	inputDir    = flag.String("input", "data/ita/lists", "Input directory containing markdown files")
	outputFile  = flag.String("output", "", "Output file for results (optional)")
	filePattern = flag.String("pattern", "*.md", "File pattern to match")
	verbose     = flag.Bool("verbose", false, "Enable verbose logging")
	listFiles   = flag.Bool("list", false, "List available files and exit")
	parseFile   = flag.String("file", "", "Parse specific file")
	listParsers = flag.Bool("parsers", false, "List available parsers and exit")
	dryRun      = flag.Bool("dry-run", false, "Parse but don't save to database")
	mongoURI    = flag.String("mongo-uri", "", "MongoDB URI (default from env MONGO_URI)")
	dbName      = flag.String("db-name", "", "Database name (default from env DB_NAME)")
)

func main() {
	flag.Parse()

	fmt.Println("🧙‍♂️ D&D 5e SRD CLI Parser")
	fmt.Println("==========================")

	// Handle simple listing commands first
	if *listFiles {
		listMarkdownFiles(*inputDir)
		return
	}

	// Initialize parser with MongoDB connection
	mongoURIStr := *mongoURI
	if mongoURIStr == "" {
		mongoURIStr = getMongoURIFromEnv()
	}

	dbNameStr := *dbName
	if dbNameStr == "" {
		dbNameStr = getDBNameFromEnv()
	}

	parser, err := NewParserCLI(mongoURIStr, dbNameStr)
	if err != nil {
		log.Fatalf("❌ Failed to initialize parser: %v", err)
	}
	defer parser.Close()

	if *listParsers {
		parser.ListAvailableParsers()
		return
	}

	if *parseFile != "" {
		err := parser.ParseFile(*inputDir, *parseFile)
		if err != nil {
			log.Fatalf("❌ Failed to parse file %s: %v", *parseFile, err)
		}
		return
	}

	err = parser.ParseAllFiles(*inputDir)
	if err != nil {
		log.Fatalf("❌ Failed to parse files: %v", err)
	}
}

func listMarkdownFiles(inputDir string) {
	fmt.Printf("📁 Markdown files in %s:\n\n", inputDir)

	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			relPath, _ := filepath.Rel(inputDir, path)
			fmt.Printf("  • %s\n", relPath)
		}

		return nil
	})

	if err != nil {
		log.Printf("❌ Error listing files: %v", err)
	}
}
