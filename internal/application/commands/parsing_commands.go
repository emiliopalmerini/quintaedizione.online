package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/parsers"
)

// ParseFileCommand handles parsing of individual files
type ParseFileCommand struct {
	filePath      string
	contentType   string
	parserRegistry *parsers.ParserRegistry
}

// NewParseFileCommand creates a new command to parse a file
func NewParseFileCommand(filePath, contentType string, parserRegistry *parsers.ParserRegistry) *ParseFileCommand {
	return &ParseFileCommand{
		filePath:      filePath,
		contentType:   contentType,
		parserRegistry: parserRegistry,
	}
}

// Validate ensures the command is valid
func (c *ParseFileCommand) Validate() error {
	if c.filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	if c.contentType == "" {
		return fmt.Errorf("content type cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(c.filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", c.filePath)
	}

	return nil
}

// Execute runs the command
func (c *ParseFileCommand) Execute(ctx context.Context) error {
	parser, err := c.parserRegistry.GetParser(c.contentType)
	if err != nil {
		return fmt.Errorf("failed to get parser for type %s: %w", c.contentType, err)
	}

	content, err := os.ReadFile(c.filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", c.filePath, err)
	}

	_, err = parser.Parse(string(content), c.filePath)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", c.filePath, err)
	}

	return nil
}

// ParseDirectoryCommand handles parsing all files in a directory
type ParseDirectoryCommand struct {
	directoryPath  string
	parserRegistry *parsers.ParserRegistry
}

// NewParseDirectoryCommand creates a new command to parse all files in a directory
func NewParseDirectoryCommand(directoryPath string, parserRegistry *parsers.ParserRegistry) *ParseDirectoryCommand {
	return &ParseDirectoryCommand{
		directoryPath:  directoryPath,
		parserRegistry: parserRegistry,
	}
}

// Validate ensures the command is valid
func (c *ParseDirectoryCommand) Validate() error {
	if c.directoryPath == "" {
		return fmt.Errorf("directory path cannot be empty")
	}

	// Check if directory exists
	if stat, err := os.Stat(c.directoryPath); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", c.directoryPath)
	} else if !stat.IsDir() {
		return fmt.Errorf("path is not a directory: %s", c.directoryPath)
	}

	return nil
}

// Execute runs the command
func (c *ParseDirectoryCommand) Execute(ctx context.Context) error {
	// Walk through directory and find .md files
	return filepath.Walk(c.directoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if info.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		// Determine content type from filename
		baseName := filepath.Base(path)
		contentType := c.getContentTypeFromFilename(baseName)
		if contentType == "" {
			// Skip files that don't match known content types
			return nil
		}

		// Create and execute parse file command
		parseCmd := NewParseFileCommand(path, contentType, c.parserRegistry)
		if err := parseCmd.Validate(); err != nil {
			return fmt.Errorf("invalid parse command for %s: %w", path, err)
		}

		return parseCmd.Execute(ctx)
	})
}

// getContentTypeFromFilename determines content type based on filename
func (c *ParseDirectoryCommand) getContentTypeFromFilename(filename string) string {
	// Map common filenames to content types
	fileMapping := map[string]string{
		"incantesimi.md":        "incantesimo",
		"mostri.md":             "mostro",
		"classi.md":             "classe",
		"backgrounds.md":        "background",
		"equipaggiamenti.md":    "equipaggiamento",
		"armi.md":               "arma",
		"armature.md":           "armatura",
		"oggetti_magici.md":     "oggetto_magico",
		"talenti.md":            "talento",
		"servizi.md":            "servizio",
		"strumenti.md":          "strumento",
		"animali.md":            "animale",
		"regole.md":             "regola",
		"cavalcature_veicoli.md": "cavalcatura_veicolo",
	}

	return fileMapping[filename]
}