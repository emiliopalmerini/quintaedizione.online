package models

// PageData represents common page data for all templates
type PageData struct {
	Title       string
	Description string
	Collection  string
	DocTitle    string
	DocID       string
	QueryString string
}

// Collection represents a content collection
type Collection struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	Count int64  `json:"count"`
}

// HomePageData represents data for the home page
type HomePageData struct {
	PageData
	Collections []Collection
	Total       int64
}

// Document represents a document in a collection
type Document struct {
	ID              string                 `json:"_id"`   // Slug/identifier from Document model
	Title           string                 `json:"title"` // Display name from Document model
	DisplayElements []DocumentDisplayField `json:"display_elements"`
	Translated      bool                   `json:"translated"`
}

// DocumentDisplayField represents a field to display in document lists
type DocumentDisplayField struct {
	Value string `json:"value"`
}

// CollectionPageData represents data for collection list pages
type CollectionPageData struct {
	PageData
	Documents  []Document
	Query      string
	Page       int
	PageSize   int
	Total      int64
	TotalPages int
	HasNext    bool
	HasPrev    bool
	StartItem  int
	EndItem    int
}

// ItemPageData represents data for individual item pages
type ItemPageData struct {
	PageData
	BodyRaw         string
	BodyHTML        string
	PrevID          string
	NextID          string
	CollectionLabel string
}

// ErrorPageData represents data for error pages
type ErrorPageData struct {
	PageData
	ErrorTitle   string
	ErrorMessage string
	ErrorCode    int
}

// CollectionSearchResult represents search results from a single collection
type CollectionSearchResult struct {
	CollectionName  string
	CollectionLabel string
	Documents       []Document
	Total           int64
	HasMore         bool
}

// SearchPageData represents data for global search results page
type SearchPageData struct {
	PageData
	Query   string
	Results []CollectionSearchResult
	Total   int64
}
