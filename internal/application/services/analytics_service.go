package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type AnalyticsService struct {
	baseURL    string
	apiKey     string
	siteID     string
	httpClient *http.Client
}

type AnalyticsServiceConfig struct {
	BaseURL string
	APIKey  string
	SiteID  string
}

func NewAnalyticsService(config AnalyticsServiceConfig) *AnalyticsService {
	return &AnalyticsService{
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		siteID:  config.SiteID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type AggregateStats struct {
	Visitors      int64   `json:"visitors"`
	Pageviews     int64   `json:"pageviews"`
	BounceRate    float64 `json:"bounce_rate"`
	VisitDuration float64 `json:"visit_duration"`
}

type PageStats struct {
	Page      string `json:"page"`
	Visitors  int64  `json:"visitors"`
	Pageviews int64  `json:"pageviews"`
}

type ContentHeatmapItem struct {
	Collection string `json:"collection"`
	Slug       string `json:"slug"`
	Views      int64  `json:"views"`
	Visitors   int64  `json:"visitors"`
}

type SearchStat struct {
	Query    string `json:"query"`
	Count    int64  `json:"count"`
	Visitors int64  `json:"visitors"`
}

func (s *AnalyticsService) GetAggregateStats(ctx context.Context, period string) (*AggregateStats, error) {
	endpoint := fmt.Sprintf("%s/api/v1/stats/aggregate", s.baseURL)

	params := url.Values{}
	params.Set("site_id", s.siteID)
	params.Set("period", period)
	params.Set("metrics", "visitors,pageviews,bounce_rate,visit_duration")

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Results AggregateStats `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Results, nil
}

func (s *AnalyticsService) GetTopPages(ctx context.Context, period string, limit int) ([]PageStats, error) {
	endpoint := fmt.Sprintf("%s/api/v1/stats/breakdown", s.baseURL)

	params := url.Values{}
	params.Set("site_id", s.siteID)
	params.Set("period", period)
	params.Set("property", "event:page")
	params.Set("metrics", "visitors,pageviews")
	params.Set("limit", fmt.Sprintf("%d", limit))

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pages: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Results []PageStats `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Results, nil
}

func (s *AnalyticsService) GetContentHeatmap(ctx context.Context, period string, limit int) ([]ContentHeatmapItem, error) {
	pages, err := s.GetTopPages(ctx, period, limit*2)
	if err != nil {
		return nil, err
	}

	heatmap := make([]ContentHeatmapItem, 0, limit)

	for _, page := range pages {
		collection, slug := parsePagePath(page.Page)
		if collection == "" || slug == "" {
			continue
		}

		heatmap = append(heatmap, ContentHeatmapItem{
			Collection: collection,
			Slug:       slug,
			Views:      page.Pageviews,
			Visitors:   page.Visitors,
		})

		if len(heatmap) >= limit {
			break
		}
	}

	return heatmap, nil
}

func (s *AnalyticsService) GetSearchStats(ctx context.Context, period string, limit int) ([]SearchStat, error) {
	endpoint := fmt.Sprintf("%s/api/v1/stats/breakdown", s.baseURL)

	params := url.Values{}
	params.Set("site_id", s.siteID)
	params.Set("period", period)
	params.Set("property", "event:props:query")
	params.Set("metrics", "visitors,events")
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("filters", "event:name==Search")

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch search stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var rawResult struct {
		Results []map[string]interface{} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rawResult); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	stats := make([]SearchStat, 0, len(rawResult.Results))
	for _, r := range rawResult.Results {
		query, _ := r["query"].(string)
		events, _ := r["events"].(float64)
		visitors, _ := r["visitors"].(float64)

		if query != "" {
			stats = append(stats, SearchStat{
				Query:    query,
				Count:    int64(events),
				Visitors: int64(visitors),
			})
		}
	}

	return stats, nil
}

func parsePagePath(path string) (collection, slug string) {
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	parts := strings.Split(path, "/")
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}

	return "", ""
}
