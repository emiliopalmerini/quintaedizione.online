package models

type PageData struct {
	Title       string
	Description string
	Collection  string
	DocTitle    string
	DocID       string
	QueryString string
	TotalItems  int64
}

type Collection struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Count       int64  `json:"count"`
}

type HomePageData struct {
	PageData
	Collections []Collection
	Total       int64
}

type Document struct {
	ID              string                 `json:"_id"`
	Title           string                 `json:"title"`
	DisplayElements []DocumentDisplayField `json:"display_elements"`
	Translated      bool                   `json:"translated"`
}

type DocumentDisplayField struct {
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}

type FilterValueOption struct {
	Value string
	Count int64
}

type FilterOption struct {
	Name          string
	Label         string
	Values        []FilterValueOption
	CurrentValue  string
	CurrentValues []string
}

type CollectionPageData struct {
	PageData
	Documents  []Document
	Filters    []FilterOption
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

type ItemPageData struct {
	PageData
	BodyRaw         string
	BodyHTML        string
	PrevID          string
	NextID          string
	Position        int
	Total           int
	CollectionLabel string
	SourceShort     string // edition badge (e.g. "5e", "5.5e"); empty if single source
}

type ErrorPageData struct {
	PageData
	ErrorTitle   string
	ErrorMessage string
	ErrorCode    int
}

type CollectionSearchResult struct {
	CollectionName  string
	CollectionLabel string
	Documents       []Document
	Total           int64
	HasMore         bool
}

type SearchPageData struct {
	PageData
	Query   string
	Results []CollectionSearchResult
	Total   int64
}
