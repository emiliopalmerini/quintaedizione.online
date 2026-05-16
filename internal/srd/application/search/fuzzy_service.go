package search

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/lithammer/fuzzysearch/fuzzy"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/collections"
	domainsearch "github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/search"
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
	tokens := tokenize(normalize(query))
	if len(tokens) == 0 {
		return nil, nil
	}

	if svc.index.NeedsRefresh() {
		if err := svc.RefreshIndex(ctx); err != nil {
			return nil, err
		}
	}

	allItems := svc.index.GetAll()

	var results []domainsearch.SearchResultSet

	for collection, items := range allItems {
		if len(items) == 0 {
			continue
		}

		collectionResults, totalMatches := svc.searchInCollection(items, tokens, limitPerCollection)
		if totalMatches == 0 {
			continue
		}

		results = append(results, domainsearch.SearchResultSet{
			Collection: collection,
			Results:    collectionResults,
			Total:      int64(totalMatches),
		})
	}

	// Rank collections by aggregate score of their top results, with a small
	// boost for collections that returned more matches. Avoids letting one
	// outlier hit jump a low-quality collection above a richer one.
	sort.Slice(results, func(i, j int) bool {
		si := aggregateScore(results[i].Results)
		sj := aggregateScore(results[j].Results)
		if si != sj {
			return si > sj
		}
		return results[i].Total > results[j].Total
	})

	return results, nil
}

func (svc *FuzzySearchService) SearchCollection(ctx context.Context, collection, query string, limit int) ([]domainsearch.SearchResult, error) {
	tokens := tokenize(normalize(query))
	if len(tokens) == 0 {
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

	results, _ := svc.searchInCollection(items, tokens, limit)
	return results, nil
}

func (svc *FuzzySearchService) searchInCollection(items []domainsearch.SearchableItem, tokens []string, limit int) ([]domainsearch.SearchResult, int) {
	type rankedItem struct {
		item  domainsearch.SearchableItem
		score int
	}

	ranked := make([]rankedItem, 0, len(items))

	for _, item := range items {
		title := normalize(item.Title)
		desc := normalize(item.Description)
		titleWords := tokenize(title)

		score, ok := scoreItem(tokens, title, titleWords, desc)
		if !ok {
			continue
		}
		ranked = append(ranked, rankedItem{item: item, score: score})
	}

	totalMatches := len(ranked)

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].score > ranked[j].score
	})

	if limit > 0 && len(ranked) > limit {
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

// scoreItem returns the item's total score and whether every query token
// found a match. All tokens must match (AND semantics) for the item to
// qualify; this prevents "palla fuoco" from returning every spell with
// either "palla" or "fuoco".
func scoreItem(tokens []string, title string, titleWords []string, desc string) (int, bool) {
	total := 0
	for _, t := range tokens {
		s := scoreToken(t, title, titleWords, desc)
		if s == 0 {
			return 0, false
		}
		total += s
	}
	// Whole-query exact title match: small bonus to lift the canonical hit.
	joined := strings.Join(tokens, " ")
	if strings.Contains(title, joined) {
		total += 200
	}
	return total, true
}

// scoreToken returns the best score for a single token against an item.
// 0 means no match.
func scoreToken(token, title string, titleWords []string, desc string) int {
	// Title-word prefix match (strongest signal: "fire" → "fireball").
	for _, w := range titleWords {
		if strings.HasPrefix(w, token) {
			return 1000 + len(token)*10
		}
	}
	// Title substring (e.g. token is in the middle of a compound word).
	if strings.Contains(title, token) {
		return 800 + len(token)*10
	}
	// Fuzzy on title words. Distance budget grows with token length; min 1.
	budget := fuzzyBudget(len(token))
	if budget > 0 {
		best := -1
		for _, w := range titleWords {
			d := fuzzy.LevenshteinDistance(token, w)
			if d <= budget && (best == -1 || d < best) {
				best = d
			}
		}
		if best >= 0 {
			return 500 - best*50
		}
	}
	// Description substring (weaker than any title hit).
	if desc != "" && strings.Contains(desc, token) {
		return 100 + len(token)*2
	}
	return 0
}

// fuzzyBudget returns the max Levenshtein distance accepted for a token of
// the given length. Two-char tokens get one typo; longer tokens roughly
// 30%. Returns 0 for tokens too short to fuzzy-match safely.
func fuzzyBudget(tokenLen int) int {
	if tokenLen < 2 {
		return 0
	}
	if tokenLen <= 4 {
		return 1
	}
	return tokenLen / 3
}

// aggregateScore sums the top-N result scores for a collection. Used to
// rank collections relative to each other.
func aggregateScore(results []domainsearch.SearchResult) int {
	total := 0
	for _, r := range results {
		total += r.Score
	}
	return total
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
