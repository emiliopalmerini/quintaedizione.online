package commands

import (
	"context"
	"fmt"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/services"
)

// SearchCommand encapsulates search operations with filters
type SearchCommand struct {
	collection  string
	query       string
	filters     map[string]string
	page        int
	pageSize    int
	service     *services.ContentService
}

// NewSearchCommand creates a new search command
func NewSearchCommand(collection, query string, filters map[string]string, page, pageSize int, service *services.ContentService) *SearchCommand {
	return &SearchCommand{
		collection: collection,
		query:      query,
		filters:    filters,
		page:       page,
		pageSize:   pageSize,
		service:    service,
	}
}

// Validate ensures the command is valid
func (c *SearchCommand) Validate() error {
	if c.collection == "" {
		return fmt.Errorf("collection cannot be empty")
	}

	if c.page < 1 {
		return fmt.Errorf("page must be >= 1")
	}

	if c.pageSize < 1 || c.pageSize > 100 {
		return fmt.Errorf("page size must be between 1 and 100")
	}

	return nil
}

// Execute runs the search command
func (c *SearchCommand) Execute(ctx context.Context) ([]map[string]interface{}, int64, error) {
	if len(c.filters) > 0 {
		return c.service.GetCollectionItemsWithFilters(ctx, c.collection, c.query, c.filters, c.page, c.pageSize)
	}

	return c.service.GetCollectionItems(ctx, c.collection, c.query, c.page, c.pageSize)
}

// BulkSearchCommand handles searching across multiple collections
type BulkSearchCommand struct {
	collections []string
	query       string
	limit       int
	service     *services.ContentService
}

// NewBulkSearchCommand creates a new bulk search command
func NewBulkSearchCommand(collections []string, query string, limit int, service *services.ContentService) *BulkSearchCommand {
	return &BulkSearchCommand{
		collections: collections,
		query:       query,
		limit:       limit,
		service:     service,
	}
}

// Validate ensures the command is valid
func (c *BulkSearchCommand) Validate() error {
	if len(c.collections) == 0 {
		return fmt.Errorf("at least one collection must be specified")
	}

	if c.query == "" {
		return fmt.Errorf("search query cannot be empty")
	}

	if c.limit < 1 || c.limit > 50 {
		return fmt.Errorf("limit must be between 1 and 50")
	}

	return nil
}

// Execute runs the bulk search across collections
func (c *BulkSearchCommand) Execute(ctx context.Context) (map[string][]map[string]interface{}, error) {
	results := make(map[string][]map[string]interface{})

	for _, collection := range c.collections {
		items, _, err := c.service.GetCollectionItems(ctx, collection, c.query, 1, c.limit)
		if err != nil {
			return nil, fmt.Errorf("search failed for collection %s: %w", collection, err)
		}

		results[collection] = items
	}

	return results, nil
}