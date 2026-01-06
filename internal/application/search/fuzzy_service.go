package search

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/lithammer/fuzzysearch/fuzzy"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/collections"
	domainsearch "github.com/emiliopalmerini/quintaedizione.online/internal/domain/search"
)

type FuzzySearchService struct {
	repo  domainsearch.SearchRepository
	index *SearchIndex
}

func NewFuzzySearchService(repo domainsearch.SearchRepository) *FuzzySearchService {
	return &FuzzySearchService{
		repo:  repo,
		index: NewSearchIndex(10 * time.Minute),
	}
}

func (svc *FuzzySearchService) Search(ctx context.Context, query string, limitPerCollection int) ([]domainsearch.SearchResultSet, error) {
	if query == "" {
		return nil, nil
	}

	if svc.index.NeedsRefresh() {
		if err := svc.RefreshIndex(ctx); err != nil {
			return nil, err
		}
	}

	query = strings.ToLower(strings.TrimSpace(query))
	allItems := svc.index.GetAll()

	var results []domainsearch.SearchResultSet

	for collection, items := range allItems {
		if len(items) == 0 {
			continue
		}

		collectionResults, totalMatches := svc.searchInCollection(items, query, limitPerCollection)
		if totalMatches == 0 {
			continue
		}

		results = append(results, domainsearch.SearchResultSet{
			Collection: collection,
			Results:    collectionResults,
			Total:      int64(totalMatches),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if len(results[i].Results) == 0 || len(results[j].Results) == 0 {
			return len(results[i].Results) > len(results[j].Results)
		}
		return results[i].Results[0].Score > results[j].Results[0].Score
	})

	return results, nil
}

func (svc *FuzzySearchService) SearchCollection(ctx context.Context, collection, query string, limit int) ([]domainsearch.SearchResult, error) {
	if query == "" {
		return nil, nil
	}

	items := svc.index.Get(collection)
	if len(items) == 0 {
		var err error
		items, err = svc.repo.GetSearchableItems(ctx, collection)
		if err != nil {
			return nil, err
		}
		svc.index.Set(collection, items)
	}

	query = strings.ToLower(strings.TrimSpace(query))
	results, _ := svc.searchInCollection(items, query, limit)
	return results, nil
}

func (svc *FuzzySearchService) searchInCollection(items []domainsearch.SearchableItem, query string, limit int) ([]domainsearch.SearchResult, int) {
	type rankedItem struct {
		item  domainsearch.SearchableItem
		score int
	}

	var ranked []rankedItem

	for _, item := range items {
		searchText := strings.ToLower(item.Title)
		if len(item.Keywords) > 0 {
			searchText += " " + strings.ToLower(strings.Join(item.Keywords, " "))
		}

		if !fuzzy.Match(query, searchText) {
			continue
		}

		score := fuzzy.RankMatch(query, searchText)
		if score == -1 {
			continue
		}

		titleScore := fuzzy.RankMatch(query, strings.ToLower(item.Title))
		if titleScore != -1 {
			score = titleScore * 10
		}

		ranked = append(ranked, rankedItem{item: item, score: score})
	}

	totalMatches := len(ranked)

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].score > ranked[j].score
	})

	if len(ranked) > limit {
		ranked = ranked[:limit]
	}

	results := make([]domainsearch.SearchResult, 0, len(ranked))
	for _, r := range ranked {
		results = append(results, domainsearch.SearchResult{
			ID:         r.item.ID,
			Collection: r.item.Collection,
			Title:      r.item.Title,
			Score:      r.score,
		})
	}

	return results, totalMatches
}

func (svc *FuzzySearchService) RefreshIndex(ctx context.Context) error {
	allCollections := collections.GetAllCollections()

	for _, col := range allCollections {
		items, err := svc.repo.GetSearchableItems(ctx, col.String())
		if err != nil {
			return err
		}
		svc.index.Set(col.String(), items)
	}

	return nil
}
