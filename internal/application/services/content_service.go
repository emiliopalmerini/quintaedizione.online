package services

import (
	"context"
	"fmt"
	"time"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/collections"
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/filters"
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/repositories"
	"github.com/emiliopalmerini/quintaedizione.online/internal/infrastructure"
)

type ContentService struct {
	documentReader repositories.DocumentReader
	documentStats  repositories.DocumentStatistics
	documentNav    repositories.DocumentNavigation
	filterService  filters.FilterService
	cache          *infrastructure.SimpleCache
}

func NewContentService(repo repositories.DocumentRepository, filterService filters.FilterService, cache *infrastructure.SimpleCache) *ContentService {
	return &ContentService{
		documentReader: repo,
		documentStats:  repo,
		documentNav:    repo,
		filterService:  filterService,
		cache:          cache,
	}
}

func (s *ContentService) GetCollectionItems(ctx context.Context, collection, search string, filterParams map[string]string, page, limit int) ([]map[string]any, int64, error) {
	skip := int64((page - 1) * limit)

	collectionType, _ := collections.FromString(collection)

	searchPred := s.filterService.BuildSearchPredicate(collectionType, search)

	if len(filterParams) == 0 {
		items, totalCount, err := s.documentReader.FindMaps(ctx, collection, searchPred, skip, int64(limit))
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get collection items: %w", err)
		}
		return items, totalCount, nil
	}

	filterSet, err := s.filterService.ParseFilters(collectionType, filterParams)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse filters: %w", err)
	}

	fieldPred, err := s.filterService.BuildFilter(filterSet)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build field filter: %w", err)
	}

	combined := s.filterService.CombinePredicates(fieldPred, searchPred)

	items, totalCount, err := s.documentReader.FindMaps(ctx, collection, combined, skip, int64(limit))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get collection items with filters: %w", err)
	}

	return items, totalCount, nil
}

func (s *ContentService) GetItem(ctx context.Context, collection, slug string) (map[string]any, error) {
	cacheKey := fmt.Sprintf("item:%s:%s", collection, slug)
	if cached, found := s.cache.Get(cacheKey); found {
		if item, ok := cached.(map[string]any); ok {
			return item, nil
		}
	}

	item, err := s.documentReader.FindMapByID(ctx, collection, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to find item: %w", err)
	}

	s.cache.Set(cacheKey, item, 10*time.Minute)

	return item, nil
}

func (s *ContentService) GetStats(ctx context.Context) (map[string]any, error) {
	allStats, err := s.documentStats.GetAllCollectionStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection stats: %w", err)
	}

	stats := map[string]any{
		"collections": make(map[string]int64),
		"total_items": int64(0),
	}

	for _, collection := range allStats {
		if name, ok := collection["collection"].(string); ok {
			if count, ok := collection["count"].(int64); ok {
				stats["collections"].(map[string]int64)[name] = count
				stats["total_items"] = stats["total_items"].(int64) + count
			}
		}
	}

	return stats, nil
}

func (s *ContentService) GetCollectionStats(ctx context.Context) ([]map[string]any, error) {
	return s.documentStats.GetAllCollectionStats(ctx)
}

func (s *ContentService) GetAdjacentItems(ctx context.Context, collection, currentSlug string) (prevSlug, nextSlug *string, err error) {
	return s.documentNav.GetAdjacentMaps(ctx, collection, currentSlug)
}

func (s *ContentService) GetAvailableFilters(collection string) ([]filters.FilterDefinition, error) {
	collectionType, ok := collections.FromString(collection)
	if !ok {
		return nil, nil
	}
	return s.filterService.GetAvailableFilters(collectionType)
}

type SearchResult struct {
	Collection string           `json:"collection"`
	Items      []map[string]any `json:"items"`
	Total      int64            `json:"total"`
}

func (s *ContentService) GlobalSearch(ctx context.Context, query string, limitPerCollection int) ([]SearchResult, error) {
	if query == "" {
		return []SearchResult{}, nil
	}

	allCollections, err := s.documentStats.GetAllCollectionStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get collections: %w", err)
	}

	results := make([]SearchResult, 0)

	for _, collectionInfo := range allCollections {
		collectionName, ok := collectionInfo["collection"].(string)
		if !ok {
			continue
		}

		collectionType, _ := collections.FromString(collectionName)
		searchPred := s.filterService.BuildSearchPredicate(collectionType, query)

		items, total, err := s.documentReader.FindMaps(ctx, collectionName, searchPred, 0, int64(limitPerCollection))
		if err != nil {
			fmt.Printf("Warning: Failed to search in collection %s: %v\n", collectionName, err)
			continue
		}

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
