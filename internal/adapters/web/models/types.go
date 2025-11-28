package models

type PageData struct {
	Title       string
	Description string
	Collection  string
	DocTitle    string
	DocID       string
	QueryString string
}

type Collection struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	Count int64  `json:"count"`
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
}

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

type ItemPageData struct {
	PageData
	BodyRaw         string
	BodyHTML        string
	PrevID          string
	NextID          string
	CollectionLabel string
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
