package services

import (
	"context"
	"fmt"
	"time"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/collections"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/filters"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/repositories"
)

type ContentService struct {
	documentReader repositories.DocumentReader
	documentStats  repositories.DocumentStatistics
	documentNav    repositories.DocumentNavigation
	filterService  filters.FilterService
	cache          domain.Cache
}

func NewContentService(repo repositories.DocumentRepository, filterService filters.FilterService, cache domain.Cache) *ContentService {
	return &ContentService{
		documentReader: repo,
		documentStats:  repo,
		documentNav:    repo,
		filterService:  filterService,
		cache:          cache,
	}
}

func (s *ContentService) GetCollectionItems(ctx context.Context, collection, search string, filterParams map[string]string, page, limit int) ([]*domain.Document, int64, error) {
	skip := int64((page - 1) * limit)

	collectionType, _ := collections.FromString(collection)

	searchPred := s.filterService.BuildSearchPredicate(collectionType, search)

	if len(filterParams) == 0 {
		return s.documentReader.FindByPredicate(ctx, collection, searchPred, skip, int64(limit))
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

	return s.documentReader.FindByPredicate(ctx, collection, combined, skip, int64(limit))
}

func (s *ContentService) GetItem(ctx context.Context, collection, slug string) (*domain.Document, error) {
	cacheKey := fmt.Sprintf("item:%s:%s", collection, slug)
	if cached, found := s.cache.Get(cacheKey); found {
		if item, ok := cached.(*domain.Document); ok {
			return item, nil
		}
	}

	item, err := s.documentReader.FindByID(ctx, collection, slug)
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

func (s *ContentService) GetAdjacentItems(ctx context.Context, collection, currentSlug string) (prevSlug, nextSlug *string, position, total int, err error) {
	return s.documentNav.GetAdjacent(ctx, collection, currentSlug)
}

// GetFacetCounts returns, for each filter definition, a map of value → count.
// For each filter, counts are computed with all OTHER active filters + search applied,
// so counts reflect what would happen if you selected that value.
func (s *ContentService) GetFacetCounts(ctx context.Context, collection, search string, filterParams map[string]string) (map[string]map[string]int64, error) {
	collectionType, ok := collections.FromString(collection)
	if !ok {
		return nil, nil
	}

	defs, err := s.filterService.GetAvailableFilters(collectionType)
	if err != nil {
		return nil, err
	}

	searchPred := s.filterService.BuildSearchPredicate(collectionType, search)
	result := make(map[string]map[string]int64, len(defs))

	for _, def := range defs {
		if len(def.EnumValues) == 0 {
			continue
		}

		// Build predicate from all OTHER active filters (exclude current filter)
		otherParams := make(map[string]string)
		for k, v := range filterParams {
			if k != def.Name {
				otherParams[k] = v
			}
		}

		var combinedPred filters.DocumentPredicate
		if len(otherParams) > 0 {
			filterSet, err := s.filterService.ParseFilters(collectionType, otherParams)
			if err != nil {
				continue
			}
			fieldPred, err := s.filterService.BuildFilter(filterSet)
			if err != nil {
				continue
			}
			combinedPred = s.filterService.CombinePredicates(fieldPred, searchPred)
		} else {
			combinedPred = searchPred
		}

		counts, err := s.documentStats.AggregateField(ctx, collection, def.FieldPath, combinedPred)
		if err != nil {
			continue
		}
		result[def.Name] = counts
	}

	return result, nil
}

func (s *ContentService) GetAvailableFilters(collection string) ([]filters.FilterDefinition, error) {
	collectionType, ok := collections.FromString(collection)
	if !ok {
		return nil, nil
	}
	return s.filterService.GetAvailableFilters(collectionType)
}

type SearchResult struct {
	Collection string             `json:"collection"`
	Items      []*domain.Document `json:"items"`
	Total      int64              `json:"total"`
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

		items, total, err := s.documentReader.FindByPredicate(ctx, collectionName, searchPred, 0, int64(limitPerCollection))
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
