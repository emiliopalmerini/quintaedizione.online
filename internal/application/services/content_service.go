package services

import (
	"context"
	"fmt"
	"time"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain/filters"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain/repositories"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/infrastructure"
)

// ContentService provides business logic for content operations
type ContentService struct {
	documentRepo  repositories.DocumentRepository
	filterService filters.FilterService
	cache         *infrastructure.SimpleCache
}

// NewContentService creates a new ContentService instance
func NewContentService(documentRepo repositories.DocumentRepository, filterService filters.FilterService) *ContentService {
	return &ContentService{
		documentRepo:  documentRepo,
		filterService: filterService,
		cache:         infrastructure.GetGlobalCache(),
	}
}

// GetCollectionItems retrieves items from a collection with pagination, search, and optional field filters
// Pass nil or empty filterParams map to skip field filtering
func (s *ContentService) GetCollectionItems(ctx context.Context, collection, search string, filterParams map[string]string, page, limit int) ([]map[string]any, int64, error) {
	// Calculate skip
	skip := int64((page - 1) * limit)

	// Get collection type for filter building
	collectionType := filters.CollectionType(collection)

	// Build search filter
	searchFilter := s.filterService.BuildSearchFilter(collectionType, search)

	// If no field filters, use search filter only
	if len(filterParams) == 0 {
		items, totalCount, err := s.documentRepo.FindMaps(ctx, collection, searchFilter, skip, int64(limit))
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get collection items: %w", err)
		}
		return items, totalCount, nil
	}

	// Parse and apply field filters
	filterSet, err := s.filterService.ParseFilters(collectionType, filterParams)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse filters: %w", err)
	}

	// Build field filter
	fieldFilter, err := s.filterService.BuildMongoFilter(filterSet)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build field filter: %w", err)
	}

	// Combine filters
	combinedFilter := s.filterService.CombineFilters(fieldFilter, searchFilter)

	// Get items using the document repository
	items, totalCount, err := s.documentRepo.FindMaps(ctx, collection, combinedFilter, skip, int64(limit))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get collection items with filters: %w", err)
	}

	return items, totalCount, nil
}

// GetItem retrieves a specific item by slug
func (s *ContentService) GetItem(ctx context.Context, collection, slug string) (map[string]any, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("item:%s:%s", collection, slug)
	if cached, found := s.cache.Get(cacheKey); found {
		if item, ok := cached.(map[string]any); ok {
			return item, nil
		}
	}

	item, err := s.documentRepo.FindMapByID(ctx, collection, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to find item: %w", err)
	}

	// Cache the item for 10 minutes
	s.cache.Set(cacheKey, item, 10*time.Minute)

	return item, nil
}

// GetStats retrieves database statistics
func (s *ContentService) GetStats(ctx context.Context) (map[string]any, error) {
	collections, err := s.documentRepo.GetAllCollectionStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection stats: %w", err)
	}

	stats := map[string]any{
		"collections": make(map[string]int64),
		"total_items": int64(0),
	}

	for _, collection := range collections {
		if name, ok := collection["collection"].(string); ok {
			if count, ok := collection["count"].(int64); ok {
				stats["collections"].(map[string]int64)[name] = count
				stats["total_items"] = stats["total_items"].(int64) + count
			}
		}
	}

	return stats, nil
}

// GetCollectionStats retrieves statistics for all collections
func (s *ContentService) GetCollectionStats(ctx context.Context) ([]map[string]any, error) {
	return s.documentRepo.GetAllCollectionStats(ctx)
}

// GetAdjacentItems gets the previous and next items for navigation
func (s *ContentService) GetAdjacentItems(ctx context.Context, collection, currentSlug string) (prevSlug, nextSlug *string, err error) {
	return s.documentRepo.GetAdjacentMaps(ctx, collection, currentSlug)
}

// SearchResult represents a search result from a single collection
type SearchResult struct {
	Collection string         `json:"collection"`
	Items      []map[string]any `json:"items"`
	Total      int64          `json:"total"`
}

// GlobalSearch searches across all collections and returns results grouped by collection
func (s *ContentService) GlobalSearch(ctx context.Context, query string, limitPerCollection int) ([]SearchResult, error) {
	if query == "" {
		return []SearchResult{}, nil
	}

	// Get all collection names from the cache or repository
	allCollections, err := s.documentRepo.GetAllCollectionStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get collections: %w", err)
	}

	results := make([]SearchResult, 0)

	// Search in each collection
	for _, collectionInfo := range allCollections {
		collectionName, ok := collectionInfo["collection"].(string)
		if !ok {
			continue
		}

		// Build search filter for this collection
		collectionType := filters.CollectionType(collectionName)
		searchFilter := s.filterService.BuildSearchFilter(collectionType, query)

		// Search in this collection with limit
		items, total, err := s.documentRepo.FindMaps(ctx, collectionName, searchFilter, 0, int64(limitPerCollection))
		if err != nil {
			// Log error but continue with other collections
			fmt.Printf("Warning: Failed to search in collection %s: %v\n", collectionName, err)
			continue
		}

		// Only include collections with results
		if total > 0 {
			results = append(results, SearchResult{
				Collection: collectionName,
				Items:      items,
				Total:      total,
			})
		}
	}

	return results, nil
}
