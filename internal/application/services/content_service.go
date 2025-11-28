package services

import (
	"context"
	"fmt"
	"time"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/filters"
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/repositories"
	"github.com/emiliopalmerini/quintaedizione.online/internal/infrastructure"
)

type ContentService struct {
	documentRepo  repositories.DocumentRepository
	filterService filters.FilterService
	cache         *infrastructure.SimpleCache
}

func NewContentService(documentRepo repositories.DocumentRepository, filterService filters.FilterService) *ContentService {
	return &ContentService{
		documentRepo:  documentRepo,
		filterService: filterService,
		cache:         infrastructure.GetGlobalCache(),
	}
}

func (s *ContentService) GetCollectionItems(ctx context.Context, collection, search string, filterParams map[string]string, page, limit int) ([]map[string]any, int64, error) {

	skip := int64((page - 1) * limit)

	collectionType := filters.CollectionType(collection)

	searchFilter := s.filterService.BuildSearchFilter(collectionType, search)

	if len(filterParams) == 0 {
		items, totalCount, err := s.documentRepo.FindMaps(ctx, collection, searchFilter, skip, int64(limit))
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get collection items: %w", err)
		}
		return items, totalCount, nil
	}

	filterSet, err := s.filterService.ParseFilters(collectionType, filterParams)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse filters: %w", err)
	}

	fieldFilter, err := s.filterService.BuildMongoFilter(filterSet)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build field filter: %w", err)
	}

	combinedFilter := s.filterService.CombineFilters(fieldFilter, searchFilter)

	items, totalCount, err := s.documentRepo.FindMaps(ctx, collection, combinedFilter, skip, int64(limit))
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

	item, err := s.documentRepo.FindMapByID(ctx, collection, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to find item: %w", err)
	}

	s.cache.Set(cacheKey, item, 10*time.Minute)

	return item, nil
}

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

func (s *ContentService) GetCollectionStats(ctx context.Context) ([]map[string]any, error) {
	return s.documentRepo.GetAllCollectionStats(ctx)
}

func (s *ContentService) GetAdjacentItems(ctx context.Context, collection, currentSlug string) (prevSlug, nextSlug *string, err error) {
	return s.documentRepo.GetAdjacentMaps(ctx, collection, currentSlug)
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

	allCollections, err := s.documentRepo.GetAllCollectionStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get collections: %w", err)
	}

	results := make([]SearchResult, 0)

	for _, collectionInfo := range allCollections {
		collectionName, ok := collectionInfo["collection"].(string)
		if !ok {
			continue
		}

		collectionType := filters.CollectionType(collectionName)
		searchFilter := s.filterService.BuildSearchFilter(collectionType, query)

		items, total, err := s.documentRepo.FindMaps(ctx, collectionName, searchFilter, 0, int64(limitPerCollection))
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
